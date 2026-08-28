package meilisearch

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
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
	Field    string
	Operator string
	Required bool
	// 新增字段
	Group        string   // 用于分组过滤条件
	Transform    string   // 值转换函数
	Validate     string   // 验证函数
	IgnoreZero   bool     // 是否忽略零值
	Dependencies []string // 依赖字段
	Nested       bool     // 是否嵌套结构体
	DateFormat   string   // 日期格式
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
	Operator string // "IN", "NOT IN", "CONTAINS", "NOT CONTAINS"
}

// Builder 用于构建 MeiliSearch 查询
type Builder struct {
	params meilisearch.SearchRequest
	// 新增字段
	filterGroups map[string][]FilterCondition // 分组过滤条件
	validators   map[string]func(any) error   // 自定义验证器
	transformers map[string]func(any) any     // 自定义转换器
}

// NewBuilder 创建新的 Builder 实例
func NewBuilder() *Builder {
	return &Builder{
		params: meilisearch.SearchRequest{
			Page:        0,
			HitsPerPage: meilisearch.Int64Ptr(20),
		},
		filterGroups: make(map[string][]FilterCondition),
		validators:   make(map[string]func(any) error),
		transformers: make(map[string]func(any) any),
	}
}

// SetQuery 设置搜索查询字符串
func (b *Builder) SetQuery(query string) *Builder {
	b.params.Query = query
	return b
}

// SetPage 设置页码
func (b *Builder) SetPage(page int64) *Builder {
	b.params.Page = int64(page)
	return b
}

// SetHitsPerPage 设置每页结果数量
func (b *Builder) SetHitsPerPage(hitsPerPage int64) *Builder {
	b.params.HitsPerPage = meilisearch.Int64Ptr(hitsPerPage)
	return b
}

func (b *Builder) AddFilter(field string, operator string, value any) *Builder {
	if !isValidOperator(operator) {
		xlog.Jupiter().Warn("invalid operator", xlog.String("operator", operator), xlog.String("field", field))
		return b
	}

	filterStr := formatFilter(Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})

	b.appendFilter(filterStr)
	return b
}

func (b *Builder) AddFilterCondition(condition FilterCondition) *Builder {
	var filterStrs []string

	for _, filter := range condition.Filters {
		filterStrs = append(filterStrs, formatFilter(filter))
	}

	var combinedFilter string
	if condition.Operator == "OR" {
		combinedFilter = "(" + strings.Join(filterStrs, " OR ") + ")"
	} else {
		combinedFilter = "(" + strings.Join(filterStrs, " AND ") + ")"
	}

	b.appendFilter(combinedFilter)
	return b
}

// AddFilterGroup 添加分组过滤条件
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

func (b *Builder) AddGeoFilter(geoFilter GeoFilter) *Builder {
	if geoFilter.Unit == "" {
		geoFilter.Unit = "m"
	}

	filterStr := fmt.Sprintf("_geoRadius(%f, %f, %f)",
		geoFilter.Point.Lat, geoFilter.Point.Lng, geoFilter.Radius)

	b.appendFilter(filterStr)
	return b
}

// AddDateRangeFilter 添加日期范围过滤
func (b *Builder) AddDateRangeFilter(dateFilter DateRangeFilter) *Builder {
	startFilter := Filter{
		Field:    dateFilter.Field,
		Operator: ">=",
		Value:    dateFilter.Start.Format(time.RFC3339),
	}

	endFilter := Filter{
		Field:    dateFilter.Field,
		Operator: "<=",
		Value:    dateFilter.End.Format(time.RFC3339),
	}

	condition := FilterCondition{
		Filters:  []Filter{startFilter, endFilter},
		Operator: "AND",
	}

	return b.AddFilterCondition(condition)
}

// AddArrayFilter 添加数组过滤
func (b *Builder) AddArrayFilter(arrayFilter ArrayFilter) *Builder {
	if arrayFilter.Operator == "" {
		arrayFilter.Operator = "IN"
	}

	return b.AddFilter(arrayFilter.Field, arrayFilter.Operator, arrayFilter.Values)
}

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

