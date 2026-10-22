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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toJSON 把查询体转换为可比较的 map
func toJSON(t *testing.T, query Query) map[string]any {
	t.Helper()

	buf, err := json.Marshal(query)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(buf, &out))
	return out
}

func TestBuilder_Pagination(t *testing.T) {
	query := NewBuilder().SetFrom(20).SetSize(10).Build()
	body := toJSON(t, query)

	assert.Equal(t, float64(20), body["from"])
	assert.Equal(t, float64(10), body["size"])
}

func TestBuilder_DefaultSize(t *testing.T) {
	body := toJSON(t, NewBuilder().Build())

	assert.Equal(t, float64(20), body["size"])
}

func TestBuilder_MatchAllWhenEmpty(t *testing.T) {
	body := toJSON(t, NewBuilder().Build())

	assert.Equal(t, map[string]any{"match_all": map[string]any{}}, body["query"])
}

func TestBuilder_SetQuery(t *testing.T) {
	t.Run("全字段查询", func(t *testing.T) {
		body := toJSON(t, NewBuilder().SetQuery("jupiter").Build())
		must := body["query"].(map[string]any)["bool"].(map[string]any)["must"].([]any)

		assert.Equal(t, map[string]any{
			"query_string": map[string]any{"query": "jupiter"},
		}, must[0])
	})

	t.Run("指定字段查询", func(t *testing.T) {
		body := toJSON(t, NewBuilder().SetQuery("jupiter", "title", "content").Build())
		must := body["query"].(map[string]any)["bool"].(map[string]any)["must"].([]any)

		assert.Equal(t, map[string]any{
			"multi_match": map[string]any{
				"query":  "jupiter",
				"fields": []any{"title", "content"},
			},
		}, must[0])
	})
}

func TestBuilder_AddFilter(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		value    any
		expected map[string]any
	}{
		{
			name:     "等值",
			operator: "=",
			value:    "published",
			expected: map[string]any{"term": map[string]any{"status": "published"}},
		},
		{
			name:     "大于等于",
			operator: ">=",
			value:    10,
			expected: map[string]any{"range": map[string]any{"age": map[string]any{"gte": float64(10)}}},
		},
		{
			name:     "小于",
			operator: "<",
			value:    100,
			expected: map[string]any{"range": map[string]any{"age": map[string]any{"lt": float64(100)}}},
		},
		{
			name:     "IN",
			operator: "IN",
			value:    []string{"a", "b"},
			expected: map[string]any{"terms": map[string]any{"tags": []any{"a", "b"}}},
		},
		{
			name:     "EXISTS",
			operator: "EXISTS",
			value:    nil,
			expected: map[string]any{"exists": map[string]any{"field": "status"}},
		},
		{
			name:     "CONTAINS",
			operator: "CONTAINS",
			value:    "hello world",
			expected: map[string]any{"match_phrase": map[string]any{"status": "hello world"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := "status"
			if tt.operator == ">=" || tt.operator == "<" {
				field = "age"
			}
			if tt.operator == "IN" {
				field = "tags"
			}

			body := toJSON(t, NewBuilder().AddFilter(field, tt.operator, tt.value).Build())
			filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)

			assert.Equal(t, tt.expected, filter[0])
		})
	}
}

func TestBuilder_NegativeFilterGoesToMustNot(t *testing.T) {
	body := toJSON(t, NewBuilder().AddFilter("status", "!=", "deleted").Build())
	boolQuery := body["query"].(map[string]any)["bool"].(map[string]any)

	assert.NotContains(t, boolQuery, "filter")
	assert.Equal(t, []any{
		map[string]any{"term": map[string]any{"status": "deleted"}},
	}, boolQuery["must_not"])
}

func TestBuilder_InvalidOperator(t *testing.T) {
	// 非法 operator 应被忽略，不产生过滤条件
	body := toJSON(t, NewBuilder().AddFilter("status", "LIKE", "x").Build())

	assert.Equal(t, map[string]any{"match_all": map[string]any{}}, body["query"])
}

