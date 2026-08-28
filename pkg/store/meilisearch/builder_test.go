package meilisearch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_Pagination(t *testing.T) {
	builder := NewBuilder()

	request := builder.SetPage(2).SetHitsPerPage(50).Build()

	assert.Equal(t, int64(2), request.Page)
	if assert.NotNil(t, request.HitsPerPage) {
		assert.Equal(t, int64(50), *request.HitsPerPage)
	}
}

func TestBuilder_SetFacets(t *testing.T) {
	builder := NewBuilder()

	request := builder.
		AddFacets("brand").
		SetFacets("category", "genres").
		Build()

	assert.Equal(t, []string{"category", "genres"}, request.Facets)
}

func TestBuilder_ClearFacets(t *testing.T) {
	builder := NewBuilder()

	request := builder.
		AddFacets("brand", "category").
		ClearFacets().
		Build()

	assert.Empty(t, request.Facets)
}

func TestBuilder_AddFilterCondition(t *testing.T) {
	builder := NewBuilder()

	// 测试 AND 条件
	andCondition := FilterCondition{
		Filters: []Filter{
			{Field: "status", Operator: "=", Value: "active"},
			{Field: "verified", Operator: "=", Value: true},
		},
		Operator: "AND",
	}

	builder.AddFilterCondition(andCondition)
	request := builder.Build()

	// 验证过滤条件
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "status = \"active\"")
	assert.Contains(t, filter, "verified = \"true\"")
	assert.Contains(t, filter, "AND")
}

func TestBuilder_AddFilterCondition_OR(t *testing.T) {
	builder := NewBuilder()

	// 测试 OR 条件
	orCondition := FilterCondition{
		Filters: []Filter{
			{Field: "status", Operator: "=", Value: "active"},
			{Field: "status", Operator: "=", Value: "pending"},
		},
		Operator: "OR",
	}

	builder.AddFilterCondition(orCondition)
	request := builder.Build()

	// 验证过滤条件
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "status = \"active\"")
	assert.Contains(t, filter, "status = \"pending\"")
	assert.Contains(t, filter, "OR")
}

func TestBuilder_AddGeoFilter(t *testing.T) {
	builder := NewBuilder()

	geoFilter := GeoFilter{
		Point:  GeoPoint{Lat: 40.7128, Lng: -74.0060},
		Radius: 10000,
		Unit:   "m",
	}

	builder.AddGeoFilter(geoFilter)
	request := builder.Build()

	// 验证地理位置过滤
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "_geoRadius(40.712800, -74.006000, 10000.000000)")
}

func TestBuilder_AddDateRangeFilter(t *testing.T) {
	builder := NewBuilder()

	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	end := now

	dateFilter := DateRangeFilter{
		Field: "created_at",
		Start: start,
		End:   end,
	}

	builder.AddDateRangeFilter(dateFilter)
	request := builder.Build()

	// 验证日期范围过滤
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "created_at >= \""+start.Format(time.RFC3339)+"\"")
	assert.Contains(t, filter, "created_at <= \""+end.Format(time.RFC3339)+"\"")
}

func TestBuilder_AddArrayFilter(t *testing.T) {
	builder := NewBuilder()

	arrayFilter := ArrayFilter{
		Field:    "categories",
		Values:   []any{"electronics", "books"},
		Operator: "IN",
	}

	builder.AddArrayFilter(arrayFilter)
	request := builder.Build()

	// 验证数组过滤
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "categories IN [\"electronics\", \"books\"]")
}

func TestBuilder_AddFilterGroup(t *testing.T) {
	builder := NewBuilder()

	// 添加第一个条件组
	condition1 := FilterCondition{
		Filters: []Filter{
			{Field: "price", Operator: ">=", Value: 100},
		},
		Operator: "AND",
	}
	builder.AddFilterGroup("price", condition1)

	// 添加第二个条件组
	condition2 := FilterCondition{
		Filters: []Filter{
			{Field: "rating", Operator: ">=", Value: 4},
		},
		Operator: "AND",
	}
	builder.AddFilterGroup("quality", condition2)

	// 应用第一个条件组
	builder.ApplyFilterGroup("price")
	request := builder.Build()

	// 验证只有第一个条件组被应用
	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "price >= \"100\"")
	assert.NotContains(t, filter, "rating")
}

func TestBuilder_RegisterValidator(t *testing.T) {
	builder := NewBuilder()

	// 注册验证器
	builder.RegisterValidator("age", func(value any) error {
		if age, ok := value.(int); ok {
			if age < 0 || age > 150 {
				return assert.AnError
			}
		}
		return nil
	})

	// 验证验证器已注册
	assert.NotNil(t, builder.validators)
	assert.NotNil(t, builder.validators["age"])
}

func TestBuilder_RegisterTransformer(t *testing.T) {
	builder := NewBuilder()

	// 注册转换器
	builder.RegisterTransformer("email", func(value any) any {
		if email, ok := value.(string); ok {
			return "transformed_" + email
		}
		return value
	})

	// 验证转换器已注册
	assert.NotNil(t, builder.transformers)
	assert.NotNil(t, builder.transformers["email"])
}