func (b *Builder) AddFilterStructWithGroup(groupName string, filterStruct any) *Builder {
	filters, err := structToFilters(filterStruct, b)
	if err != nil {
		xlog.Jupiter().Warn("struct to filters failed", xlog.FieldErr(err))
		return b
	}

	if len(filters) == 0 {
		return b
	}

	condition := FilterCondition{
		Filters:  filters,
		Operator: "AND",
	}

	return b.AddFilterGroup(groupName, condition)
}

// RegisterValidator 注册自定义验证器
func (b *Builder) RegisterValidator(field string, validator func(any) error) *Builder {
	if b.validators == nil {
		b.validators = make(map[string]func(any) error)
	}
	b.validators[field] = validator
	return b
}

// RegisterTransformer 注册自定义转换器
func (b *Builder) RegisterTransformer(field string, transformer func(any) any) *Builder {
	if b.transformers == nil {
		b.transformers = make(map[string]func(any) any)
	}
	b.transformers[field] = transformer
	return b
}

// AddFacets 添加分面字段
func (b *Builder) AddFacets(facets ...string) *Builder {
	b.params.Facets = append(b.params.Facets, facets...)
	return b
}

// SetFacets 设置分面字段，会覆盖已有分面字段
func (b *Builder) SetFacets(facets ...string) *Builder {
	b.params.Facets = facets
	return b
}

// ClearFacets 清空分面字段
func (b *Builder) ClearFacets() *Builder {
	b.params.Facets = nil
	return b
}

// AddSort 添加排序字段
func (b *Builder) AddSort(field string, ascending bool) *Builder {
	direction := "asc"
	if !ascending {
		direction = "desc"
	}
	b.params.Sort = append(b.params.Sort, field+":"+direction)
	return b
}

func (b *Builder) SetRankingScoreThreshold(rankingScoreThreshold float64) *Builder {
	b.params.RankingScoreThreshold = rankingScoreThreshold
	return b
}

// AddHighlight 添加高亮字段
func (b *Builder) AddHighlight(fields ...string) *Builder {
	b.params.AttributesToHighlight = append(b.params.AttributesToHighlight, fields...)
	return b
}

// HighlightPreTag 设置高亮标签前缀
func (b *Builder) HighlightPreTag(preTag string) *Builder {
	b.params.HighlightPreTag = preTag
	return b
}

// HighlightPostTag 设置高亮标签后缀
func (b *Builder) HighlightPostTag(postTag string) *Builder {
	b.params.HighlightPostTag = postTag
	return b
}

// AddAttributesToRetrieve 添加要返回的字段
func (b *Builder) AddAttributesToRetrieve(fields ...string) *Builder {
	b.params.AttributesToRetrieve = append(b.params.AttributesToRetrieve, fields...)
	return b
}

// AttributesToSearchOn 搜索字段
func (b *Builder) AttributesToSearchOn(fields ...string) *Builder {
	b.params.AttributesToSearchOn = append(b.params.AttributesToSearchOn, fields...)
	return b
}

// EnableMatches 启用匹配信息
func (b *Builder) EnableMatches() *Builder {
	b.params.ShowMatchesPosition = true
	return b
}

// ShowRankingScore 显示排名分数
func (b *Builder) ShowRankingScore() *Builder {
	b.params.ShowRankingScore = true
	return b
}

// ShowRankingScoreDetails 显示排名分数详细信息
func (b *Builder) ShowRankingScoreDetails() *Builder {
	b.params.ShowRankingScoreDetails = true
	return b
}

// Build 构建最终的搜索参数
func (b *Builder) Build() meilisearch.SearchRequest {
	// 移除对私有方法 validate() 的调用
	return b.params
}

