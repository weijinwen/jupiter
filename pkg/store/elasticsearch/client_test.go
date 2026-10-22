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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient 基于 httptest 构造一个指向 mock 服务的客户端。
// 官方客户端会校验 X-Elastic-Product 响应头，这里统一补上。
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		handler(w, r)
	}))

	config := DefaultConfig()
	config.Name = "test-client"
	config.Addr = server.URL
	config.EnableMetric = false
	config.SlowThreshold = 0

	client, err := config.Build()
	require.NoError(t, err)

	return client, server.Close
}

func TestClient_Index(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer closeFn()

	index := client.Index("articles")

	assert.Equal(t, "articles", index.Name())
	assert.Equal(t, client, index.client)
}

func TestClient_Ping(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	defer closeFn()

	assert.NoError(t, client.Ping())
}

func TestIndex_Exists(t *testing.T) {
	t.Run("存在", func(t *testing.T) {
		client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		defer closeFn()

		exists, err := client.Index("articles").Exists()
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("不存在", func(t *testing.T) {
		client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		defer closeFn()

		exists, err := client.Index("articles").Exists()
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestIndex_CreateAndDelete(t *testing.T) {
	var created bool
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/articles":
			created = true
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/articles":
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer closeFn()

	index := client.Index("articles")
	require.NoError(t, index.Create(map[string]any{
		"mappings": map[string]any{"properties": map[string]any{"title": map[string]any{"type": "text"}}},
	}))
	assert.True(t, created)

	require.NoError(t, index.Delete())
}

func TestIndex_AddDocument(t *testing.T) {
	var gotBody map[string]any
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"_id":"1","result":"created"}`))
	})
	defer closeFn()

	res, err := client.Index("articles").AddDocument("1", map[string]any{"title": "hello"})
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, "hello", gotBody["title"])
}

func TestIndex_GetDocument(t *testing.T) {
	t.Run("命中", func(t *testing.T) {
		client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"_id":"1","_source":{"title":"hello","views":3}}`))
		})
		defer closeFn()

		var out struct {
			Title string `json:"title"`
			Views int    `json:"views"`
		}
		require.NoError(t, client.Index("articles").GetDocument("1", &out))
		assert.Equal(t, "hello", out.Title)
		assert.Equal(t, 3, out.Views)
	})

	t.Run("未命中", func(t *testing.T) {
		client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"_index":"articles","found":false}`))
		})
		defer closeFn()

		var out map[string]any
		err := client.Index("articles").GetDocument("1", &out)
		assert.True(t, errors.Is(err, ErrDocumentNotFound))
	})
}

func TestIndex_DocumentExists(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/articles/_doc/1" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer closeFn()

	index := client.Index("articles")

	exists, err := index.DocumentExists("1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = index.DocumentExists("2")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIndex_Search(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"took": 3,
			"timed_out": false,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"max_score": 1.5,
				"hits": [
					{"_id":"1","_source":{"title":"jupiter one"},"highlight":{"title":["<em>jupiter</em> one"]}},
					{"_id":"2","_source":{"title":"jupiter two"}}
				]
			}
		}`))
	})
	defer closeFn()

	var articles []struct {
		Title string `json:"title"`
	}

	result, err := client.Index("articles").Search(NewBuilder().SetQuery("jupiter"), &articles)
	require.NoError(t, err)

	assert.Equal(t, 3, result.Took)
	assert.Equal(t, int64(2), result.Hits.Total.Value)
	assert.Len(t, articles, 2)
	assert.Equal(t, "jupiter one", articles[0].Title)

	source, ok := result.Source(0)
	require.True(t, ok)
	assert.Contains(t, string(source), "jupiter one")

	highlight, ok := result.Highlights(0)
	require.True(t, ok)
	assert.Equal(t, []string{"<em>jupiter</em> one"}, highlight["title"])

	_, ok = result.Source(99)
	assert.False(t, ok)
}

