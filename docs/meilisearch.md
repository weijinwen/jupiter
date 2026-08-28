# MeiliSearch 存储组件

MeiliSearch 是一个轻量级、快速、开源的搜索引擎。本组件提供了 Jupiter 框架与 MeiliSearch 的集成，包括客户端管理、查询构建器等功能。

## 功能特性

- 客户端连接池管理
- 查询构建器，支持复杂的搜索条件
- 自动指标采集和日志记录
- 慢查询监控
- 支持结构体过滤条件
- 支持高亮、排序、分面等高级搜索功能
- 内置重试机制，提高系统可靠性
- 错误状态收敛，减少 Prometheus 指标基数

## 配置说明

### 基本配置

MeiliSearch 组件的基本配置项如下：

```toml
[search.default]
# MeiliSearch 服务器地址
addr = "http://localhost:7700"
# API 密钥
apiKey = "your-api-key"
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
# 创建连接失败时的处理方式："panic" 或 "error"
onDialError = "panic"
# 重试次数
retry = 2
# 重试等待时间
retryWaitTime = "200ms"
# 慢查询阈值，0 表示不记录
slowThreshold = "1s"
```

### 功能开关

MeiliSearch 组件提供了多个功能开关，可以根据需要启用或禁用特定功能：

```toml
[search.default]
# 是否启用调试模式
debug = false
# 是否启用指标收集
enableMetric = true
# 是否启用访问日志
enableAccessLog = false
```

### 完整配置示例

```toml
[search.default]
addr = "http://localhost:7700"
apiKey = "your-api-key"
timeout = "3000ms"
keepAlive = "30s"
maxIdleConns = 100
maxIdleConnsPerHost = 100
idleConnTimeout = "30s"
disableKeepAlives = false
maxConnsPerHost = 100
responseHeaderTimeout = "30s"
onDialError = "panic"
retry = 2
retryWaitTime = "200ms"
slowThreshold = "1s"
debug = false
enableMetric = true
enableAccessLog = false
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    client, err := meilisearch.StdConfig("default").Singleton()
    if err != nil {
        panic(err)
    }
    
    // 在指定索引上执行搜索
    resp, err := client.Search("products", &meilisearch.SearchRequest{
        Query: "laptop",
        Page:  1,
        HitsPerPage: meilisearch.Int64Ptr(20),
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 使用 Builder 进行复杂查询

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    client, err := meilisearch.StdConfig("default").Singleton()
    if err != nil {
        panic(err)
    }
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("laptop").
        AddFilter("price", "<", 1000).
        AddFilter("brand", "IN", []string{"Apple", "Dell"}).
        AddSort("price", true).
        SetPage(1).
        SetHitsPerPage(10)
    
    // 执行搜索
    resp, err := client.Search("products", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

## 客户端操作

### 文档操作

```go
// 添加文档
docs := []map[string]any{
    {"id": 1, "name": "Product 1", "price": 99.99},
    {"id": 2, "name": "Product 2", "price": 149.99},
}
task, err := client.AddDocuments("products", docs, "id")
if err != nil {
    panic(err)
}

// 更新文档
task, err = client.UpdateDocuments("products", docs, "id")
if err != nil {
    panic(err)
}

// 获取单个文档
var result map[string]any
err = client.GetDocument("products", "1", &meilisearch.DocumentQuery{}, &result)
if err != nil {
    panic(err)
}

// 获取多个文档
var results meilisearch.DocumentsResult
err = client.GetDocuments("products", &meilisearch.DocumentsQuery{
    Limit: 10,
}, &results)
if err != nil {
    panic(err)
}

// 删除单个文档
task, err = client.DeleteDocument("products", "1")
if err != nil {
    panic(err)
}

// 删除多个文档
task, err = client.DeleteDocuments("products", []string{"1", "2"})
if err != nil {
    panic(err)
}

// 删除所有文档
task, err = client.DeleteAllDocuments("products")
if err != nil {
    panic(err)
}
```

## 查询构建器

### 基本过滤

```go
builder := meilisearch.NewBuilder().
    SetQuery("search term").
    AddFilter("price", ">=", 100).
    AddFilter("category", "=", "electronics").
    SetPage(1).
    SetHitsPerPage(20)
```

### 使用结构体过滤

```go
type ProductFilter struct {
    Name     string `filter:"name,op=CONTAINS"`
    Price    int    `filter:"price,op=>="`
    Category string `filter:"category,op=IN"`
}

filter := ProductFilter{
    Name:     "laptop",
    Price:    1000,
    Category: "electronics",
}

builder := meilisearch.NewBuilder().
    SetQuery("").
    AddFilterStruct(filter)
