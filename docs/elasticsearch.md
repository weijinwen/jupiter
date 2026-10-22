# Elasticsearch 存储组件

本组件提供 Jupiter 框架与 [Elasticsearch](https://www.elastic.co/elasticsearch/) 的集成，基于官方客户端 [go-elasticsearch/v8](https://github.com/elastic/go-elasticsearch)，包含客户端管理、索引级便捷方法、查询 DSL 构建器，以及指标采集、慢查询、访问日志与重试等能力。

## 功能特性

- 多节点地址与连接池管理
- 支持 Basic Auth、API Key、自定义 CA 证书
- 索引级便捷方法（增删改查、批量、搜索、计数）
- 查询 DSL 构建器（bool / term / range / terms / geo / 分页 / 排序 / 高亮 / 聚合）
- 结构体过滤条件（`filter` tag），与 MeiliSearch 组件保持一致的语义
- 内置重试（指数退避 + 抖动），请求体可重放
- 自动指标采集、慢查询监控、访问日志

## 配置说明

配置键为 `jupiter.es.{name}`，注意必须带 `jupiter.` 前缀（由 `constant.ConfigKey` 拼接），否则会静默回落默认值。

### 基本配置

```toml
[jupiter.es.default]
# 目标服务地址，支持多节点
addresses = ["http://127.0.0.1:9200"]
# 单地址简写，addresses 为空时生效
# addr = "http://127.0.0.1:9200"

# Basic Auth
username = ""
password = ""
# API Key，与 username/password 互斥且优先级更高
apiKey = ""

# 自定义 CA 证书（PEM 文件路径），连 HTTPS 自签集群时使用
caCertFile = ""

# 请求超时时间
timeout = "3000ms"
# 保持连接时间
keepAlive = "30s"
# 最大空闲连接数
maxIdleConns = 100
# 每个主机的最大空闲连接数
maxIdleConnsPerHost = 100
# 空闲连接超时时间
idleConnTimeout = "30s"
# 是否禁用长连接
disableKeepAlives = false
# 每个主机的最大连接数
maxConnsPerHost = 100
# 响应头超时时间
responseHeaderTimeout = "30s"

# 是否启用调试模式
debug = false
# 是否启用指标收集
enableMetric = true
# 是否启用访问日志
enableAccessLog = false
# 慢查询阈值，0 表示不记录
slowThreshold = "1s"

# 创建连接失败时的处理方式："panic" 或 "error"
onDialError = "panic"
# 重试次数
retry = 2
# 重试等待时间
retryWaitTime = "200ms"
```

### 代码中使用配置

```go
// 读取配置并创建客户端（单例）
client := elasticsearch.StdConfig("default").MustSingleton()

// 非单例方式
client, err := elasticsearch.StdConfig("default").Build()

// 直接指定配置键
config := elasticsearch.RawConfig("es.custom")
```

### 重试机制说明

- `retry`：重试次数，默认 2，表示首次尝试 + 2 次重试。设为 0 表示不重试。
- `retryWaitTime`：初始等待时间，默认 200ms，采用指数退避并在每次等待时加入抖动。

重试触发条件：

1. **网络错误**：连接失败、超时等
2. **服务器错误**：HTTP 状态码 ≥ 500

> 注意：官方传输层的自动重试已关闭（`DisableRetry = true`），重试统一由本组件的 `RoundTripper` 处理，避免双层重试放大流量。重试时会自动缓存并重放请求体。

### 错误状态收敛

为控制 Prometheus 指标基数，错误信息会按前缀收敛到有限标签：

- **elasticsearch**：ES 返回的错误
- **request**：HTTP 请求相关错误
- **search**：搜索操作相关错误
- **server**：服务器错误
- **Unknown**：未知错误类型

## 快速开始

```go
package main

import (
    "fmt"

    "github.com/weijinwen/jupiter/pkg/store/elasticsearch"
)

func main() {
    client := elasticsearch.StdConfig("default").MustSingleton()
    index := client.Index("articles")

    // 添加文档
    res, err := index.AddDocument("1", map[string]any{
        "title":  "Jupiter 微服务框架",
        "status": "published",
    })
    if err != nil {
        panic(err)
    }
    defer res.Body.Close()

    // 搜索
    var articles []struct {
        Title string `json:"title"`
    }
    result, err := index.Search(
        elasticsearch.NewBuilder().
            SetQuery("Jupiter", "title").
            AddFilter("status", "=", "published").
            SetFrom(0).SetSize(10),
        &articles,
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("total:", result.Hits.Total.Value)
    for _, a := range articles {
        fmt.Println(a.Title)
    }
}
```

## 客户端 API

### Client

| 方法 | 说明 |
|---|---|
| `Index(name)` | 返回绑定索引名的 `*Index` 管理器 |
| `Ping()` | 检查集群可用性 |
| `Refresh(indices...)` | 刷新索引，使写入对搜索可见 |
| `Config()` | 返回客户端使用的配置 |

### Index

| 方法 | 说明 |
|---|---|
| `Name()` | 索引名 |
| `Exists()` | 判断索引是否存在 |
| `Create(mapping)` | 创建索引，`mapping` 为索引设置与映射，传 `nil` 用默认配置 |
| `Delete()` | 删除索引 |
| `AddDocument(id, doc)` | 添加/整条覆盖单个文档，`id` 为空则自动生成 |
| `AddDocuments(docs, idField)` | 批量添加（bulk），`idField` 为作为 `_id` 的字段名 |
| `UpdateDocument(id, doc)` | 按 id 局部更新 |
| `UpdateDocuments(docs, idField)` | 批量局部更新（bulk update） |
| `GetDocument(id, &out)` | 按 id 获取文档，反序列化到 `out` |
| `DocumentExists(id)` | 判断文档是否存在 |
| `DeleteDocument(id)` | 按 id 删除文档 |
| `DeleteByQuery(query)` | 按条件批量删除 |
| `DeleteAllDocuments()` | 删除索引内全部文档（保留索引） |
| `Search(query, &out)` | 执行搜索，`query` 可为 `Query`、`map` 或 `*Builder` |
| `Count(query)` | 统计满足条件的文档数，`query` 传 `nil` 统计全部 |

### 搜索结果

`Search` 返回 `*SearchResult`：

```go
result.Took              // 耗时（毫秒）
result.TimedOut          // 是否超时
result.Hits.Total.Value  // 命中总数
result.Hits.Total.Relation // "eq" 或 "gte"
result.Aggs              // 聚合结果原始 JSON
result.Raw               // 完整响应原始 JSON

result.UnmarshalHits(&out)  // 把 _source 反序列化为切片
result.Source(i)            // 第 i 条命中的 _source
result.Highlights(i)        // 第 i 条命中的 highlight
```

### 批量结果

`BulkResult` 提供 `Errors`（是否存在失败条目）与 `FailedCount()`：

```go
result, err := index.AddDocuments(docs, "id")
if err != nil {
    panic(err)
}
if result.Errors {
    fmt.Println("失败条目数:", result.FailedCount())
}
```

## 查询构建器

```go
query := elasticsearch.NewBuilder().
    // 全文检索：不传字段走 query_string，传字段走 multi_match
    SetQuery("jupiter", "title", "content").
    // 过滤条件
    AddFilter("status", "=", "published").
    AddFilter("views", ">=", 100).
    AddFilter("tags", "IN", []string{"go", "es"}).
    AddFilter("deleted_at", "NOT EXISTS", nil).
    // 日期范围
    AddDateRangeFilter(elasticsearch.DateRangeFilter{
        Field: "created_at",
        Start: time.Now().AddDate(0, 0, -7),
        End:   time.Now(),
    }).
    // 地理位置
    AddGeoFilter(elasticsearch.GeoFilter{
        Field:  "location",
        Point:  elasticsearch.GeoPoint{Lat: 31.2, Lng: 121.5},
        Radius: 5000,
    }).
    // 排序、返回字段、高亮、分页
    AddSort("created_at", false).
    SetSource("title", "created_at").
    AddHighlight("title").
    HighlightPreTag("<em>").HighlightPostTag("</em>").
    SetFrom(0).SetSize(20).
    // 聚合
    AddTermsAgg("by_status", "status", 10).
    Build()

result, err := index.Search(query, &articles)
```

### 支持的过滤 operator

| operator | 生成的 DSL |
|---|---|
| `=` | `term` |
| `!=` | `must_not term` |
| `>`, `>=`, `<`, `<=` | `range`（`gt`/`gte`/`lt`/`lte`） |
| `IN` | `terms` |
| `NOT IN` | `must_not terms` |
| `EXISTS` | `exists` |
| `NOT EXISTS` | `must_not exists` |
| `CONTAINS` | `match_phrase` |
| `NOT CONTAINS` | `must_not match_phrase` |

非法 operator 会被忽略并记录 warn 日志，不会中断构建。

### 组合条件

```go
// OR 组合进入 should，AND 组合进入 must
builder.AddFilterCondition(elasticsearch.FilterCondition{
    Operator: "OR",
    Filters: []elasticsearch.Filter{
        {Field: "status", Operator: "=", Value: "published"},
        {Field: "status", Operator: "=", Value: "review"},
    },
})

// 分组条件，后续统一应用
builder.AddFilterStructWithGroup("base", &ArticleQuery{Status: "published"})
builder.ApplyFilterGroup("base")
```

### 结构体过滤

与 MeiliSearch 组件一致的 `filter` tag：

```go
type ArticleQuery struct {
    Status    string    `filter:"status"`
    MinViews  int       `filter:"views,op=>="`
    Tags      []string  `filter:"tags,op=IN"`
    CreatedAt time.Time `filter:"created_at,datefmt=2006-01-02"`
    Author    Author    `filter:"author,nested"`
    Internal  string    `filter:"-"`
}

builder := elasticsearch.NewBuilder().AddFilterStruct(&ArticleQuery{Status: "published", MinViews: 100})
```

tag 支持的修饰项：`op=`、`required`、`ignorezero`、`ignorezero=false`、`nested`、`datefmt=`、`group=`、`transform=`、`validate=`、`deps=`。零值字段默认被忽略。

### 自定义校验与转换

```go
builder := elasticsearch.NewBuilder().
    RegisterValidator("status", func(v any) error {
        if v.(string) == "forbidden" {
            return errors.New("status not allowed")
        }
        return nil
    }).
    RegisterTransformer("status", func(v any) any {
        return strings.ToLower(v.(string))
    }).
    AddFilterStruct(&query)
```

## 错误定义

| 错误 | 说明 |
|---|---|
| `ErrIndexNotFound` | 索引不存在（`index_not_found_exception`） |
| `ErrDocumentNotFound` | 文档不存在 |
| `ErrInvalidQuery` | 查询条件非法 |
| `ErrSlowQuery` | 查询过慢（慢日志标记） |
| `ErrMaxRetriesExceeded` | 超过最大重试次数 |

```go
_, err := index.Search(query, &out)
if errors.Is(err, elasticsearch.ErrIndexNotFound) {
    // 索引尚未创建，按需初始化
}
```

## 指标

启用 `enableMetric` 后，通过 `pkg/core/metric` 上报：

- `client_handle_total{type="elasticsearch", name, method, server, code}`
- `client_handle_seconds{type="elasticsearch", name, method, server}`

其中 `name` 为配置名（即 `jupiter.es.{name}` 中的 `{name}`），`server` 为请求路径，`code` 为状态或收敛后的错误标签。

## 注意事项

- **配置前缀**：必须写 `[jupiter.es.default]`，写 `[es.default]` 会静默回落默认值。
- **键名大小写**：多词键使用 camelCase（如 `slowThreshold`、`enableAccessLog`），conf 不处理下划线。
- **批量主键**：`AddDocuments`/`UpdateDocuments` 的 `idField` 是文档字段名，会映射为 bulk 的 `_id`；传空则由 ES 自动生成。
- **文档可见性**：写入后若需立刻可搜索，调用 `client.Refresh("articles")` 或索引配置 `refresh_interval`。
- **ES 版本**：客户端为 v8，可通过兼容模式连接 ES 7（官方 `EnableCompatibilityMode`）。
