# MeiliSearch 存储组件

MeiliSearch 是一个轻量级、快速、开源的搜索引擎。本组件提供了 Jupiter 框架与 MeiliSearch 的集成，包括客户端管理、查询构建器等功能。

## 功能特性

- 客户端连接池管理
- 查询构建器，支持复杂的搜索条件
- 自动指标采集和日志记录
- 慢查询监控
- 支持结构体过滤条件
- 支持高亮、排序、分面等高级搜索功能
- SearchReader 接口，提供完整的搜索功能

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

### 重试机制说明

MeiliSearch 组件内置了重试机制，可以在网络请求失败或服务器返回错误时自动重试请求，提高系统的可靠性。

#### 重试配置

- `retry`: 重试次数，默认为 2 次。设置为 0 表示不重试。
- `retryWaitTime`: 重试等待时间，默认为 200ms。重试等待时间会采用指数退避策略，每次重试等待时间会翻倍。

#### 重试触发条件

重试机制会在以下情况下触发：

1. **HTTP 请求失败**：网络错误、连接超时等
2. **服务器错误**：HTTP 状态码为 5xx 的服务器内部错误
3. **搜索操作失败**：搜索请求执行失败

#### 重试策略

- 采用指数退避算法，每次重试等待时间会翻倍
- 添加随机抖动，避免多个请求同时重试造成的雷群效应
- 重试次数耗尽后，返回最后一次的错误

### 错误状态收敛

为了减少 Prometheus 指标的基数，MeiliSearch 组件会对错误信息进行收敛处理。系统会根据错误信息的前缀将其归类到有限的状态类别中：

- **meilisearch**：MeiliSearch 相关错误
- **request**：HTTP 请求相关错误
- **search**：搜索操作相关错误
- **server**：服务器相关错误
- **Unknown**：未知错误类型

这种收敛机制可以避免因错误信息变化过大而导致 Prometheus 指标标签过多的问题。

#### 重试示例

```go
// 配置重试参数
config := &meilisearch.Config{
    Addr:          "http://localhost:7700",
    ApiKey:        "your-api-key",
    Retry:         3,                    // 最多重试 3 次
    RetryWaitTime: 100 * time.Millisecond, // 初始等待时间 100ms
}

// 创建客户端
client := meilisearch.Build(config)

// 执行搜索，失败时会自动重试
resp, err := client.SearchWithQuery("users", "john")
if err != nil {
    // 重试次数耗尽后的错误处理
    if errors.Is(err, meilisearch.ErrMaxRetriesExceeded) {
        // 处理超过最大重试次数的错误
    }
}
```
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

### 代码中使用配置

