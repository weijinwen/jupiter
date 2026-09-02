# Elasticsearch 存储组件

Elasticsearch 是一个分布式全文搜索与分析引擎。本组件提供 Jupiter 框架与 Elasticsearch 的集成，包括客户端生命周期管理、常用操作封装、可观测性接入等功能。

## 功能特性

- 客户端单例管理，同名配置只创建一次
- 连接池参数完整暴露，支持精细化调优
- 重试由官方客户端内置提供，框架不叠加额外层
- 常用文档操作封装（索引、更新、删除、查询），消除样板代码
- 泛型搜索方法 `SearchInto[T]`，自动反序列化命中列表
- Prometheus 指标自动采集（请求数、延迟分布）
- 访问日志与慢查询日志
- Governor 调试端点 `/debug/elasticsearch/stats` 实时查看集群健康

## 配置说明

### 基本配置

配置路径约定为 `jupiter.elasticsearch.<name>`，通过 `StdConfig(name)` 加载。

```toml
[jupiter.elasticsearch.default]
# ES 节点地址列表
addresses = ["http://localhost:9200"]
# Basic Auth（可选）
username = ""
password = ""
# API Key 认证（设置后覆盖 Basic Auth，可选）
apiKey = ""
# Elastic Cloud 部署 ID（可选）
cloudID = ""
# SHA256 证书指纹，用于 HTTPS 校验（可选）
certificateFingerprint = ""
```

### 重试配置

重试由官方客户端内置，框架直接透传，不叠加额外重试层。

```toml
[jupiter.elasticsearch.default]
# 最大重试次数，默认 3
maxRetries = 3
# 关闭重试，默认 false
disableRetry = false
# 触发重试的 HTTP 状态码，默认 [502, 503, 504]
retryOnStatus = [502, 503, 504]
```

### 连接池配置

```toml
[jupiter.elasticsearch.default]
# 建立连接超时，默认 3s
dialTimeout = "3s"
# TCP keepalive 间隔，默认 30s
keepAlive = "30s"
# 最大空闲连接数，默认 100
maxIdleConns = 100
# 每个 host 的最大空闲连接数，默认 100
maxIdleConnsPerHost = 100
# 每个 host 的最大连接数，默认 100
maxConnsPerHost = 100
# 空闲连接存活时间，默认 90s
idleConnTimeout = "90s"
# 等待响应头超时，默认 30s
responseHeaderTimeout = "30s"
# 关闭长连接，默认 false
disableKeepAlives = false
```

### 可观测性配置

```toml
[jupiter.elasticsearch.default]
# 调试模式，打印每次请求和响应体，默认 false
debug = false
# 指标采集开关，默认 true
enableMetric = true
# 访问日志开关（成功请求），默认 false
enableAccessLog = false
# 慢查询阈值，0 表示关闭，默认 1s
slowThreshold = "1s"
# 启动时连通性检查失败的处理方式："panic" | "error"，默认 "error"
onDialError = "error"
```

### 完整配置示例

```toml
[jupiter.elasticsearch.default]
addresses = ["http://localhost:9200"]
username = ""
password = ""
maxRetries = 3
disableRetry = false
dialTimeout = "3s"
keepAlive = "30s"
maxIdleConns = 100
maxIdleConnsPerHost = 100
maxConnsPerHost = 100
idleConnTimeout = "90s"
responseHeaderTimeout = "30s"
debug = false
enableMetric = true
enableAccessLog = false
slowThreshold = "1s"
onDialError = "error"
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/weijinwen/jupiter/pkg/store/elasticsearch"
)

type Book struct {
    Title string `json:"title"`
    Pages int    `json:"pages"`
}

func main() {
    // 从标准配置路径加载并创建单例客户端
    client := elasticsearch.NewElasticsearch("default")

    ctx := context.Background()

    // 索引文档
    err := client.IndexDocument(ctx, "books", "1", Book{Title: "Go 语言", Pages: 380})
    if err != nil {
        panic(err)
    }

    // 读取文档
    var book Book
    err = client.GetSource(ctx, "books", "1", &book)
    if err != nil {
        panic(err)
    }
    fmt.Println(book.Title) // Go 语言

    // 搜索
    books, err := client.SearchInto[Book](ctx, "books", nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("共 %d 本\n", len(books))
}
```

## 客户端操作

### 初始化

```go
// 方式一：单例（推荐），同名只创建一次
client := elasticsearch.NewElasticsearch("default")

// 方式二：手动构建单例，可处理错误
client, err := elasticsearch.StdConfig("default").Singleton()

// 方式三：每次新建客户端
client, err := elasticsearch.StdConfig("default").Build()
```

### 索引文档

```go
type Book struct {
    Title string `json:"title"`
    Pages int    `json:"pages"`
}

// 文档已存在则整条覆盖
err := client.IndexDocument(ctx, "books", "1", Book{Title: "Go 语言", Pages: 380})
```

### 部分更新文档

```go
// 只更新传入的字段，不删除已有字段
err := client.UpdateDocument(ctx, "books", "1", map[string]any{"pages": 400})
```

### 删除文档

```go
err := client.DeleteDocument(ctx, "books", "1")
```

