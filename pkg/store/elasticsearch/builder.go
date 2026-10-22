// Copyright 2023 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/weijinwen/jupiter/pkg/xlog"
)

// Filter 表示单个过滤条件
type Filter struct {
	Field    string
	Value    any
	Operator string
}

// FilterCondition 表示过滤条件组合
type FilterCondition struct {
	Filters  []Filter
	Operator string // "AND" 或 "OR"
}

// FilterTag 定义结构体字段的过滤行为
type FilterTag struct {
	Field      string
	Operator   string
	Required   bool
	Group      string   // 用于分组过滤条件
	Transform  string   // 值转换函数
	Validate   string   // 验证函数
	IgnoreZero bool     // 是否忽略零值
	Nested     bool     // 是否嵌套结构体
	DateFormat string   // 日期格式
	Deps       []string // 依赖字段
}

// GeoPoint 表示地理位置点
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GeoFilter 表示地理位置过滤
type GeoFilter struct {
	Field  string
	Point  GeoPoint
	Radius float64 // 半径，单位：米
	Unit   string  // 单位，默认为 "m"
}

// DateRangeFilter 表示日期范围过滤
type DateRangeFilter struct {
	Field string
	Start time.Time
	End   time.Time
}

// ArrayFilter 表示数组过滤
type ArrayFilter struct {
	Field    string
	Values   []any
	Operator string // "IN", "NOT IN"
}

// Query 表示一个已构建完成的查询体，可直接交给 Client/Index 的 Search 方法
type Query map[string]any

// Reader 把查询体序列化为请求体
func (q Query) Reader() (io.Reader, error) {
	buf, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}
	return bytes.NewReader(buf), nil
}

// Builder 用于构建 Elasticsearch 查询 DSL
type Builder struct {
	must      []map[string]any
	filter    []map[string]any
	should    []map[string]any
	mustNot   []map[string]any
	sort      []map[string]any
	source    []string
	highlight map[string]any
	aggs      map[string]any
	script    []map[string]any
	from      *int
	size      *int
	minScore  *float64
	trackHits any
	timeout   string

	filterGroups map[string][]FilterCondition
	validators   map[string]func(any) error
	transformers map[string]func(any) any
}

// NewBuilder 创建新的 Builder 实例，默认返回前 20 条
func NewBuilder() *Builder {
	size := 20
	return &Builder{
		size:         &size,
		highlight:    make(map[string]any),
		aggs:         make(map[string]any),
		filterGroups: make(map[string][]FilterCondition),
		validators:   make(map[string]func(any) error),
		transformers: make(map[string]func(any) any),
	}
}

// SetQuery 设置全文检索条件。
// fields 为空时使用 query_string 做全字段查询，否则使用 multi_match。
func (b *Builder) SetQuery(query string, fields ...string) *Builder {
	if len(fields) == 0 {
		b.must = append(b.must, map[string]any{
			"query_string": map[string]any{"query": query},
		})
		return b
	}

	b.must = append(b.must, map[string]any{
		"multi_match": map[string]any{
			"query":  query,
			"fields": fields,
		},
	})
	return b
}

// MatchAll 显式声明匹配全部文档
func (b *Builder) MatchAll() *Builder {
	b.must = append(b.must, map[string]any{"match_all": map[string]any{}})
	return b
}

// SetFrom 设置偏移量，用于分页
func (b *Builder) SetFrom(from int) *Builder {
	b.from = &from
	return b
}

// SetSize 设置返回条数
func (b *Builder) SetSize(size int) *Builder {
	b.size = &size
	return b
}

// SetMinScore 设置最低评分
func (b *Builder) SetMinScore(score float64) *Builder {
	b.minScore = &score
	return b
}

// SetTimeout 设置查询超时，例如 "1s"
func (b *Builder) SetTimeout(timeout string) *Builder {
	b.timeout = timeout
	return b
}

// TrackTotalHits 设置 total 命中数的统计方式，传 true 表示精确统计
func (b *Builder) TrackTotalHits(v any) *Builder {
	b.trackHits = v
	return b
}

