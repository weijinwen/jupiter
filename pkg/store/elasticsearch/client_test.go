package elasticsearch

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_IndexDocument(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","result":"created","_shards":{"total":2,"successful":1,"failed":0},"_version":1,"_seq_no":0,"_primary_term":1}`))
	})
	client, err := testConfig("index-doc", srv.URL).Build()
	require.NoError(t, err)

	err = client.IndexDocument(t.Context(), "books", "1", map[string]any{"title": "Go 语言", "pages": 380})
	require.NoError(t, err)
}

func TestClient_UpdateDocument(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","result":"updated","_shards":{"total":2,"successful":1,"failed":0},"_version":2,"_seq_no":1,"_primary_term":1}`))
	})
	client, err := testConfig("update-doc", srv.URL).Build()
	require.NoError(t, err)

	err = client.UpdateDocument(t.Context(), "books", "1", map[string]any{"pages": 400})
	require.NoError(t, err)
}

func TestClient_DeleteDocument(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","result":"deleted","_shards":{"total":2,"successful":1,"failed":0},"_version":2,"_seq_no":1,"_primary_term":1}`))
	})
	client, err := testConfig("delete-doc", srv.URL).Build()
	require.NoError(t, err)

	err = client.DeleteDocument(t.Context(), "books", "1")
	require.NoError(t, err)
}

func TestClient_GetSource(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","found":true,"_source":{"title":"Go 语言","pages":380}}`))
	})

	client, err := testConfig("get-source", srv.URL).Build()
	require.NoError(t, err)

	var book struct {
		Title string `json:"title"`
		Pages int    `json:"pages"`
	}
	require.NoError(t, client.GetSource(t.Context(), "books", "1", &book))

	assert.Equal(t, "Go 语言", book.Title)
	assert.Equal(t, 380, book.Pages)
}

func TestClient_GetSource_NotFound(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"_index":"books","_id":"404","found":false}`))
	})

	client, err := testConfig("get-source-404", srv.URL).Build()
	require.NoError(t, err)

	var book struct {
		Title string `json:"title"`
	}
	err = client.GetSource(t.Context(), "books", "404", &book)

	assert.ErrorIs(t, err, ErrDocumentNotFound)
}

func TestClient_SearchInto(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"took": 3,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{"_index":"books","_id":"1","_source":{"title":"Go"}},
					{"_index":"books","_id":"2","_source":{"title":"Rust"}}
				]
			}
		}`))
	})

	client, err := testConfig("search-into", srv.URL).Build()
	require.NoError(t, err)

	type book struct {
		Title string `json:"title"`
	}
	books, err := client.SearchInto[book](t.Context(), "books", nil)

	require.NoError(t, err)
	require.Len(t, books, 2)
	assert.Equal(t, "Go", books[0].Title)
	assert.Equal(t, "Rust", books[1].Title)
}

// 命中列表为空时应返回空切片而非 nil，且不报错。
func TestClient_SearchInto_NoHits(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"took":1,"hits":{"total":{"value":0,"relation":"eq"},"hits":[]}}`))
	})

	client, err := testConfig("search-empty", srv.URL).Build()
	require.NoError(t, err)

	type book struct {
		Title string `json:"title"`
	}
	books, err := client.SearchInto[book](t.Context(), "books", nil)

	require.NoError(t, err)
	assert.NotNil(t, books)
	assert.Empty(t, books)
}