func TestIndex_SearchIndexNotFound(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"type":"index_not_found_exception","reason":"no such index"},"status":404}`))
	})
	defer closeFn()

	_, err := client.Index("missing").Search(NewBuilder().Build(), nil)
	assert.True(t, errors.Is(err, ErrIndexNotFound))
}

func TestIndex_Count(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"count":42,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0}}`))
	})
	defer closeFn()

	count, err := client.Index("articles").Count(nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)

	count, err = client.Index("articles").Count(NewBuilder().AddFilter("status", "=", "published").Build())
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestIndex_AddDocuments(t *testing.T) {
	var lines []string
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		for _, line := range splitLines(string(raw)) {
			if line != "" {
				lines = append(lines, line)
			}
		}
		_, _ = w.Write([]byte(`{"took":2,"errors":false,"items":[{"index":{"_id":"1","status":201}},{"index":{"_id":"2","status":201}}]}`))
	})
	defer closeFn()

	type doc struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	result, err := client.Index("articles").AddDocuments([]doc{
		{ID: "1", Title: "a"},
		{ID: "2", Title: "b"},
	}, "id")
	require.NoError(t, err)

	assert.False(t, result.Errors)
	assert.Equal(t, 0, result.FailedCount())
	// 两条文档 -> 4 行 NDJSON
	assert.Len(t, lines, 4)

	var action map[string]map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &action))
	assert.Equal(t, "articles", action["index"]["_index"])
	assert.Equal(t, "1", action["index"]["_id"])
}

func TestIndex_UpdateDocuments(t *testing.T) {
	var lines []string
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		for _, line := range splitLines(string(raw)) {
			if line != "" {
				lines = append(lines, line)
			}
		}
		_, _ = w.Write([]byte(`{"took":1,"errors":false,"items":[{"update":{"_id":"1","status":200}}]}`))
	})
	defer closeFn()

	type doc struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	_, err := client.Index("articles").UpdateDocuments([]doc{{ID: "1", Title: "a"}}, "id")
	require.NoError(t, err)

	require.Len(t, lines, 2)
	// update 动作的文档行需要包一层 doc
	assert.Contains(t, lines[1], `"doc"`)
}

func TestIndex_AddDocumentsEmpty(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer closeFn()

	result, err := client.Index("articles").AddDocuments([]map[string]any{}, "id")
	require.NoError(t, err)
	assert.Equal(t, 0, result.FailedCount())
}

func TestIndex_DeleteDocument(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":"deleted"}`))
	})
	defer closeFn()

	res, err := client.Index("articles").DeleteDocument("1")
	require.NoError(t, err)
	defer res.Body.Close()
}

func TestIndex_DeleteByQuery(t *testing.T) {
	client, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"deleted":3}`))
	})
	defer closeFn()

	res, err := client.Index("articles").DeleteByQuery(Query{
		"query": map[string]any{"match_all": map[string]any{}},
	})
	require.NoError(t, err)
	defer res.Body.Close()

	res, err = client.Index("articles").DeleteAllDocuments()
	require.NoError(t, err)
	defer res.Body.Close()
}

func TestBulkResult_FailedCount(t *testing.T) {
	result := &BulkResult{
		Errors: true,
		Items: []map[string]struct {
			ID     string          `json:"_id"`
			Index  string          `json:"_index"`
			Status int             `json:"status"`
			Result string          `json:"result"`
			Error  json.RawMessage `json:"error"`
		}{
			{"index": {ID: "1", Status: 201}},
			{"index": {ID: "2", Status: 400, Error: json.RawMessage(`{"type":"x"}`)}},
		},
	}

	assert.Equal(t, 1, result.FailedCount())

	var nilResult *BulkResult
	assert.Equal(t, 0, nilResult.FailedCount())
}

func TestSearchResult_UnmarshalHits(t *testing.T) {
	var result *SearchResult
	assert.Error(t, result.UnmarshalHits(nil))

	result = &SearchResult{Hits: SearchHits{Hits: []json.RawMessage{
		json.RawMessage(`{"_source":{"title":"a"}}`),
		json.RawMessage(`{"_source":{"title":"b"}}`),
	}}}
	var out []struct {
		Title string `json:"title"`
	}
	require.NoError(t, result.UnmarshalHits(&out))
	assert.Len(t, out, 2)
	assert.Equal(t, "b", out[1].Title)
}

func TestToBodyReader(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		reader, err := toBodyReader(nil)
		require.NoError(t, err)
		assert.Nil(t, reader)
	})

	t.Run("Query", func(t *testing.T) {
		reader, err := toBodyReader(Query{"query": map[string]any{"match_all": map[string]any{}}})
		require.NoError(t, err)
		assert.NotNil(t, reader)
	})

	t.Run("Builder", func(t *testing.T) {
		reader, err := toBodyReader(NewBuilder())
		require.NoError(t, err)
		assert.NotNil(t, reader)
	})

	t.Run("map", func(t *testing.T) {
		reader, err := toBodyReader(map[string]any{"query": map[string]any{"match_all": map[string]any{}}})
		require.NoError(t, err)
		assert.NotNil(t, reader)
	})
}

// splitLines 按行切分 NDJSON
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