// AddFilter 添加单个过滤条件，多个条件之间为 AND 关系。
// 支持的 operator：=、!=、>、>=、<、<=、IN、NOT IN、EXISTS、NOT EXISTS、CONTAINS、NOT CONTAINS。
func (b *Builder) AddFilter(field string, operator string, value any) *Builder {
	clause, err := formatFilter(Filter{Field: field, Operator: operator, Value: value})
	if err != nil {
		xlog.Jupiter().Warn("invalid filter", xlog.FieldErr(err), xlog.String("field", field), xlog.String("operator", operator))
		return b
	}

	if clause == nil {
		return b
	}

	switch clause.kind {
	case clauseMustNot:
		b.mustNot = append(b.mustNot, clause.body)
	default:
		b.filter = append(b.filter, clause.body)
	}
	return b
}

// AddFilterCondition 添加组合过滤条件，Operator 为 "OR" 时整组进入 should
func (b *Builder) AddFilterCondition(condition FilterCondition) *Builder {
	if len(condition.Filters) == 0 {
		return b
	}

	group := make([]map[string]any, 0, len(condition.Filters))
	for _, f := range condition.Filters {
		clause, err := formatFilter(f)
		if err != nil {
			xlog.Jupiter().Warn("invalid filter", xlog.FieldErr(err), xlog.String("field", f.Field), xlog.String("operator", f.Operator))
			continue
		}
		if clause == nil {
			continue
		}
		group = append(group, clause.body)
	}

	if len(group) == 0 {
		return b
	}

	// 单个条件直接平铺，避免无意义的 bool 嵌套
	if len(group) == 1 {
		b.filter = append(b.filter, group[0])
		return b
	}

	key := "must"
	if strings.EqualFold(condition.Operator, "OR") {
		key = "should"
	}
	b.filter = append(b.filter, map[string]any{
		"bool": map[string]any{key: group},
	})
	return b
}

// AddFilterGroup 添加分组过滤条件，后续用 ApplyFilterGroup 应用
func (b *Builder) AddFilterGroup(groupName string, condition FilterCondition) *Builder {
	if b.filterGroups == nil {
		b.filterGroups = make(map[string][]FilterCondition)
	}
	b.filterGroups[groupName] = append(b.filterGroups[groupName], condition)
	return b
}

// ApplyFilterGroup 应用分组过滤条件
func (b *Builder) ApplyFilterGroup(groupName string) *Builder {
	conditions, exists := b.filterGroups[groupName]
	if !exists {
		return b
	}

	for _, condition := range conditions {
		b.AddFilterCondition(condition)
	}
	return b
}

// AddGeoFilter 添加地理位置过滤，映射为 geo_distance
func (b *Builder) AddGeoFilter(geoFilter GeoFilter) *Builder {
	unit := geoFilter.Unit
	if unit == "" {
		unit = "m"
	}

	b.filter = append(b.filter, map[string]any{
		"geo_distance": map[string]any{
			"distance": fmt.Sprintf("%v%s", geoFilter.Radius, unit),
			geoFilter.Field: map[string]any{
				"lat": geoFilter.Point.Lat,
				"lon": geoFilter.Point.Lng,
			},
		},
	})
	return b
}

// AddDateRangeFilter 添加日期范围过滤
func (b *Builder) AddDateRangeFilter(dateFilter DateRangeFilter) *Builder {
	rng := make(map[string]any, 2)
	if !dateFilter.Start.IsZero() {
		rng["gte"] = dateFilter.Start.Format(time.RFC3339)
	}
	if !dateFilter.End.IsZero() {
		rng["lte"] = dateFilter.End.Format(time.RFC3339)
	}
	if len(rng) == 0 {
		return b
	}

	b.filter = append(b.filter, map[string]any{
		"range": map[string]any{dateFilter.Field: rng},
	})
	return b
}

// AddArrayFilter 添加数组过滤，IN 映射为 terms，NOT IN 映射为 must_not terms
func (b *Builder) AddArrayFilter(arrayFilter ArrayFilter) *Builder {
	operator := arrayFilter.Operator
	if operator == "" {
		operator = "IN"
	}
	return b.AddFilter(arrayFilter.Field, operator, arrayFilter.Values)
}