func TestBuilder_AddFilterCondition(t *testing.T) {
	t.Run("OR 组合", func(t *testing.T) {
		body := toJSON(t, NewBuilder().AddFilterCondition(FilterCondition{
			Operator: "OR",
			Filters: []Filter{
				{Field: "status", Operator: "=", Value: "a"},
				{Field: "status", Operator: "=", Value: "b"},
			},
		}).Build())

		filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
		should := filter[0].(map[string]any)["bool"].(map[string]any)["should"]

		assert.Len(t, should, 2)
	})

	t.Run("AND 组合", func(t *testing.T) {
		body := toJSON(t, NewBuilder().AddFilterCondition(FilterCondition{
			Operator: "AND",
			Filters: []Filter{
				{Field: "a", Operator: "=", Value: 1},
				{Field: "b", Operator: "=", Value: 2},
			},
		}).Build())

		filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
		must := filter[0].(map[string]any)["bool"].(map[string]any)["must"]

		assert.Len(t, must, 2)
	})

	t.Run("单条件平铺", func(t *testing.T) {
		body := toJSON(t, NewBuilder().AddFilterCondition(FilterCondition{
			Operator: "OR",
			Filters:  []Filter{{Field: "a", Operator: "=", Value: 1}},
		}).Build())

		filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)

		assert.Equal(t, map[string]any{"term": map[string]any{"a": float64(1)}}, filter[0])
	})
}