### 读取文档

```go
var book Book
err := client.GetSource(ctx, "books", "1", &book)
if errors.Is(err, elasticsearch.ErrDocumentNotFound) {
    // 文档不存在
}
```

### 搜索

```go
import "github.com/elastic/go-elasticsearch/v8/typedapi/types"

// 简单搜索，自动反序列化命中列表
books, err := client.SearchInto[Book](ctx, "books", nil)

// 带查询条件
query := &types.Query{
    Match: map[string]types.MatchQuery{
        "title": {Query: "Go"},
    },
}
books, err = client.SearchInto[Book](ctx, "books", query)
```

需要聚合、总命中数、分页游标等完整响应时，直接使用官方 API：

```go
resp, err := client.Search().
    Index("books").
    Query(query).
    From(0).Size(10).
    Do(ctx)

fmt.Println(resp.Hits.Total.Value) // 总命中数
for _, hit := range resp.Hits.Hits {
    // 自行处理 hit.Source_
}
```

### 批量写入

批量写入使用官方 `esutil.BulkIndexer`，直接访问 `client.TypedClient` 即可：

```go
import "github.com/elastic/go-elasticsearch/v8/esutil"

bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
    Client: client.TypedClient,
    Index:  "books",
    NumWorkers: 4,
    OnError: func(ctx context.Context, err error) {
        log.Println("bulk error:", err)
    },
})
if err != nil {
    panic(err)
}

err = bi.Add(ctx, esutil.BulkIndexerItem{
    Action:     "index",
    DocumentID: "1",
    Body:       strings.NewReader(`{"title":"Go 语言","pages":380}`),
})

_ = bi.Close(ctx)
```

### 索引管理

索引管理通过 `client.Indices` 访问：

```go
// 创建索引
_, err := client.Indices.Create("books").Do(ctx)

// 删除索引
_, err = client.Indices.Delete("books").Do(ctx)

// 查询索引是否存在
exists, err := client.Indices.Exists("books").Do(ctx)
```

## 可观测性

### 指标采集

启用 `enableMetric = true` 后，每次请求自动上报以下 Prometheus 指标：

| 指标 | 类型 | 说明 |
|------|------|------|
| `elasticsearch_client_requests_total` | Counter | 请求总数，按 `name`、`method`、`addr`、`code` 分组 |
| `elasticsearch_client_request_duration_seconds` | Histogram | 请求耗时分布，按 `name`、`method`、`addr` 分组 |

### 访问日志

`enableAccessLog = true` 时，每次成功请求会输出一条 `INFO` 级别日志，包含：

```
INFO  access  {"key":"default","meth":"POST","addr":"localhost:9200","path":"/_search","code":200,"cost":"0.012"}
```

网络错误或 HTTP 5xx 响应无论开关状态始终输出 `ERROR` 日志。

### 慢查询日志

`slowThreshold` 大于 0 时，超过阈值的请求输出 `WARN` 日志：

```
WARN  slow  {"key":"default","meth":"POST","addr":"localhost:9200","path":"/_search","code":200,"cost":"1.234"}
```

## 调试端点

包初始化时自动向 Governor 注册 `/debug/elasticsearch/stats` 端点，实时查询各实例的集群健康状态：

```bash
curl http://localhost:9091/debug/elasticsearch/stats
```

响应示例：

```json
{
  "default": {
    "cluster_name": "my-cluster",
    "status": "green",
    "number_of_nodes": 3,
    "active_shards": 20,
    "active_shards_percent": 100
  }
}
```

## 错误处理

```go
var (
    // ErrEmptyAddresses 未配置任何节点地址
    ErrEmptyAddresses = errors.New("elasticsearch addresses is empty")

    // ErrConnectFailed 启动时连通性检查失败
    ErrConnectFailed = errors.New("elasticsearch connect failed")

    // ErrSlowQuery 请求耗时超过慢查询阈值（出现在慢查询日志中）
    ErrSlowQuery = errors.New("elasticsearch slow query")

    // ErrDocumentNotFound 文档不存在
    ErrDocumentNotFound = errors.New("elasticsearch document not found")
)
```

```go
err := client.GetSource(ctx, "books", "missing-id", &book)
if errors.Is(err, elasticsearch.ErrDocumentNotFound) {
    // 文档不存在，按业务逻辑处理
}
```

## 注意事项

1. `IndexDocument` 会整条覆盖已有文档；只修改部分字段请使用 `UpdateDocument`
2. `SearchInto[T]` 只返回文档列表，需要总命中数、聚合结果时请直接调用 `client.Search()`
3. 批量写入请使用官方 `esutil.BulkIndexer`，通过 `client.TypedClient` 传入客户端
4. `onDialError = "panic"` 时，启动期无法连接 ES 会导致进程退出，生产环境建议保持默认值 `"error"` 并在应用层做好降级
5. `debug = true` 会打印完整请求和响应体，生产环境请关闭

## 下一步

- 了解 [存储组件](./stores.md)
- 探索 [缓存组件](./caches.md)
- 查看 [MeiliSearch 组件](./meilisearch.md)
- 学习 [日志组件](./xlog.md)