// AddFilterStruct 根据结构体的 filter tag 添加过滤条件
func (b *Builder) AddFilterStruct(filterStruct any) *Builder {
	filters, err := structToFilters(filterStruct, b)
	if err != nil {
		xlog.Jupiter().Warn("struct to filters failed", xlog.FieldErr(err))
		return b
	}

	for _, filter := range filters {
		b.AddFilter(filter.Field, filter.Operator, filter.Value)
	}
	return b
}

// AddFilterStructWithGroup 根据结构体添加过滤条件并归入指定分组
func (b *Builder) AddFilterStructWithGroup(groupName string, filterStruct any) *Builder {
	filters, err := structToFilters(filterStruct, b)
	if err != nil {
		xlog.Jupiter().Warn("struct to filters failed", xlog.FieldErr(err))
		return b
	}

	if len(filters) == 0 {
		return b
	}

	return b.AddFilterGroup(groupName, FilterCondition{Filters: filters, Operator: "AND"})
}

// RegisterValidator 注册自定义验证器
func (b *Builder) RegisterValidator(field string, validator func(any) error) *Builder {
	b.validators[field] = validator
	return b
}

// RegisterTransformer 注册自定义转换器
func (b *Builder) RegisterTransformer(field string, transformer func(any) any) *Builder {
	b.transformers[field] = transformer
	return b
}

// AddSort 添加排序字段，ascending 为 false 时降序
func (b *Builder) AddSort(field string, ascending bool) *Builder {
	order := "asc"
	if !ascending {
		order = "desc"
	}
	b.sort = append(b.sort, map[string]any{field: map[string]any{"order": order}})
	return b
}

// AddSortWithMode 添加排序字段并指定排序模式（如 min/max/avg，适用于数组字段）
func (b *Builder) AddSortWithMode(field string, ascending bool, mode string) *Builder {
	order := "asc"
	if !ascending {
		order = "desc"
	}
	b.sort = append(b.sort, map[string]any{field: map[string]any{"order": order, "mode": mode}})
	return b
}

// SetSource 设置返回字段，会覆盖已有配置
func (b *Builder) SetSource(fields ...string) *Builder {
	b.source = fields
	return b
}

// AddSource 追加返回字段
func (b *Builder) AddSource(fields ...string) *Builder {
	b.source = append(b.source, fields...)
	return b
}

// AddHighlight 添加高亮字段
func (b *Builder) AddHighlight(fields ...string) *Builder {
	existing, _ := b.highlight["fields"].(map[string]any)
	if existing == nil {
		existing = make(map[string]any)
	}
	for _, field := range fields {
		existing[field] = map[string]any{}
	}
	b.highlight["fields"] = existing
	return b
}

// HighlightPreTag 设置高亮标签前缀
func (b *Builder) HighlightPreTag(preTag string) *Builder {
	b.highlight["pre_tags"] = []string{preTag}
	return b
}

// HighlightPostTag 设置高亮标签后缀
func (b *Builder) HighlightPostTag(postTag string) *Builder {
	b.highlight["post_tags"] = []string{postTag}
	return b
}

// AddScriptField 添加脚本字段
func (b *Builder) AddScriptField(name string, script map[string]any) *Builder {
	b.script = append(b.script, map[string]any{name: map[string]any{"script": script}})
	return b
}

// AddTermsAgg 添加 terms 聚合
func (b *Builder) AddTermsAgg(name string, field string, size int) *Builder {
	terms := map[string]any{"field": field}
	if size > 0 {
		terms["size"] = size
	}
	b.aggs[name] = map[string]any{"terms": terms}
	return b
}

// AddAggregation 添加自定义聚合
func (b *Builder) AddAggregation(name string, agg map[string]any) *Builder {
	b.aggs[name] = agg
	return b
}