// FacetSearchBuilder 用于构建 MeiliSearch 分面搜索请求
// 分面搜索用于在某个 facet 字段的值集合中做 autocomplete。
type FacetSearchBuilder struct {
	params meilisearch.FacetSearchRequest
}

// NewFacetSearchBuilder 创建新的 FacetSearchBuilder 实例
func NewFacetSearchBuilder() *FacetSearchBuilder {
	return &FacetSearchBuilder{}
}

// SetFacetName 设置要搜索的分面字段名
func (b *FacetSearchBuilder) SetFacetName(name string) *FacetSearchBuilder {
	b.params.FacetName = name
	return b
}

// SetFacetQuery 设置分面值搜索词
func (b *FacetSearchBuilder) SetFacetQuery(query string) *FacetSearchBuilder {
	b.params.FacetQuery = query
	return b
}

// SetQuery 设置主搜索词，用于限定分面搜索上下文
func (b *FacetSearchBuilder) SetQuery(query string) *FacetSearchBuilder {
	b.params.Q = query
	return b
}

// SetFilter 设置过滤条件，会覆盖已有过滤条件
func (b *FacetSearchBuilder) SetFilter(filter any) *FacetSearchBuilder {
	b.params.Filter = filter
	return b
}

// AddFilter 添加过滤条件
func (b *FacetSearchBuilder) AddFilter(field string, operator string, value any) *FacetSearchBuilder {
	if !isValidOperator(operator) {
		xlog.Jupiter().Warn("invalid operator", xlog.String("operator", operator), xlog.String("field", field))
		return b
	}

	filterStr := formatFilter(Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})

	b.appendFilter(filterStr)
	return b
}

// AddFilterCondition 添加组合过滤条件
func (b *FacetSearchBuilder) AddFilterCondition(condition FilterCondition) *FacetSearchBuilder {
	var filterStrs []string

	for _, filter := range condition.Filters {
		filterStrs = append(filterStrs, formatFilter(filter))
	}

	var combinedFilter string
	if condition.Operator == "OR" {
		combinedFilter = "(" + strings.Join(filterStrs, " OR ") + ")"
	} else {
		combinedFilter = "(" + strings.Join(filterStrs, " AND ") + ")"
	}

	b.appendFilter(combinedFilter)
	return b
}

// AttributesToSearchOn 设置用于限定主搜索词搜索范围的字段
func (b *FacetSearchBuilder) AttributesToSearchOn(fields ...string) *FacetSearchBuilder {
	b.params.AttributesToSearchOn = append(b.params.AttributesToSearchOn, fields...)
	return b
}

// SetMatchingStrategy 设置匹配策略
func (b *FacetSearchBuilder) SetMatchingStrategy(strategy string) *FacetSearchBuilder {
	b.params.MatchingStrategy = strategy
	return b
}

// EnableExhaustiveFacetCount 启用精确分面计数
func (b *FacetSearchBuilder) EnableExhaustiveFacetCount() *FacetSearchBuilder {
	b.params.ExhaustiveFacetCount = true
	return b
}

// Build 构建最终的分面搜索参数
func (b *FacetSearchBuilder) Build() meilisearch.FacetSearchRequest {
	return b.params
}

func (b *FacetSearchBuilder) appendFilter(filterStr string) {
	switch currentFilter := b.params.Filter.(type) {
	case nil:
		b.params.Filter = filterStr
	case string:
		b.params.Filter = []string{currentFilter, filterStr}
	case []string:
		b.params.Filter = append(currentFilter, filterStr)
	default:
		b.params.Filter = []string{fmt.Sprintf("%v", currentFilter), filterStr}
	}
}

var validOperators = map[string]bool{
	"=": true, "!=": true, ">": true, "<": true,
	">=": true, "<=": true, "IN": true, "NOT IN": true,
	"EXISTS": true, "NOT EXISTS": true, "CONTAINS": true, "NOT CONTAINS": true,
}

func isValidOperator(operator string) bool {
	return validOperators[operator]
}

