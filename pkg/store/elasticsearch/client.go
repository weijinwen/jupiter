package elasticsearch

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// Client 是 Elasticsearch 客户端的封装，内嵌 *elasticsearch.TypedClient。
// 官方全部强类型 API 可直接使用（client.Indices.Create、client.Search 等），
// 本类型在其上只补充消除常见样板的便捷方法。
type Client struct {
	*elasticsearch.TypedClient
	config *Config
}

// IndexDocument 按 id 索引文档，文档已存在则整条覆盖。
//
// 等价于官方 client.Index(index).Id(id).Document(doc).Do(ctx)，
// 只需要知道操作是否成功时用此方法，需要返回的 _version、result 等字段时请直接调官方 API。
func (c *Client) IndexDocument(ctx context.Context, index, id string, doc any) error {
	_, err := c.TypedClient.Index(index).Id(id).Document(doc).Do(ctx)
	return err
}

// UpdateDocument 按 id 部分更新文档，仅更新 doc 中传入的字段，不删除已有字段。
//
// 对应 ES 的 Update API（doc 脚本），与 IndexDocument 的整条覆盖不同。
func (c *Client) UpdateDocument(ctx context.Context, index, id string, doc any) error {
	_, err := c.TypedClient.Update(index, id).Doc(doc).Do(ctx)
	return err
}

// DeleteDocument 按 id 删除文档。
//
// 文档不存在时 ES 返回 404，go-elasticsearch 会将其作为错误透传给调用方。
func (c *Client) DeleteDocument(ctx context.Context, index, id string) error {
	_, err := c.TypedClient.Delete(index, id).Do(ctx)
	return err
}

// GetSource 按 id 获取文档，并把 _source 反序列化到 result。
// 文档不存在时返回 ErrDocumentNotFound。
func (c *Client) GetSource(ctx context.Context, index, id string, result any) error {
	resp, err := c.Get(index, id).Do(ctx)
	if err != nil {
		return err
	}
	if !resp.Found || len(resp.Source_) == 0 {
		return ErrDocumentNotFound
	}
	return json.Unmarshal(resp.Source_, result)
}

// SearchInto 执行搜索并把命中的 _source 反序列化为 []T。
//
// 需要聚合、总命中数等完整响应时，请直接用 client.Search()。
//
//	books, err := client.SearchInto[Book](ctx, "books", query)
func (c *Client) SearchInto[T any](ctx context.Context, index string, query *types.Query) ([]T, error) {
	resp, err := c.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]T, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		if len(hit.Source_) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(hit.Source_, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}