```go
// 使用默认配置
client := meilisearch.StdConfig("default").Build()

// 使用自定义配置
config := &meilisearch.Config{
    Addr:                  "http://localhost:7700",
    ApiKey:                "your-api-key",
    Timeout:               3000 * time.Millisecond,
    KeepAlive:             30 * time.Second,
    MaxIdleConns:          100,
    MaxIdleConnsPerHost:   100,
    IdleConnTimeout:       30 * time.Second,
    DisableKeepAlives:     false,
    MaxConnsPerHost:       100,
    ResponseHeaderTimeout: 30 * time.Second,
    OnDialError:           "panic",
    Retry:                 2,
    RetryWaitTime:         200 * time.Millisecond,
    SlowThreshold:         1 * time.Second,
    Debug:                 false,
    EnableMetric:          true,
    EnableAccessLog:       false,
}
client := meilisearch.Build(config)
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
    config := meilisearch.StdConfig("default")
    client, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    // 直接使用客户端进行搜索
    resp, err := client.SearchWithQuery("users", "john")
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
    config := meilisearch.StdConfig("default")
    client, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("laptop").
        AddFilter("price", "<", 1000).
        AddFilter("brand", "IN", []string{"Apple", "Dell"})
    
    // 直接使用客户端和 Builder 进行搜索
    resp, err := client.SearchWithBuilder("products", builder)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 使用结构体过滤

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

type UserFilter struct {
    Name     string `filter:"name,op=CONTAINS"`
    Age      int    `filter:"age,op=>="`
    Status   string `filter:"status,op=IN"`
}

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    // 创建过滤条件
    filter := UserFilter{
        Name:   "john",
        Age:    18,
        Status: "active",
    }
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("").
        AddFilterStruct(filter)
    
    // 直接使用客户端和 Builder 进行搜索
    resp, err := client.SearchWithBuilder("users", builder)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 传统方式（仍然支持）

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    indexWrapper := client.GetIndex("users")
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("").
        AddFilterStruct(filter)
    
    // 执行搜索 - 注意：Search方法现在需要两个参数：查询字符串和搜索请求
    resp, err := indexWrapper.Search("", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引
    index := client.Index("movies")
    
    // 搜索
    resp, err := index.Search("Batman")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 使用查询构建器

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    indexWrapper := meilisearch.GetIndex(client, "movies", config)
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("Batman").
        SetPage(1).
        SetHitsPerPage(10).
        AddFilter("year", ">", 2010).
        AddFilter("genre", "IN", []string{"Action", "Adventure"}).
        AddSort("title", true).
        AddHighlight("title", "overview").
        EnableMatches()
    
    // 执行搜索 - 注意：Search方法现在需要两个参数：查询字符串和搜索请求
    resp, err := indexWrapper.Search("Batman", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

## SearchReader 接口

SearchReader 接口提供了完整的搜索功能，包括基本搜索、原始搜索、分面搜索和相似文档搜索。IndexWrapper 实现了该接口，所有方法都内置了重试机制、指标采集和日志记录功能。

### 接口定义

```go
// SearchReader 提供完整的搜索功能
type SearchReader interface {
    // Search 执行搜索操作
    Search(query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)
    
    // SearchWithContext 使用上下文执行搜索操作
    SearchWithContext(ctx context.Context, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)
    
    // SearchRaw 执行原始搜索操作，返回JSON响应
    SearchRaw(query string, request *meilisearch.SearchRequest) (*json.RawMessage, error)
    
    // SearchRawWithContext 使用上下文执行原始搜索操作，返回JSON响应
    SearchRawWithContext(ctx context.Context, query string, request *meilisearch.SearchRequest) (*json.RawMessage, error)
    
    // FacetSearch 执行分面搜索操作，返回JSON响应
    FacetSearch(request *meilisearch.FacetSearchRequest) (*json.RawMessage, error)
    
    // FacetSearchWithContext 使用上下文执行分面搜索操作，返回JSON响应
    FacetSearchWithContext(ctx context.Context, request *meilisearch.FacetSearchRequest) (*json.RawMessage, error)
    
    // SearchSimilarDocuments 执行相似文档搜索操作
    SearchSimilarDocuments(param *meilisearch.SimilarDocumentQuery, resp *meilisearch.SimilarDocumentResult) error
    
    // SearchSimilarDocumentsWithContext 使用上下文执行相似文档搜索操作
    SearchSimilarDocumentsWithContext(ctx context.Context, param *meilisearch.SimilarDocumentQuery, resp *meilisearch.SimilarDocumentResult) error
}
```

### 使用示例

#### 基本搜索

```go
package main

