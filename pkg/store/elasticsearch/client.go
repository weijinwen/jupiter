package elasticsearch

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// Client 是 Elasticsearch 客户端的封装，内嵌 *elasticsearch.TypedClient，
// 因此官方的全部强类型 API 可直接使用：
//
//	_, err := client.Index("books").Id("1").Document(book).Do(ctx)
//	_, err := client.Delete("books", "1").Do(ctx)
//	_, err := client.Indices.Create("books").Do(ctx)
//
// 本类型只在官方 API 之上补充少量消除样板代码的方法，不重复包装已有能力。
type Client struct {
	*elasticsearch.TypedClient
	config *Config
}

// GetSource 按 id 获取文档，并把 _source 反序列化到 result。
//
// 官方 Get API 返回的 Source_ 是 json.RawMessage，调用方总要自己判断 Found 再反序列化，
// 这里把这段样板收敛掉。文档不存在时返回 ErrDocumentNotFound。
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

// NewBulkIndexer 创建官方的批量写入器，自动注入当前客户端。
//
// 批量写入的并发、按字节/时间自动 flush、逐条成功失败回调等能力，
// 官方 esutil.BulkIndexer 已完整提供，本包只做接线，不重复实现。
// 除 Client 外的其余参数（NumWorkers、FlushBytes、OnError 等）由调用方按需设置。
//
//	bi, err := client.NewBulkIndexer(esutil.BulkIndexerConfig{Index: "books"})
//	if err != nil {
//	    return err
//	}
//	err = bi.Add(ctx, esutil.BulkIndexerItem{Action: "index", Body: body})
//	err = bi.Close(ctx)
func (c *Client) NewBulkIndexer(cfg esutil.BulkIndexerConfig) (esutil.BulkIndexer, error) {
	cfg.Client = c.TypedClient
	return esutil.NewBulkIndexer(cfg)
}

// SearchInto 执行搜索并把命中的 _source 反序列化为 []T。
//
// 因为 Go 的方法不支持类型参数，所以它是包级函数而非 Client 的方法。
// 需要聚合、总命中数等完整响应时，请直接用 client.Search()。
//
//	books, err := elasticsearch.SearchInto[Book](ctx, client, "books", query)
func SearchInto[T any](ctx context.Context, c *Client, index string, query *types.Query) ([]T, error) {
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