func TestBuilder_AddDateRangeFilter(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	body := toJSON(t, NewBuilder().AddDateRangeFilter(DateRangeFilter{
		Field: "created_at",
		Start: start,
		End:   end,
	}).Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	rng := filter[0].(map[string]any)["range"].(map[string]any)["created_at"].(map[string]any)

	assert.Equal(t, start.Format(time.RFC3339), rng["gte"])
	assert.Equal(t, end.Format(time.RFC3339), rng["lte"])
}

func TestBuilder_AddDateRangeFilterOnlyStart(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	body := toJSON(t, NewBuilder().AddDateRangeFilter(DateRangeFilter{
		Field: "created_at",
		Start: start,
	}).Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	rng := filter[0].(map[string]any)["range"].(map[string]any)["created_at"].(map[string]any)

	assert.Equal(t, start.Format(time.RFC3339), rng["gte"])
	assert.NotContains(t, rng, "lte")
}

func TestBuilder_AddGeoFilter(t *testing.T) {
	body := toJSON(t, NewBuilder().AddGeoFilter(GeoFilter{
		Field:  "location",
		Point:  GeoPoint{Lat: 31.2, Lng: 121.5},
		Radius: 1000,
	}).Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	geo := filter[0].(map[string]any)["geo_distance"].(map[string]any)

	assert.Equal(t, "1000m", geo["distance"])
	assert.Equal(t, map[string]any{"lat": 31.2, "lon": 121.5}, geo["location"])
}

func TestBuilder_AddArrayFilter(t *testing.T) {
	body := toJSON(t, NewBuilder().AddArrayFilter(ArrayFilter{
		Field:  "tags",
		Values: []any{"go", "es"},
	}).Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)

	assert.Equal(t, map[string]any{"terms": map[string]any{"tags": []any{"go", "es"}}}, filter[0])
}

func TestBuilder_Sort(t *testing.T) {
	body := toJSON(t, NewBuilder().AddSort("created_at", false).AddSort("id", true).Build())

	assert.Equal(t, []any{
		map[string]any{"created_at": map[string]any{"order": "desc"}},
		map[string]any{"id": map[string]any{"order": "asc"}},
	}, body["sort"])
}

func TestBuilder_Source(t *testing.T) {
	body := toJSON(t, NewBuilder().SetSource("title").AddSource("content").Build())

	assert.Equal(t, []any{"title", "content"}, body["_source"])
}

func TestBuilder_Highlight(t *testing.T) {
	body := toJSON(t, NewBuilder().
		AddHighlight("title", "content").
		HighlightPreTag("<em>").
		HighlightPostTag("</em>").
		Build())

	highlight := body["highlight"].(map[string]any)
	fields := highlight["fields"].(map[string]any)

	assert.Contains(t, fields, "title")
	assert.Contains(t, fields, "content")
	assert.Equal(t, []any{"<em>"}, highlight["pre_tags"])
	assert.Equal(t, []any{"</em>"}, highlight["post_tags"])
}

func TestBuilder_Aggregation(t *testing.T) {
	body := toJSON(t, NewBuilder().
		AddTermsAgg("by_status", "status", 10).
		AddAggregation("avg_price", map[string]any{"avg": map[string]any{"field": "price"}}).
		Build())

	aggs := body["aggs"].(map[string]any)

	assert.Equal(t, map[string]any{"terms": map[string]any{"field": "status", "size": float64(10)}}, aggs["by_status"])
	assert.Contains(t, aggs, "avg_price")
}

func TestBuilder_MiscOptions(t *testing.T) {
	body := toJSON(t, NewBuilder().
		SetMinScore(1.5).
		SetTimeout("1s").
		TrackTotalHits(true).
		Build())

	assert.Equal(t, 1.5, body["min_score"])
	assert.Equal(t, "1s", body["timeout"])
	assert.Equal(t, true, body["track_total_hits"])
}

func TestBuilder_AddFilterStruct(t *testing.T) {
	type filterStruct struct {
		Status string    `filter:"status"`
		Age    int       `filter:"age,op=>="`
		Empty  string    `filter:"empty"`
		Skip   string    `filter:"-"`
		Tags   []string  `filter:"tags,op=IN"`
		Time   time.Time `filter:"created_at,datefmt=2006-01-02"`
	}

	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	body := toJSON(t, NewBuilder().AddFilterStruct(&filterStruct{
		Status: "published",
		Age:    18,
		Tags:   []string{"go"},
		Time:   now,
	}).Build())

	boolQuery := body["query"].(map[string]any)["bool"].(map[string]any)
	filter := boolQuery["filter"].([]any)

	require.Len(t, filter, 4)
	assert.Equal(t, map[string]any{"term": map[string]any{"status": "published"}}, filter[0])
	assert.Equal(t, map[string]any{"range": map[string]any{"age": map[string]any{"gte": float64(18)}}}, filter[1])
	assert.Equal(t, map[string]any{"terms": map[string]any{"tags": []any{"go"}}}, filter[2])
	assert.Equal(t, map[string]any{"term": map[string]any{"created_at": "2024-06-01"}}, filter[3])
}

func TestBuilder_AddFilterStructRequired(t *testing.T) {
	type filterStruct struct {
		Status string `filter:"status,required"`
	}

	// required 字段为零值时，不产生过滤条件（错误被吞掉并记录日志）
	body := toJSON(t, NewBuilder().AddFilterStruct(&filterStruct{}).Build())

	assert.Equal(t, map[string]any{"match_all": map[string]any{}}, body["query"])
}

func TestBuilder_AddFilterStructWithGroup(t *testing.T) {
	type filterStruct struct {
		A int `filter:"a"`
		B int `filter:"b"`
	}

	body := toJSON(t, NewBuilder().
		AddFilterStructWithGroup("base", &filterStruct{A: 1, B: 2}).
		ApplyFilterGroup("base").
		Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	must := filter[0].(map[string]any)["bool"].(map[string]any)["must"]

	assert.Len(t, must, 2)
}

func TestBuilder_ApplyFilterGroupUnknown(t *testing.T) {
	body := toJSON(t, NewBuilder().ApplyFilterGroup("missing").Build())

	assert.Equal(t, map[string]any{"match_all": map[string]any{}}, body["query"])
}

func TestBuilder_RegisterValidatorAndTransformer(t *testing.T) {
	type filterStruct struct {
		Name string `filter:"name"`
	}

	validatorErr := errors.New("bad name")
	builder := NewBuilder().
		RegisterValidator("name", func(v any) error {
			if v.(string) == "bad" {
				return validatorErr
			}
			return nil
		})

	// 校验失败时整个过滤条件被丢弃
	body := toJSON(t, builder.AddFilterStruct(&filterStruct{Name: "bad"}).Build())
	assert.Equal(t, map[string]any{"match_all": map[string]any{}}, body["query"])

	builder = NewBuilder().
		RegisterTransformer("name", func(v any) any {
			return "transformed"
		})
	body = toJSON(t, builder.AddFilterStruct(&filterStruct{Name: "original"}).Build())

	filter := body["query"].(map[string]any)["bool"].(map[string]any)["filter"].([]any)
	assert.Equal(t, map[string]any{"term": map[string]any{"name": "transformed"}}, filter[0])
}

func TestFormatFilterInvalid(t *testing.T) {
	_, err := formatFilter(Filter{Field: "", Operator: "=", Value: 1})
	assert.True(t, errors.Is(err, ErrInvalidQuery))

	_, err = formatFilter(Filter{Field: "a", Operator: "LIKE", Value: 1})
	assert.True(t, errors.Is(err, ErrInvalidQuery))
}

func TestQueryReader(t *testing.T) {
	reader, err := Query{"query": map[string]any{"match_all": map[string]any{}}}.Reader()
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.NewDecoder(reader).Decode(&out))
	assert.Contains(t, out, "query")
}
