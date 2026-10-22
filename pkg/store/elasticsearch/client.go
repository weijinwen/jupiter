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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client 是 Elasticsearch 客户端的封装，内嵌官方 *elasticsearch.Client，
// 在其基础上提供按索引名路由的便捷方法与统一的错误处理。
type Client struct {
	*elastic.Client
	config *Config
}

// Config 返回客户端使用的配置
func (c *Client) Config() *Config {
	return c.config
}

// Index 返回指定索引的索引管理器，所有文档与搜索操作均通过它进行。
func (c *Client) Index(name string) *Index {
	return &Index{client: c, name: name}
}

// Ping 检查集群是否可用
func (c *Client) Ping() error {
	res, err := c.Client.Ping(c.Client.Ping.WithContext(context.Background()))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return responseError(res)
	}
	return nil
}

// Refresh 刷新指定索引，使最近的写入对搜索可见
func (c *Client) Refresh(indices ...string) error {
	res, err := c.Client.Indices.Refresh(
		c.Client.Indices.Refresh.WithIndex(indices...),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return responseError(res)
	}
	return nil
}

// Index 索引管理器，绑定单个索引名
type Index struct {
	client *Client
	name   string
}

// Name 返回索引名
func (i *Index) Name() string {
	return i.name
}

// Exists 判断索引是否存在
func (i *Index) Exists() (bool, error) {
	res, err := i.client.Client.Indices.Exists(
		[]string{i.name},
		i.client.Client.Indices.Exists.WithContext(context.Background()),
	)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, responseError(res)
	}
}

// Create 创建索引，mapping 为索引设置与映射，传 nil 则使用默认配置
func (i *Index) Create(mapping map[string]any) error {
	opts := []func(*esapi.IndicesCreateRequest){
		i.client.Client.Indices.Create.WithContext(context.Background()),
	}
	if mapping != nil {
		body, err := json.Marshal(mapping)
		if err != nil {
			return fmt.Errorf("marshal mapping: %w", err)
		}
		opts = append(opts, i.client.Client.Indices.Create.WithBody(bytes.NewReader(body)))
	}

	res, err := i.client.Client.Indices.Create(i.name, opts...)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return responseError(res)
	}
	return nil
}

// Delete 删除索引
func (i *Index) Delete() error {
	res, err := i.client.Client.Indices.Delete(
		[]string{i.name},
		i.client.Client.Indices.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return responseError(res)
	}
	return nil
}

// AddDocument 添加（或整条覆盖）单个文档，id 为空时由 ES 自动生成
func (i *Index) AddDocument(id string, document any) (*esapi.Response, error) {
	body, err := marshalBody(document)
	if err != nil {
		return nil, err
	}

	opts := []func(*esapi.IndexRequest){
		i.client.Client.Index.WithContext(context.Background()),
	}
	if id != "" {
		opts = append(opts, i.client.Client.Index.WithDocumentID(id))
	}

	res, err := i.client.Client.Index(i.name, body, opts...)
	if err != nil {
		return nil, err
	}
	return res, checkResponse(res)
}

// AddDocuments 批量添加文档，文档需实现 id 提取逻辑。
// idField 为文档中作为 _id 的字段名，传空则交给 ES 自动生成。
func (i *Index) AddDocuments(documents any, idField string) (*BulkResult, error) {
	return i.bulk(documents, idField, "index")
}

// UpdateDocuments 批量更新文档（仅更新传入字段）
func (i *Index) UpdateDocuments(documents any, idField string) (*BulkResult, error) {
	return i.bulk(documents, idField, "update")
}

// UpdateDocument 按 id 局部更新单个文档，document 可以是 map 或带 doc 字段的结构体
func (i *Index) UpdateDocument(id string, document any) (*esapi.Response, error) {
	payload := document
	if _, ok := document.(map[string]any); !ok {
		payload = map[string]any{"doc": document}
	}

	body, err := marshalBody(payload)
	if err != nil {
		return nil, err
	}

	res, err := i.client.Client.Update(
		i.name,
		id,
		body,
		i.client.Client.Update.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}
	return res, checkResponse(res)
}

// GetDocument 按 id 获取单个文档，结果反序列化到 result
func (i *Index) GetDocument(id string, result any) error {
	res, err := i.client.Client.Get(
		i.name,
		id,
		i.client.Client.Get.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return ErrDocumentNotFound
	}
	if res.IsError() {
		return responseError(res)
	}

	return decodeSource(res.Body, result)
}

// Exists 判断文档是否存在
func (i *Index) DocumentExists(id string) (bool, error) {
	res, err := i.client.Client.Exists(
		i.name,
		id,
		i.client.Client.Exists.WithContext(context.Background()),
	)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, responseError(res)
	}
}

// DeleteDocument 按 id 删除单个文档
func (i *Index) DeleteDocument(id string) (*esapi.Response, error) {
	res, err := i.client.Client.Delete(
		i.name,
		id,
		i.client.Client.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}
	return res, checkResponse(res)
}