// buildQuery 组装 query 部分
func (b *Builder) buildQuery() map[string]any {
	boolQuery := make(map[string]any)
	if len(b.must) > 0 {
		boolQuery["must"] = b.must
	}
	if len(b.filter) > 0 {
		boolQuery["filter"] = b.filter
	}
	if len(b.should) > 0 {
		boolQuery["should"] = b.should
	}
	if len(b.mustNot) > 0 {
		boolQuery["must_not"] = b.mustNot
	}

	if len(boolQuery) == 0 {
		return map[string]any{"match_all": map[string]any{}}
	}
	return map[string]any{"bool": boolQuery}
}

// Build 构建最终的查询体
func (b *Builder) Build() Query {
	body := map[string]any{"query": b.buildQuery()}

	if b.from != nil {
		body["from"] = *b.from
	}
	if b.size != nil {
		body["size"] = *b.size
	}
	if b.minScore != nil {
		body["min_score"] = *b.minScore
	}
	if b.timeout != "" {
		body["timeout"] = b.timeout
	}
	if b.trackHits != nil {
		body["track_total_hits"] = b.trackHits
	}
	if len(b.sort) > 0 {
		body["sort"] = b.sort
	}
	if len(b.source) > 0 {
		body["_source"] = b.source
	}
	if len(b.highlight) > 0 {
		body["highlight"] = b.highlight
	}
	if len(b.script) > 0 {
		body["script_fields"] = b.script
	}
	if len(b.aggs) > 0 {
		body["aggs"] = b.aggs
	}

	return Query(body)
}

// clauseKind 区分过滤条件最终落在 filter 还是 must_not
type clauseKind int

const (
	clauseFilter clauseKind = iota
	clauseMustNot
)

type clause struct {
	kind clauseKind
	body map[string]any
}

// formatFilter 把 Filter 转换为 Elasticsearch 查询子句
func formatFilter(f Filter) (*clause, error) {
	if f.Field == "" {
		return nil, fmt.Errorf("%w: empty field", ErrInvalidQuery)
	}

	operator := strings.ToUpper(strings.TrimSpace(f.Operator))
	if operator == "" {
		operator = "="
	}

	switch operator {
	case "=":
		return &clause{kind: clauseFilter, body: map[string]any{"term": map[string]any{f.Field: f.Value}}}, nil
	case "!=":
		return &clause{kind: clauseMustNot, body: map[string]any{"term": map[string]any{f.Field: f.Value}}}, nil
	case ">", ">=", "<", "<=":
		return &clause{kind: clauseFilter, body: map[string]any{
			"range": map[string]any{f.Field: map[string]any{rangeKey(operator): f.Value}},
		}}, nil
	case "IN":
		return &clause{kind: clauseFilter, body: map[string]any{"terms": map[string]any{f.Field: toAnySlice(f.Value)}}}, nil
	case "NOT IN":
		return &clause{kind: clauseMustNot, body: map[string]any{"terms": map[string]any{f.Field: toAnySlice(f.Value)}}}, nil
	case "EXISTS":
		return &clause{kind: clauseFilter, body: map[string]any{"exists": map[string]any{"field": f.Field}}}, nil
	case "NOT EXISTS":
		return &clause{kind: clauseMustNot, body: map[string]any{"exists": map[string]any{"field": f.Field}}}, nil
	case "CONTAINS":
		return &clause{kind: clauseFilter, body: map[string]any{"match_phrase": map[string]any{f.Field: f.Value}}}, nil
	case "NOT CONTAINS":
		return &clause{kind: clauseMustNot, body: map[string]any{"match_phrase": map[string]any{f.Field: f.Value}}}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported operator %q", ErrInvalidQuery, f.Operator)
	}
}

func rangeKey(operator string) string {
	switch operator {
	case ">":
		return "gt"
	case ">=":
		return "gte"
	case "<":
		return "lt"
	default:
		return "lte"
	}
}

func toAnySlice(v any) []any {
	switch val := v.(type) {
	case []any:
		return val
	case []string:
		result := make([]any, len(val))
		for i, s := range val {
			result[i] = s
		}
		return result
	case []int:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result
	case []int64:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result
	case []float64:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result
	default:
		// 单值自动包装成数组，避免调用方重复处理
		if v == nil {
			return nil
		}
		return []any{v}
	}
}