func (b *Builder) appendFilter(filterStr string) {
	switch currentFilter := b.params.Filter.(type) {
	case nil:
		b.params.Filter = filterStr
	case string:
		b.params.Filter = []string{currentFilter, filterStr}
	case []string:
		b.params.Filter = append(currentFilter, filterStr)
	default:
		b.params.Filter = []string{fmt.Sprintf("%v", currentFilter), filterStr}
	}
}

func formatFilter(f Filter) string {
	switch f.Operator {
	case "IN", "NOT IN":
		values := toAnySlice(f.Value)
		var quotedValues []string
		for _, v := range values {
			quotedValues = append(quotedValues, strconv.Quote(toString(v)))
		}
		return f.Field + " " + f.Operator + " [" + strings.Join(quotedValues, ", ") + "]"
	case "CONTAINS", "NOT CONTAINS":
		return f.Field + " " + f.Operator + " " + strconv.Quote(toString(f.Value))
	case "EXISTS", "NOT EXISTS":
		return f.Field + " " + f.Operator
	default:
		return f.Field + " " + f.Operator + " " + strconv.Quote(toString(f.Value))
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
		return nil
	}
}

// structToFilters 将结构体转换为过滤条件
func structToFilters(filterStruct any, builder *Builder) ([]Filter, error) {
	var filters []Filter

	val := reflect.ValueOf(filterStruct)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("filter")

		if tag == "" || tag == "-" {
			continue
		}

		filterTag := parseFilterTag(tag)
		fieldValue := val.Field(i)

		// 检查字段是否有效
		if !fieldValue.IsValid() {
			continue
		}

		// 检查零值和空值
		if isZeroOrEmpty(fieldValue) {
			if filterTag.Required {
				return nil, &RequiredFilterError{Field: filterTag.Field}
			}
			// 默认行为：忽略零值和空值，除非明确设置了 IgnoreZero=false
			if filterTag.IgnoreZero || !strings.Contains(tag, "ignorezero=false") {
				continue
			}
		}

		// 处理嵌套结构体
		if filterTag.Nested && fieldValue.Kind() == reflect.Struct {
			nestedFilters, err := structToFilters(fieldValue.Interface(), builder)
			if err != nil {
				return nil, fmt.Errorf("error processing nested struct field %s: %v", field.Name, err)
			}
			filters = append(filters, nestedFilters...)
			continue
		}

		// 应用自定义转换器
		if builder != nil && builder.transformers != nil {
			if transformer, exists := builder.transformers[filterTag.Field]; exists {
				fieldValue = reflect.ValueOf(transformer(fieldValue.Interface()))
			}
		}

		// 应用自定义验证器
		if builder != nil && builder.validators != nil {
			if validator, exists := builder.validators[filterTag.Field]; exists {
				if err := validator(fieldValue.Interface()); err != nil {
					return nil, fmt.Errorf("validation failed for field %s: %v", filterTag.Field, err)
				}
			}
		}

		// 处理日期格式
		if filterTag.DateFormat != "" {
			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				dateValue := fieldValue.Interface().(time.Time)
				fieldValue = reflect.ValueOf(dateValue.Format(filterTag.DateFormat))
			}
		}

		filter := Filter{
			Field:    filterTag.Field,
			Operator: filterTag.Operator,
			Value:    fieldValue.Interface(),
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

// parseFilterTag 解析filter tag
func parseFilterTag(tag string) FilterTag {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return FilterTag{}
	}

	filter := FilterTag{
		Field: parts[0],
	}

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
			deps := strings.TrimPrefix(part, "deps=")
			filter.Dependencies = strings.Split(deps, ";")
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
	// 检查是否为零值
	if reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface()) {
		return true
	}

	// 针对特定类型进行额外检查
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	case reflect.Struct:
		// 特殊处理 time.Time
		if val.Type() == reflect.TypeOf(time.Time{}) {
			return val.Interface().(time.Time).IsZero()
		}
	}

	return false
}

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
	case uint32:
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