// DeleteByQuery 按查询条件批量删除文档
func (i *Index) DeleteByQuery(query Query) (*esapi.Response, error) {
	body, err := query.Reader()
	if err != nil {
		return nil, err
	}

	res, err := i.client.Client.DeleteByQuery(
		[]string{i.name},
		body,
		i.client.Client.DeleteByQuery.WithContext(context.Background()),
		i.client.Client.DeleteByQuery.WithRefresh(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return res, checkResponse(res)
}

// DeleteAllDocuments 删除索引中的全部文档（保留索引本身）
func (i *Index) DeleteAllDocuments() (*esapi.Response, error) {
	return i.DeleteByQuery(Query{"query": map[string]any{"match_all": map[string]any{}}})
}

// Search 执行搜索，query 可以是 Query、map[string]any 或 Builder
func (i *Index) Search(query any, result any) (*SearchResult, error) {
	body, err := toBodyReader(query)
	if err != nil {
		return nil, err
	}

	res, err := i.client.Client.Search(
		i.client.Client.Search.WithContext(context.Background()),
		i.client.Client.Search.WithIndex(i.name),
		i.client.Client.Search.WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, responseError(res)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var searchResult SearchResult
	if err := json.Unmarshal(raw, &searchResult); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}
	searchResult.Raw = raw

	if result != nil {
		if err := searchResult.UnmarshalHits(result); err != nil {
			return nil, err
		}
	}

	return &searchResult, nil
}

// Count 统计满足条件的文档数，query 传 nil 表示统计全部
func (i *Index) Count(query any) (int64, error) {
	var body io.Reader
	if query != nil {
		b, err := toBodyReader(query)
		if err != nil {
			return 0, err
		}
		body = b
	}

	res, err := i.client.Client.Count(
		i.client.Client.Count.WithContext(context.Background()),
		i.client.Client.Count.WithIndex(i.name),
		i.client.Client.Count.WithBody(body),
	)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, responseError(res)
	}

	var payload struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return 0, fmt.Errorf("decode count response: %w", err)
	}
	return payload.Count, nil
}

// bulk 组装 NDJSON 请求体并执行批量操作，action 为 index 或 update
func (i *Index) bulk(documents any, idField string, action string) (*BulkResult, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	docs, err := normalizeDocuments(documents)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return &BulkResult{}, nil
	}

	for _, doc := range docs {
		meta := map[string]any{"_index": i.name}
		if idField != "" {
			if id, ok := fieldValue(doc, idField); ok && id != nil && fmt.Sprintf("%v", id) != "" {
				meta["_id"] = fmt.Sprintf("%v", id)
			}
		}

		// action 行
		if err := encoder.Encode(map[string]any{action: meta}); err != nil {
			return nil, fmt.Errorf("encode bulk action: %w", err)
		}

		// 文档行：update 需要包一层 doc
		payload := doc
		if action == "update" {
			payload = map[string]any{"doc": doc}
		}
		if err := encoder.Encode(payload); err != nil {
			return nil, fmt.Errorf("encode bulk document: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("flush bulk body: %w", err)
	}

	res, err := i.client.Client.Bulk(
		bytes.NewReader(buf.Bytes()),
		i.client.Client.Bulk.WithContext(context.Background()),
		i.client.Client.Bulk.WithIndex(i.name),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, responseError(res)
	}

	var result BulkResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode bulk response: %w", err)
	}
	return &result, nil
}

// BulkResult 是批量操作的响应，Errors 为 true 时表示存在失败条目
type BulkResult struct {
	Took   int  `json:"took"`
	Errors bool `json:"errors"`
	Items  []map[string]struct {
		ID     string          `json:"_id"`
		Index  string          `json:"_index"`
		Status int             `json:"status"`
		Result string          `json:"result"`
		Error  json.RawMessage `json:"error"`
	} `json:"items"`
}

// FailedCount 返回失败的条目数量
func (r *BulkResult) FailedCount() int {
	if r == nil {
		return 0
	}
	count := 0
	for _, item := range r.Items {
		for _, v := range item {
			if v.Status >= 300 || len(v.Error) > 0 {
				count++
			}
		}
	}
	return count
}

// SearchResult 是搜索响应
type SearchResult struct {
	Took     int             `json:"took"`
	TimedOut bool            `json:"timed_out"`
	Hits     SearchHits      `json:"hits"`
	Aggs     json.RawMessage `json:"aggregations"`
	Raw      json.RawMessage `json:"-"`
}

// SearchHits 是搜索响应中的 hits 部分
type SearchHits struct {
	Total    SearchTotal       `json:"total"`
	MaxScore *float64          `json:"max_score"`
	Hits     []json.RawMessage `json:"hits"`
}

// SearchTotal 是命中总数，ES 7+ 返回对象形式
type SearchTotal struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"`
}