func TestStructToFilters_NestedStruct(t *testing.T) {
	type NestedFilter struct {
		Name string `filter:"name"`
	}
	type MainFilter struct {
		Nested NestedFilter `filter:"nested,nested"`
	}

	mainFilter := MainFilter{
		Nested: NestedFilter{
			Name: "test",
		},
	}

	filters, err := structToFilters(mainFilter, nil)
	assert.NoError(t, err)
	assert.Len(t, filters, 1)
	assert.Equal(t, "name", filters[0].Field)
	assert.Equal(t, "test", filters[0].Value)
}

func TestStructToFilters_DateFormat(t *testing.T) {
	type DateFilter struct {
		Date time.Time `filter:"date,datefmt=2006-01-02"`
	}

	dateFilter := DateFilter{
		Date: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	filters, err := structToFilters(dateFilter, nil)
	assert.NoError(t, err)
	assert.Len(t, filters, 1)
	assert.Equal(t, "date", filters[0].Field)
	assert.Equal(t, "2023-05-15", filters[0].Value)
}

func TestParseFilterTag(t *testing.T) {
	// 测试基本标签
	tag := parseFilterTag("field")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, "=", tag.Operator)
	assert.False(t, tag.Required)

	// 测试带操作符的标签
	tag = parseFilterTag("field,op=CONTAINS")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, "CONTAINS", tag.Operator)

	// 测试带多个属性的标签
	tag = parseFilterTag("field,op=CONTAINS,required,ignorezero,nested")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, "CONTAINS", tag.Operator)
	assert.True(t, tag.Required)
	assert.True(t, tag.IgnoreZero)
	assert.True(t, tag.Nested)

	// 测试带分组的标签
	tag = parseFilterTag("field,group=group1")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, "group1", tag.Group)

	// 测试带日期格式的标签
	tag = parseFilterTag("field,datefmt=2006-01-02")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, "2006-01-02", tag.DateFormat)

	// 测试带依赖的标签
	tag = parseFilterTag("field,deps=field1;field2")
	assert.Equal(t, "field", tag.Field)
	assert.Equal(t, []string{"field1", "field2"}, tag.Dependencies)
}

func TestFormatFilter(t *testing.T) {
	// 测试基本操作符
	filter := Filter{Field: "status", Operator: "=", Value: "active"}
	result := formatFilter(filter)
	assert.Equal(t, "status = \"active\"", result)

	// 测试 IN 操作符
	filter = Filter{Field: "status", Operator: "IN", Value: []any{"active", "pending"}}
	result = formatFilter(filter)
	assert.Equal(t, "status IN [\"active\", \"pending\"]", result)

	// 测试 CONTAINS 操作符
	filter = Filter{Field: "name", Operator: "CONTAINS", Value: "john"}
	result = formatFilter(filter)
	assert.Equal(t, "name CONTAINS \"john\"", result)

	// 测试 EXISTS 操作符
	filter = Filter{Field: "email", Operator: "EXISTS", Value: nil}
	result = formatFilter(filter)
	assert.Equal(t, "email EXISTS", result)
}

func TestToString(t *testing.T) {
	// 测试字符串
	result := toString("test")
	assert.Equal(t, "test", result)

	// 测试整数
	result = toString(42)
	assert.Equal(t, "42", result)

	// 测试浮点数
	result = toString(3.14)
	assert.Equal(t, "3.14", result)

	// 测试布尔值
	result = toString(true)
	assert.Equal(t, "true", result)

	// 测试时间
	now := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	result = toString(now)
	assert.Equal(t, "2023-05-15T00:00:00Z", result)
}

func TestFacetSearchBuilder_BuildBasicRequest(t *testing.T) {
	request := NewFacetSearchBuilder().
		SetFacetName("genres").
		SetFacetQuery("Dra").
		SetQuery("batman").
		AttributesToSearchOn("title", "overview").
		SetMatchingStrategy("last").
		EnableExhaustiveFacetCount().
		Build()

	assert.Equal(t, "genres", request.FacetName)
	assert.Equal(t, "Dra", request.FacetQuery)
	assert.Equal(t, "batman", request.Q)
	assert.Equal(t, []string{"title", "overview"}, request.AttributesToSearchOn)
	assert.Equal(t, "last", request.MatchingStrategy)
	assert.True(t, request.ExhaustiveFacetCount)
}

func TestFacetSearchBuilder_AddFilter(t *testing.T) {
	request := NewFacetSearchBuilder().
		SetFacetName("brand").
		SetFacetQuery("app").
		AddFilter("status", "=", "online").
		Build()

	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Equal(t, "status = \"online\"", filter)
}

func TestFacetSearchBuilder_AddFilterCondition(t *testing.T) {
	condition := FilterCondition{
		Filters: []Filter{
			{Field: "year", Operator: ">", Value: 2000},
			{Field: "genres", Operator: "=", Value: "Drama"},
		},
		Operator: "AND",
	}

	request := NewFacetSearchBuilder().
		SetFacetName("genres").
		SetFacetQuery("Dra").
		AddFilterCondition(condition).
		Build()

	filter, ok := request.Filter.(string)
	assert.True(t, ok)
	assert.Contains(t, filter, "year > \"2000\"")
	assert.Contains(t, filter, "genres = \"Drama\"")
	assert.Contains(t, filter, "AND")
}

func TestFacetSearchBuilder_SetFilter(t *testing.T) {
	request := NewFacetSearchBuilder().
		SetFacetName("brand").
		SetFacetQuery("app").
		SetFilter("status = online").
		Build()

	assert.Equal(t, "status = online", request.Filter)
}