// structToFilters 将结构体转换为过滤条件
func structToFilters(filterStruct any, builder *Builder) ([]Filter, error) {
	val := reflect.ValueOf(filterStruct)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, errors.New("input must not be nil")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct")
	}

	typ := val.Type()
	var filters []Filter

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// 非导出字段跳过
			continue
		}

		tag := field.Tag.Get("filter")
		if tag == "" || tag == "-" {
			continue
		}

		filterTag := parseFilterTag(tag)
		fieldValue := val.Field(i)

		if !fieldValue.IsValid() {
			continue
		}

		if isZeroOrEmpty(fieldValue) {
			if filterTag.Required {
				return nil, &RequiredFilterError{Field: filterTag.Field}
			}
			if filterTag.IgnoreZero || !strings.Contains(tag, "ignorezero=false") {
				continue
			}
		}

		if filterTag.Nested && fieldValue.Kind() == reflect.Struct {
			nestedFilters, err := structToFilters(fieldValue.Interface(), builder)
			if err != nil {
				return nil, fmt.Errorf("error processing nested struct field %s: %w", field.Name, err)
			}
			filters = append(filters, nestedFilters...)
			continue
		}

		if builder != nil && builder.transformers != nil {
			if transformer, exists := builder.transformers[filterTag.Field]; exists {
				fieldValue = reflect.ValueOf(transformer(fieldValue.Interface()))
			}
		}

		if builder != nil && builder.validators != nil {
			if validator, exists := builder.validators[filterTag.Field]; exists {
				if err := validator(fieldValue.Interface()); err != nil {
					return nil, fmt.Errorf("validation failed for field %s: %w", filterTag.Field, err)
				}
			}
		}

		if filterTag.DateFormat != "" && fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			dateValue := fieldValue.Interface().(time.Time)
			fieldValue = reflect.ValueOf(dateValue.Format(filterTag.DateFormat))
		}

		filters = append(filters, Filter{
			Field:    filterTag.Field,
			Operator: filterTag.Operator,
			Value:    fieldValue.Interface(),
		})
	}

	return filters, nil
}

// parseFilterTag 解析 filter tag，格式：field,op==,required,ignorezero,nested,datefmt=...
func parseFilterTag(tag string) FilterTag {
	parts := strings.Split(tag, ",")
	filter := FilterTag{Field: parts[0]}

	for _, part := range parts[1:] {
		switch {
		case part == "required":
			filter.Required = true
		case part == "ignorezero":
			filter.IgnoreZero = true
		case part == "ignorezero=false":
			filter.IgnoreZero = false
		case part == "nested":
			filter.Nested = true
		case strings.HasPrefix(part, "op="):
			filter.Operator = strings.TrimPrefix(part, "op=")
		case strings.HasPrefix(part, "group="):
			filter.Group = strings.TrimPrefix(part, "group=")
		case strings.HasPrefix(part, "transform="):
			filter.Transform = strings.TrimPrefix(part, "transform=")
		case strings.HasPrefix(part, "validate="):
			filter.Validate = strings.TrimPrefix(part, "validate=")
		case strings.HasPrefix(part, "deps="):
			filter.Deps = strings.Split(strings.TrimPrefix(part, "deps="), ";")
		case strings.HasPrefix(part, "datefmt="):
			filter.DateFormat = strings.TrimPrefix(part, "datefmt=")
		}
	}

	if filter.Operator == "" {
		filter.Operator = "="
	}
	return filter
}

// isZeroOrEmpty 检查值是否为零值或空值
func isZeroOrEmpty(val reflect.Value) bool {
	if reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface()) {
		return true
	}

	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	case reflect.Struct:
		if val.Type() == reflect.TypeOf(time.Time{}) {
			return val.Interface().(time.Time).IsZero()
		}
	}

	return false
}

// toString 把任意值转换为字符串，供脚本/模板类场景复用
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		return val.Format(time.RFC3339)
	case json.Number:
		return val.String()
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// RequiredFilterError 表示必填过滤字段未设置
type RequiredFilterError struct {
	Field string
}

func (e *RequiredFilterError) Error() string {
	return "required filter field '" + e.Field + "' is missing"
}