// RawHits 返回原始 hits 数组
func (h SearchHits) RawHits() []json.RawMessage {
	return h.Hits
}

// Sources 把每条命中的 _source 抽出来，组成 JSON 数组
func (h SearchHits) Sources() ([]json.RawMessage, error) {
	sources := make([]json.RawMessage, 0, len(h.Hits))
	for _, raw := range h.Hits {
		var hit struct {
			Source json.RawMessage `json:"_source"`
		}
		if err := json.Unmarshal(raw, &hit); err != nil {
			return nil, fmt.Errorf("decode hit: %w", err)
		}
		if len(hit.Source) == 0 {
			continue
		}
		sources = append(sources, hit.Source)
	}
	return sources, nil
}

// UnmarshalHits 把命中结果的 _source 反序列化为指定类型切片。
// target 需为切片指针，例如 *[]Article。
func (r *SearchResult) UnmarshalHits(target any) error {
	if r == nil {
		return errors.New("nil search result")
	}

	sources, err := r.Hits.Sources()
	if err != nil {
		return err
	}

	buf, err := json.Marshal(sources)
	if err != nil {
		return fmt.Errorf("marshal sources: %w", err)
	}

	if err := json.Unmarshal(buf, target); err != nil {
		return fmt.Errorf("decode search hits: %w", err)
	}
	return nil
}

// Source 返回第 idx 条命中的 _source 原始 JSON
func (r *SearchResult) Source(idx int) (json.RawMessage, bool) {
	if r == nil || idx < 0 || idx >= len(r.Hits.Hits) {
		return nil, false
	}

	var hit struct {
		Source json.RawMessage `json:"_source"`
	}
	if err := json.Unmarshal(r.Hits.Hits[idx], &hit); err != nil {
		return nil, false
	}
	return hit.Source, true
}

// Highlights 返回第 idx 条命中的 highlight 结果
func (r *SearchResult) Highlights(idx int) (map[string][]string, bool) {
	if r == nil || idx < 0 || idx >= len(r.Hits.Hits) {
		return nil, false
	}

	var hit struct {
		Highlight map[string][]string `json:"highlight"`
	}
	if err := json.Unmarshal(r.Hits.Hits[idx], &hit); err != nil {
		return nil, false
	}
	return hit.Highlight, true
}

// responseError 把错误响应转换为 error，索引不存在时返回 ErrIndexNotFound
func responseError(res *esapi.Response) error {
	body, _ := io.ReadAll(res.Body)

	var payload struct {
		Error struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error.Type != "" {
		if payload.Error.Type == "index_not_found_exception" {
			return fmt.Errorf("%w: %s", ErrIndexNotFound, payload.Error.Reason)
		}
		return fmt.Errorf("elasticsearch error [%d]: %s: %s", res.StatusCode, payload.Error.Type, payload.Error.Reason)
	}

	return fmt.Errorf("elasticsearch error [%d]: %s", res.StatusCode, res.Status())
}

// checkResponse 统一处理非 2xx 响应
func checkResponse(res *esapi.Response) error {
	if res.IsError() {
		return responseError(res)
	}
	return nil
}

// marshalBody 把任意对象序列化为请求体
func marshalBody(v any) (io.Reader, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	return bytes.NewReader(buf), nil
}

// toBodyReader 把 Query、map 或 Builder 转换为请求体
func toBodyReader(query any) (io.Reader, error) {
	switch q := query.(type) {
	case nil:
		return nil, nil
	case Query:
		return q.Reader()
	case *Builder:
		return q.Build().Reader()
	case io.Reader:
		return q, nil
	default:
		return marshalBody(q)
	}
}

// decodeSource 从 Get 响应中取出 _source 并反序列化
func decodeSource(body io.Reader, result any) error {
	var payload struct {
		Source json.RawMessage `json:"_source"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(payload.Source) == 0 {
		return ErrDocumentNotFound
	}

	if err := json.Unmarshal(payload.Source, result); err != nil {
		return fmt.Errorf("decode _source: %w", err)
	}
	return nil
}

// normalizeDocuments 把文档集合统一成 []map[string]any
func normalizeDocuments(documents any) ([]map[string]any, error) {
	raw, err := json.Marshal(documents)
	if err != nil {
		return nil, fmt.Errorf("marshal documents: %w", err)
	}

	var docs []map[string]any
	if err := json.Unmarshal(raw, &docs); err != nil {
		return nil, fmt.Errorf("documents must be a slice: %w", err)
	}
	return docs, nil
}

// fieldValue 从文档 map 中按字段名取值，兼容 id/ID 等常见写法
func fieldValue(doc map[string]any, field string) (any, bool) {
	if v, ok := doc[field]; ok {
		return v, true
	}

	// 大小写不敏感兜底匹配
	for key, v := range doc {
		if strings.EqualFold(key, field) {
			return v, true
		}
	}
	return nil, false
}
