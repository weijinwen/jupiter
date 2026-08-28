package meilisearch

import (
	"encoding/json"

	"github.com/meilisearch/meilisearch-go"
)

// Client 是 Meilisearch 客户端的封装，内嵌 meilisearch.ServiceManager，
// 在其基础上提供按 indexUID 路由的便捷方法。
type Client struct {
	meilisearch.ServiceManager
}

// Index 返回指定 uid 的索引管理器，所有文档与搜索操作均通过它进行。
func (c *Client) Index(uid string) meilisearch.IndexManager {
	return c.ServiceManager.Index(uid)
}

// Search 在指定索引上执行搜索，request.Query 作为查询语句，其余参数沿用 SearchRequest。
func (c *Client) Search(indexUID string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	return c.Index(indexUID).Search(request.Query, request)
}

// FacetSearchRaw 在指定索引上执行分面搜索，返回原始 JSON（json.RawMessage），
// 适用于需要自定义解析的场景。
func (c *Client) FacetSearchRaw(indexUID string, request *meilisearch.FacetSearchRequest) (*json.RawMessage, error) {
	return c.Index(indexUID).FacetSearch(request)
}

// FacetSearch 在指定索引上执行分面搜索，并将 FacetSearchRaw 返回的原始 JSON
// 反序列化为 FacetSearchResponse。若只需原始字节，请使用 FacetSearchRaw。
func (c *Client) FacetSearch(indexUID string, request *meilisearch.FacetSearchRequest) (*meilisearch.FacetSearchResponse, error) {
	raw, err := c.FacetSearchRaw(indexUID, request)
	if err != nil {
		return nil, err
	}

	var response meilisearch.FacetSearchResponse
	if err := json.Unmarshal(*raw, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetDocuments 批量获取指定索引的文档，查询参数通过 request 传入，结果写入 result。
func (c *Client) GetDocuments(indexUID string, request *meilisearch.DocumentsQuery, result *meilisearch.DocumentsResult) error {
	return c.Index(indexUID).GetDocuments(request, result)
}

// GetDocument 按 identifier 获取指定索引中的单个文档，查询参数通过 request 传入，结果反序列化到 result。
func (c *Client) GetDocument(indexUID string, identifier string, request *meilisearch.DocumentQuery, result any) error {
	return c.Index(indexUID).GetDocument(identifier, request, result)
}

// AddDocuments 向指定索引添加文档（主键已存在则整条覆盖）。
// primaryKey 用于覆盖索引默认主键，传空字符串则使用索引配置。
func (c *Client) AddDocuments(indexUID string, documents any, primaryKey string) (*meilisearch.TaskInfo, error) {
	var opts *meilisearch.DocumentOptions
	if primaryKey != "" {
		opts = &meilisearch.DocumentOptions{PrimaryKey: &primaryKey}
	}
	return c.Index(indexUID).AddDocuments(documents, opts)
}

// UpdateDocuments 更新指定索引中的文档（仅更新传入字段，不删除已有字段）。
// primaryKey 用于覆盖索引默认主键，传空字符串则使用索引配置。
func (c *Client) UpdateDocuments(indexUID string, documents any, primaryKey string) (*meilisearch.TaskInfo, error) {
	var opts *meilisearch.DocumentOptions
	if primaryKey != "" {
		opts = &meilisearch.DocumentOptions{PrimaryKey: &primaryKey}
	}
	return c.Index(indexUID).UpdateDocuments(documents, opts)
}

// DeleteDocument 按 identifier 删除指定索引中的单个文档。
// 注意：Meilisearch 的删除端点不接受 primaryKey（主键为索引级配置），
// 因此这里与 AddDocuments/UpdateDocuments 不同，不再暴露该参数。
func (c *Client) DeleteDocument(indexUID string, identifier string) (*meilisearch.TaskInfo, error) {
	return c.Index(indexUID).DeleteDocument(identifier, nil)
}

// DeleteDocuments 按 identifiers 批量删除指定索引中的文档。
// 注意：Meilisearch 的批量删除端点不接受 primaryKey，传入会导致 400 错误
// (Unknown parameter `primaryKey`: expected one of `customMetadata`)。
func (c *Client) DeleteDocuments(indexUID string, identifiers []string) (*meilisearch.TaskInfo, error) {
	return c.Index(indexUID).DeleteDocuments(identifiers, nil)
}

// DeleteAllDocuments 删除指定索引中的全部文档。
func (c *Client) DeleteAllDocuments(indexUID string) (*meilisearch.TaskInfo, error) {
	return c.Index(indexUID).DeleteAllDocuments(nil)
}