import (
    "fmt"
    "context"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    index := meilisearch.GetIndex(client, "movies", config)
    
    // 基本搜索
    resp, err := index.Search("Batman", &meilisearch.SearchRequest{
        Limit: 10,
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

#### 带上下文的搜索

```go
package main

import (
    "fmt"
    "context"
    "time"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    index := meilisearch.GetIndex(client, "movies", config)
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 带上下文的搜索
    resp, err := index.SearchWithContext(ctx, "Batman", &meilisearch.SearchRequest{
        Limit: 10,
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

#### 原始搜索

```go
package main

import (
    "fmt"
    "encoding/json"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    index := meilisearch.GetIndex(client, "movies", config)
    
    // 原始搜索，返回JSON响应
    rawResp, err := index.SearchRaw("Batman", &meilisearch.SearchRequest{
        Limit: 10,
    })
    if err != nil {
        panic(err)
    }
    
    // 解析JSON响应
    var resp map[string]interface{}
    if err := json.Unmarshal(*rawResp, &resp); err != nil {
        panic(err)
    }
    
    fmt.Printf("Raw response: %+v\n", resp)
}
```

#### 分面搜索

```go
package main

import (
    "fmt"
    "encoding/json"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    index := meilisearch.GetIndex(client, "movies", config)
    
    // 分面搜索
    rawResp, err := index.FacetSearch(&meilisearch.FacetSearchRequest{
        FacetQuery: "Action",
        FacetName:  "genre",
        IndexUID:   "movies",
    })
    if err != nil {
        panic(err)
    }
    
    // 解析JSON响应
    var resp map[string]interface{}
    if err := json.Unmarshal(*rawResp, &resp); err != nil {
        panic(err)
    }
    
    fmt.Printf("Facet search response: %+v\n", resp)
}
```

#### 相似文档搜索

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    index := meilisearch.GetIndex(client, "movies", config)
    
    // 相似文档搜索
    var result meilisearch.SimilarDocumentResult
    err = index.SearchSimilarDocuments(&meilisearch.SimilarDocumentQuery{
        Id:       123,
        Embedder: "default",
    }, &result)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d similar documents\n", len(result.Hits))
}
```

### 特性说明

所有 SearchReader 接口方法都具有以下特性：

1. **重试机制**：自动重试失败的请求，提高系统可靠性
2. **指标采集**：记录请求次数和耗时，便于监控和性能分析
3. **日志记录**：记录请求和响应信息，便于问题排查
4. **慢查询监控**：超过阈值的请求会被记录为慢查询
5. **错误收敛**：错误信息会被归类到有限的状态类别，减少指标基数

这些特性使得 SearchReader 接口不仅提供了完整的搜索功能，还具备了生产环境所需的可靠性和可观测性。
```

### 使用结构体过滤条件

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

// MovieFilter 定义电影过滤条件
type MovieFilter struct {
    Title  string `filter:"title,op=CONTAINS"`
    Year   int    `filter:"year,op=>"`
    Genre  string `filter:"genre,op=IN"`
    Active bool   `filter:"active,required"`
}

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    indexWrapper := meilisearch.GetIndex(client, "movies", config)
    
    // 创建过滤条件
    filter := MovieFilter{
        Title:  "Batman",
        Year:   2010,
        Genre:  "Action",
        Active: true,
    }
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("").
        AddFilterStruct(filter)
    
    // 执行搜索 - 注意：Search方法现在需要两个参数：查询字符串和搜索请求
    resp, err := indexWrapper.Search("", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 使用高级过滤条件

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

// UserFilter 定义用户过滤条件
type UserFilter struct {
    Name      string    `filter:"name,op=CONTAINS"`
    Age       int       `filter:"age,op=>="`
    Status    string    `filter:"status,op=IN"`
    CreatedAt time.Time `filter:"created_at,op=BETWEEN"`
    Location  string    `filter:"location,op=GEO_DISTANCE,max=10000"`
}

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    indexWrapper := meilisearch.GetIndex(client, "users", config)
    
    // 创建过滤条件
    filter := UserFilter{
        Name:      "John",
        Age:       18,
        Status:    "active",
        CreatedAt: time.Now().AddDate(0, -1, 0), // 一个月前
        Location:  "40.7128,-74.0060",           // 纽约坐标
    }
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("").
        AddFilterStruct(filter)
    
    // 执行搜索 - 注意：Search方法现在需要两个参数：查询字符串和搜索请求
    resp, err := indexWrapper.Search("", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

### 使用条件组合

```go
package main

import (
    "fmt"
    
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 创建客户端
    config := meilisearch.StdConfig("default")
    client, err := config.Singleton()
    if err != nil {
        panic(err)
    }
    
    // 获取索引包装器
    indexWrapper := meilisearch.GetIndex(client, "products", config)
    
    // 构建查询
    builder := meilisearch.NewBuilder().
        SetQuery("laptop").
        AddFilterCondition("AND", []meilisearch.FilterCondition{
            {Field: "price", Operator: "<", Value: 1000},
            {Field: "brand", Operator: "IN", Value: []string{"Apple", "Dell"}},
        }).
        AddFilterCondition("OR", []meilisearch.FilterCondition{
            {Field: "rating", Operator: ">=", Value: 4.5},
            {Field: "reviews", Operator: ">=", Value: 100},
        })
    
    // 执行搜索 - 注意：Search方法现在需要两个参数：查询字符串和搜索请求
    resp, err := indexWrapper.Search("laptop", builder.Build())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results\n", len(resp.Hits))
}
```

## API 参考

### Config

配置结构体，用于配置 MeiliSearch 客户端。

#### 方法

- `StdConfig(name string) *Config`: 获取标准配置
- `RawConfig(key string) *Config`: 获取原始配置
- `DefaultConfig() *Config`: 获取默认配置
- `Build() (*ClientWrapper, error)`: 创建客户端包装器
- `Singleton() (*ClientWrapper, error)`: 创建单例客户端包装器
- `MustBuild() *ClientWrapper`: 创建客户端包装器，出错时 panic
- `MustSingleton() *ClientWrapper`: 创建单例客户端包装器，出错时 panic

### ClientWrapper

客户端包装器，提供直接的搜索功能。

#### 方法

- `NewClientWrapper(client meilisearch.ServiceManager, config *Config) *ClientWrapper`: 创建新的客户端包装器
- `Search(indexName string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)`: 在指定索引上执行搜索
- `SearchWithQuery(indexName string, query string) (*meilisearch.SearchResponse, error)`: 在指定索引上执行简单的文本搜索
- `SearchWithBuilder(indexName string, builder *Builder) (*meilisearch.SearchResponse, error)`: 使用 Builder 构建查询并在指定索引上执行搜索
- `GetIndex(indexName string) *IndexWrapper`: 获取索引包装器，用于更复杂的操作
- `GetClient() meilisearch.ServiceManager`: 获取原始的 MeiliSearch 客户端

### Builder

查询构建器，用于构建 MeiliSearch 查询。

#### 方法

- `NewBuilder() *Builder`: 创建新的构建器
- `SetQuery(query string) *Builder`: 设置搜索查询字符串
- `SetPage(page int64) *Builder`: 设置页码
- `SetHitsPerPage(hitsPerPage int64) *Builder`: 设置每页结果数量
- `AddFilter(field string, operator string, value any) *Builder`: 添加过滤条件
- `AddFilterStruct(filterStruct any) *Builder`: 通过结构体添加过滤条件
- `AddFilterCondition(logic string, conditions []FilterCondition) *Builder`: 添加组合过滤条件
- `AddGeoFilter(field string, point GeoPoint, distance float64) *Builder`: 添加地理位置过滤
- `AddDateRangeFilter(field string, startDate, endDate time.Time) *Builder`: 添加日期范围过滤
- `RegisterValidator(field string, validator func(any) error) *Builder`: 注册字段验证器
- `AddFacets(facets ...string) *Builder`: 添加分面字段
- `AddSort(field string, ascending bool) *Builder`: 添加排序字段
- `SetRankingScoreThreshold(rankingScoreThreshold float64) *Builder`: 设置排名分数阈值
- `AddHighlight(fields ...string) *Builder`: 添加高亮字段
- `HighlightPreTag(preTag string) *Builder`: 设置高亮标签前缀
- `HighlightPostTag(postTag string) *Builder`: 设置高亮标签后缀
- `AddAttributesToRetrieve(fields ...string) *Builder`: 添加要返回的字段
- `AttributesToSearchOn(fields ...string) *Builder`: 设置搜索字段
- `EnableMatches() *Builder`: 启用匹配信息
- `ShowRankingScore() *Builder`: 显示排名分数
- `ShowRankingScoreDetails() *Builder`: 显示排名分数详细信息
- `Build() meilisearch.SearchRequest`: 构建最终的搜索参数

### IndexWrapper

索引包装器，提供带有指标和日志的搜索功能。

#### 方法

- `NewIndexWrapper(index meilisearch.IndexManager, name string, config *Config) *IndexWrapper`: 创建新的索引包装器
- `Search(query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)`: 执行搜索操作，需要查询字符串和搜索请求两个参数

### 全局函数

- `GetIndex(client meilisearch.ServiceManager, indexName string, config *Config) *IndexWrapper`: 获取指定索引的包装器
- `GetIndexWithUID(client meilisearch.ServiceManager, uid string, config *Config) *IndexWrapper`: 使用 UID 获取索引包装器

## 支持的操作符

- `=`: 等于
- `!=`: 不等于
- `>`: 大于
- `<`: 小于
- `>=`: 大于等于
- `<=`: 小于等于
- `IN`: 包含于
- `NOT IN`: 不包含于
- `EXISTS`: 存在
- `NOT EXISTS`: 不存在
- `CONTAINS`: 包含
- `STARTS_WITH`: 以...开始
- `ENDS_WITH`: 以...结束
- `BETWEEN`: 在...之间
- `GEO_DISTANCE`: 地理位置距离

## 结构体过滤标签

结构体字段可以使用 `filter` 标签来配置过滤条件：

```go
type UserFilter struct {
    Name     string    `filter:"name,op=CONTAINS"`
    Age      int       `filter:"age,op=>=,required"`
    Status   string    `filter:"status,op=IN,transform=strings.ToLower"`
    CreateAt time.Time `filter:"created_at,op=BETWEEN,validate=customDateValidator"`
    Location string    `filter:"location,op=GEO_DISTANCE,max=10000"`
}
```

标签参数说明：
- `field`: 字段名，默认使用结构体字段名
- `op`: 操作符，默认为 `=`
- `required`: 是否必填，默认为 false
- `transform`: 转换函数，用于在过滤前转换值
- `validate`: 验证函数，用于验证值的有效性
- `max`: 最大值，用于 GEO_DISTANCE 操作符
- `min`: 最小值，用于范围操作符

## 高级功能

### 自定义验证器

```go
// 注册自定义验证器
builder.RegisterValidator("email", func(value any) error {
    email, ok := value.(string)
    if !ok {
        return fmt.Errorf("email must be a string")
    }
    
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email format")
    }
    
    return nil
})

// 在结构体中使用
type UserFilter struct {
    Email string `filter:"email,op=CONTAINS,validate=email"`
}
```

### 自定义转换器

```go
// 注册自定义转换器
builder.RegisterTransformer("phone", func(value any) (any, error) {
    phone, ok := value.(string)
    if !ok {
        return nil, fmt.Errorf("phone must be a string")
    }
    
    // 移除所有非数字字符
    reg := regexp.MustCompile(`\D`)
    return reg.ReplaceAllString(phone, ""), nil
})

// 在结构体中使用
type UserFilter struct {
    Phone string `filter:"phone,op=CONTAINS,transform=phone"`
}
```

## 注意事项

1. 确保在使用前正确配置 MeiliSearch 服务器地址和 API 密钥
2. 使用查询构建器时，确保操作符是有效的，否则会被忽略
3. 结构体过滤条件需要使用 `filter` 标签指定字段名和操作符
4. 启用访问日志和指标采集会增加一定的性能开销
5. 慢查询阈值设置过小可能会产生大量日志
6. **重要变更**：`IndexWrapper.Search` 方法现在需要两个参数：查询字符串和搜索请求，而不是只有一个搜索请求参数

## 性能优化建议

1. 合理设置连接池大小，避免过多连接消耗资源
2. 使用 `AttributesToRetrieve` 只返回需要的字段，减少网络传输
3. 使用 `AttributesToSearchOn` 限制搜索范围，提高搜索速度
4. 合理设置分页大小，避免单次查询返回过多数据
5. 使用索引包装器可以自动记录指标和日志，便于性能监控