```

### 组合过滤条件

```go
// AND 条件
andCondition := meilisearch.FilterCondition{
    Filters: []meilisearch.Filter{
        {Field: "price", Operator: ">=", Value: 100},
        {Field: "price", Operator: "<=", Value: 500},
    },
    Operator: "AND",
}

// OR 条件
orCondition := meilisearch.FilterCondition{
    Filters: []meilisearch.Filter{
        {Field: "category", Operator: "=", Value: "electronics"},
        {Field: "category", Operator: "=", Value: "computers"},
    },
    Operator: "OR",
}

builder := meilisearch.NewBuilder().
    AddFilterCondition(andCondition).
    AddFilterCondition(orCondition)
```

### 地理位置过滤

```go
geoFilter := meilisearch.GeoFilter{
    Field:  "location",
    Point:  meilisearch.GeoPoint{Lat: 40.7128, Lng: -74.0060},
    Radius: 10000, // 10公里
    Unit:   "m",
}

builder := meilisearch.NewBuilder().
    AddGeoFilter(geoFilter)
```

### 日期范围过滤

```go
dateFilter := meilisearch.DateRangeFilter{
    Field: "created_at",
    Start: time.Now().AddDate(0, -1, 0), // 一个月前
    End:   time.Now(),
}

builder := meilisearch.NewBuilder().
    AddDateRangeFilter(dateFilter)
```

### 数组过滤

```go
arrayFilter := meilisearch.ArrayFilter{
    Field:    "tags",
    Values:   []any{"golang", "rust", "python"},
    Operator: "IN",
}

builder := meilisearch.NewBuilder().
    AddArrayFilter(arrayFilter)
```

### 排序和高亮

```go
builder := meilisearch.NewBuilder().
    SetQuery("search term").
    AddSort("price", true).        // 按价格升序
    AddSort("created_at", false).  // 按创建时间降序
    AddHighlight("title", "description").
    HighlightPreTag("<em>").
    HighlightPostTag("</em>")
```

### 分面搜索

```go
builder := meilisearch.NewBuilder().
    SetQuery("search term").
    AddFacets("category", "brand", "price")
```

### 普通搜索分面统计

普通搜索分面统计用于在搜索文档的同时返回 facet 分布，例如品牌、分类或标签计数。

```go
request := meilisearch.NewBuilder().
    SetQuery("phone").
    SetFacets("brand", "category").
    Build()

resp, err := client.Search("products", &request)
if err != nil {
    return err
}

fmt.Printf("facet distribution: %s\n", resp.FacetDistribution)
```

`SetFacets` 会覆盖已有分面字段；如果需要追加字段，继续使用 `AddFacets`。

### FacetSearch 分面值自动补全

FacetSearch 用于在某个 facet 字段的值集合中做 autocomplete。字段需要先在 MeiliSearch 中配置为可过滤/可分面字段，例如 `filterableAttributes` 包含 `brand`。

```go
request := meilisearch.NewFacetSearchBuilder().
    SetFacetName("brand").
    SetFacetQuery("app").
    SetQuery("phone").
    AddFilter("status", "=", "online").
    Build()

resp, err := client.FacetSearch("products", &request)
if err != nil {
    return err
}

for _, hit := range resp.FacetHits {
    fmt.Printf("%v\n", hit)
}
```

如果需要 SDK 原始 JSON，可以使用：

```go
raw, err := client.FacetSearchRaw("products", &request)
```

## 重试机制

MeiliSearch 组件内置了重试机制，可以在网络请求失败或服务器返回错误时自动重试请求，提高系统的可靠性。

### 重试配置

```toml
[search.default]
retry = 3                    # 最多重试 3 次
retryWaitTime = "100ms"      # 初始等待时间 100ms
```

### 重试触发条件

重试机制会在以下情况下触发：

1. **HTTP 请求失败**：网络错误、连接超时等
2. **服务器错误**：HTTP 状态码为 5xx 的服务器内部错误
3. **搜索操作失败**：搜索请求执行失败

### 重试策略

- 采用指数退避算法，每次重试等待时间会翻倍
- 添加随机抖动，避免多个请求同时重试造成的雷群效应
- 重试次数耗尽后，返回最后一次的错误

```go
// 使用带重试的客户端
client, err := meilisearch.StdConfig("default").Singleton()
if err != nil {
    panic(err)
}

// 执行搜索，失败时会自动重试
resp, err := client.Search("products", &meilisearch.SearchRequest{
    Query: "laptop",
})
if err != nil {
    // 重试次数耗尽后的错误处理
    if errors.Is(err, meilisearch.ErrMaxRetriesExceeded) {
        // 处理超过最大重试次数的错误
    }
}
```

## 错误状态收敛

为了减少 Prometheus 指标的基数，MeiliSearch 组件会对错误信息进行收敛处理。系统会根据错误信息的前缀将其归类到有限的状态类别中：

- **meilisearch**：MeiliSearch 相关错误
- **request**：HTTP 请求相关错误
- **search**：搜索操作相关错误
- **server**：服务器相关错误
- **Unknown**：未知错误类型

这种收敛机制可以避免因错误信息变化过大而导致 Prometheus 指标标签过多的问题。

## 结构体过滤标签

结构体字段可以使用 `filter` 标签来配置过滤条件：

```go
type ProductFilter struct {
    Name     string    `filter:"name,op=CONTAINS"`
    Price    int       `filter:"price,op=>=,required"`
    Category string    `filter:"category,op=IN"`
    Tags     []string  `filter:"tags,op=CONTAINS"`
    InStock  bool      `filter:"in_stock,op=="`
}
```

### 标签参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `field` | 字段名 | 结构体字段名 |
| `op` | 操作符 | `=` |
| `required` | 是否必填 | `false` |
| `ignorezero` | 是否忽略零值 | `true` |
| `nested` | 是否嵌套结构体 | `false` |
| `datefmt` | 日期格式 | - |

### 支持的操作符

| 操作符 | 说明 |
|--------|------|
| `=` | 等于 |
| `!=` | 不等于 |
| `>` | 大于 |
| `<` | 小于 |
| `>=` | 大于等于 |
| `<=` | 小于等于 |
| `IN` | 包含于 |
| `NOT IN` | 不包含于 |
| `EXISTS` | 存在 |
| `NOT EXISTS` | 不存在 |
| `CONTAINS` | 包含 |

## 高级功能

### 自定义验证器

```go
builder := meilisearch.NewBuilder()

// 注册自定义验证器
builder.RegisterValidator("age", func(value any) error {
    if age, ok := value.(int); ok {
        if age < 0 || age > 150 {
            return errors.New("age must be between 0 and 150")
        }
    }
    return nil
})

// 在结构体中使用
type UserFilter struct {
    Age int `filter:"age,validate=age"`
}
```

### 自定义转换器

```go
builder := meilisearch.NewBuilder()

// 注册自定义转换器
builder.RegisterTransformer("email", func(value any) any {
    if email, ok := value.(string); ok {
        return strings.ToLower(email)
    }
    return value
})

// 在结构体中使用
type UserFilter struct {
    Email string `filter:"email,transform=email"`
}
```

### 分组过滤

```go
builder := meilisearch.NewBuilder()

// 添加分组过滤条件
condition := meilisearch.FilterCondition{
    Filters: []meilisearch.Filter{
        {Field: "rating", Operator: ">=", Value: 4},
        {Field: "reviews", Operator: ">", Value: 10},
    },
    Operator: "AND",
}

builder.AddFilterGroup("high_quality", condition)

// 应用分组
builder.ApplyFilterGroup("high_quality")
```

## 性能优化建议

1. **合理设置连接池大小**：根据实际并发量调整 `maxIdleConns` 和 `maxConnsPerHost`
2. **使用字段选择**：只返回需要的字段，减少网络传输
3. **合理设置分页大小**：避免单次查询返回过多数据
4. **启用指标采集**：便于监控和性能分析
5. **设置慢查询阈值**：及时发现性能问题

## 错误处理

```go
var (
    // ErrIndexNotFound 表示索引不存在
    ErrIndexNotFound = errors.New("meilisearch index not found")

    // ErrTaskFailed 表示任务执行失败
    ErrTaskFailed = errors.New("meilisearch task failed")

    // ErrInvalidFilter 表示无效的过滤条件
    ErrInvalidFilter = errors.New("invalid filter condition")

    // ErrSlowQuery 表示查询过慢
    ErrSlowQuery = errors.New("meilisearch slow query")

    // ErrMaxRetriesExceeded 表示超过最大重试次数
    ErrMaxRetriesExceeded = errors.New("meilisearch max retries exceeded")
)
```

## 注意事项

1. 确保在使用前正确配置 MeiliSearch 服务器地址和 API 密钥
2. 使用查询构建器时，确保操作符是有效的，否则会被忽略
3. 结构体过滤条件需要使用 `filter` 标签指定字段名和操作符
4. 启用访问日志和指标采集会增加一定的性能开销
5. 慢查询阈值设置过小可能会产生大量日志
6. 重试机制采用指数退避策略，请合理设置重试次数和等待时间

## 下一步

- 了解 [存储组件](./stores.md)
- 探索 [缓存组件](./caches.md)
- 查看 [工具函数](./utils.md)
- 学习 [日志组件](./xlog.md)
