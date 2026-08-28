# MeiliSearch 过滤器使用指南

本文档介绍了 Jupiter 框架中 MeiliSearch 组件的增强过滤器功能，特别是如何使用结构体来构建复杂的过滤条件。

## 目录

1. [基本过滤器](#基本过滤器)
2. [结构体过滤器](#结构体过滤器)
3. [高级过滤器](#高级过滤器)
4. [类型安全的数组过滤](#类型安全的数组过滤)
5. [自定义验证器和转换器](#自定义验证器和转换器)
6. [Client 便捷方法](#client-便捷方法)
7. [示例代码](#示例代码)

## 基本过滤器

### 单个过滤条件

```go
builder := NewBuilder()
builder.AddFilter("status", "=", "active")
builder.AddFilter("age", ">=", 18)
```

### 过滤条件组合

```go
// AND 条件
andCondition := FilterCondition{
    Filters: []Filter{
        {Field: "status", Operator: "=", Value: "active"},
        {Field: "verified", Operator: "=", Value: true},
    },
    Operator: "AND",
}
builder.AddFilterCondition(andCondition)

// OR 条件
orCondition := FilterCondition{
    Filters: []Filter{
        {Field: "status", Operator: "=", Value: "active"},
        {Field: "status", Operator: "=", Value: "pending"},
    },
    Operator: "OR",
}
builder.AddFilterCondition(orCondition)
```

## 结构体过滤器

使用结构体可以更方便地组织和管理过滤条件。

### 基本用法

```go
type UserFilter struct {
    Name     string `filter:"name,op=CONTAINS"`
    Age      int    `filter:"age,op=>="`
    Status   string `filter:"status,op=IN,required"`
    Email    string `filter:"email,ignorezero"`
    IsActive bool   `filter:"is_active,op=="`
}

userFilter := UserFilter{
    Name:     "John",
    Age:      25,
    Status:   "active",
    Email:    "john@example.com",
    IsActive: true,
}

builder.AddFilterStruct(userFilter)
```

### 过滤器标签语法

过滤器标签使用以下格式：

```
`filter:"字段名,属性1=值1,属性2=值2,..."`
```

#### 支持的属性

| 属性 | 描述 | 示例 |
|------|------|------|
| `op` | 操作符 | `op=CONTAINS` |
| `required` | 必填字段 | `required` |
| `ignorezero` | 忽略零值 | `ignorezero` |
| `nested` | 嵌套结构体 | `nested` |
| `group` | 分组名称 | `group=user` |
| `transform` | 转换函数 | `transform=lowercase` |
| `validate` | 验证函数 | `validate=email` |
| `deps` | 依赖字段 | `deps=field1;field2` |
| `datefmt` | 日期格式 | `datefmt=2006-01-02` |

#### 支持的操作符

| 操作符 | 描述 | 示例 |
|--------|------|------|
| `=` | 等于 | `status = "active"` |
| `!=` | 不等于 | `status != "inactive"` |
| `>` | 大于 | `age > 18` |
| `<` | 小于 | `age < 65` |
| `>=` | 大于等于 | `age >= 18` |
| `<=` | 小于等于 | `age <= 65` |
| `IN` | 在列表中 | `status IN ["active", "pending"]` |
| `NOT IN` | 不在列表中 | `status NOT IN ["inactive", "banned"]` |
| `CONTAINS` | 包含 | `name CONTAINS "john"` |
| `NOT CONTAINS` | 不包含 | `name NOT CONTAINS "admin"` |
| `EXISTS` | 字段存在 | `email EXISTS` |
| `NOT EXISTS` | 字段不存在 | `deleted_at NOT EXISTS` |

## 高级过滤器

### 地理位置过滤

```go
geoFilter := GeoFilter{
    Point:  GeoPoint{Lat: 40.7128, Lng: -74.0060}, // 纽约坐标
    Radius: 10000, // 10公里
    Unit:   "m",
}
builder.AddGeoFilter(geoFilter)
```

### 日期范围过滤

```go
dateFilter := DateRangeFilter{
    Field: "created_at",
    Start: time.Now().AddDate(-1, 0, 0), // 一年前
    End:   time.Now(),                   // 现在
}
builder.AddDateRangeFilter(dateFilter)
```

### 数组过滤

```go
arrayFilter := ArrayFilter{
    Field:    "categories",
    Values:   []any{"electronics", "books"},
    Operator: "IN",
}
builder.AddArrayFilter(arrayFilter)
```

### 分组过滤

```go
// 添加到分组
builder.AddFilterStructWithGroup("product", productFilter)

// 添加另一个过滤器到同一分组
condition := FilterCondition{
    Filters: []Filter{
        {Field: "rating", Operator: ">=", Value: 4},
        {Field: "reviews", Operator: ">", Value: 10},
    },
    Operator: "AND",
}
builder.AddFilterGroup("product", condition)

// 应用分组过滤器
builder.ApplyFilterGroup("product")
```

## 类型安全的数组过滤

从最新版本开始，`AddFilter` 方法支持多种切片类型，自动进行类型转换：

```go
builder := NewBuilder()

// 支持 []string 类型
builder.AddFilter("tags", "IN", []string{"golang", "rust", "python"})

// 支持 []int 类型
builder.AddFilter("ids", "IN", []int{1, 2, 3, 4, 5})

// 支持 []int64 类型
builder.AddFilter("ids", "IN", []int64{1, 2, 3, 4, 5})

// 支持 []float64 类型
builder.AddFilter("prices", "IN", []float64{9.99, 19.99, 29.99})

// 也支持原来的 []any 类型
builder.AddFilter("categories", "IN", []any{"electronics", "books"})
```

## 自定义验证器和转换器

### 注册验证器

```go
builder.RegisterValidator("age", func(value any) error {
    if age, ok := value.(int); ok {
        if age < 0 || age > 150 {
            return errors.New("age must be between 0 and 150")
        }
    }
    return nil
})
```

### 注册转换器

```go
builder.RegisterTransformer("email", func(value any) any {
    if email, ok := value.(string); ok {
        return strings.ToLower(email)
    }
    return value
})
```

## Client 便捷方法

Client 提供了便捷的方法来操作 MeiliSearch：

### 创建客户端

```go
// 使用标准配置
client, err := meilisearch.StdConfig("default").Singleton()
if err != nil {
    panic(err)
}

// 使用自定义配置
client, err := meilisearch.RawConfig("search.custom").Singleton()

// 必须成功创建
client = meilisearch.StdConfig("default").MustSingleton()
```

### 搜索操作

```go
// 执行搜索
resp, err := client.Search("products", &meilisearch.SearchRequest{
    Query:       "smartphone",
    Page:        1,
    HitsPerPage: 20,
})
```

### 文档操作

```go
// 获取文档列表
var result meilisearch.DocumentsResult
err := client.GetDocuments("products", &meilisearch.DocumentsQuery{
    Limit: 10,
}, &result)

// 获取单个文档
var doc map[string]any
err = client.GetDocument("products", "1", &meilisearch.DocumentQuery{}, &doc)

// 添加文档
docs := []map[string]any{
    {"id": 1, "name": "Product 1", "price": 99.99},
    {"id": 2, "name": "Product 2", "price": 149.99},
}
task, err := client.AddDocuments("products", docs, nil)

// 添加文档（指定主键）
primaryKey := "id"
task, err := client.AddDocuments("products", docs, &primaryKey)

// 更新文档
task, err := client.UpdateDocuments("products", docs, nil)

// 删除单个文档
task, err := client.DeleteDocument("products", "1")

// 删除所有文档
task, err := client.DeleteAllDocuments("products")
```

### 索引操作

```go
// 获取索引操作对象
index := client.Index("products")

// 使用索引进行更复杂的操作
info, err := index.GetStats()
```

## 示例代码

### 完整示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
)

func main() {
    // 定义过滤器结构体
    type UserFilter struct {
        Name     string     `filter:"name,op=CONTAINS"`
        Age      int        `filter:"age,op=>="`
        Status   string     `filter:"status,op=IN,required"`
        Email    string     `filter:"email,ignorezero"`
        JoinDate time.Time  `filter:"join_date,op=>=,datefmt=2006-01-02"`
        Tags     []string   `filter:"tags,op=CONTAINS"`
        IsActive bool       `filter:"is_active,op=="`
        Location Location   `filter:"location,nested"`
    }

    type Location struct {
        City    string `filter:"city"`
        Country string `filter:"country,required"`
    }

    // 创建构建器
    builder := meilisearch.NewBuilder()

    // 注册自定义验证器
    builder.RegisterValidator("age", func(value any) error {
        if age, ok := value.(int); ok {
            if age < 0 || age > 150 {
                return fmt.Errorf("age must be between 0 and 150, got %d", age)
            }
        }
        return nil
    })

    // 注册自定义转换器
    builder.RegisterTransformer("email", func(value any) any {
        if email, ok := value.(string); ok {
            return strings.ToLower(email)
        }
        return value
    })

    // 创建过滤器实例
    userFilter := UserFilter{
        Name:     "John",
        Age:      25,
        Status:   "active",
        Email:    "John.Doe@Example.com", // 将被转换为小写
        JoinDate: time.Now().AddDate(-1, 0, 0), // 一年前
        Tags:     []string{"developer", "golang"},
        IsActive: true,
        Location: Location{
            City:    "New York",
            Country: "USA",
        },
    }

    // 添加结构体过滤器
    builder.AddFilterStruct(userFilter)

    // 添加地理位置过滤
    geoFilter := meilisearch.GeoFilter{
        Point:  meilisearch.GeoPoint{Lat: 40.7128, Lng: -74.0060}, // 纽约坐标
        Radius: 10000, // 10公里
        Unit:   "m",
    }
    builder.AddGeoFilter(geoFilter)

    // 添加日期范围过滤
    dateFilter := meilisearch.DateRangeFilter{
        Field: "created_at",
        Start: time.Now().AddDate(-1, 0, 0), // 一年前
        End:   time.Now(),                   // 现在
    }
    builder.AddDateRangeFilter(dateFilter)

    // 添加其他查询参数
    builder.SetQuery("search query").
        SetPage(1).
        SetHitsPerPage(20).
        AddSort("age", false). // 按年龄降序
        AddFacets("status", "city").
        AddHighlight("name", "email").
        EnableMatches().
        ShowRankingScore()

    // 构建查询参数
    searchRequest := builder.Build()

    // 输出结果
    fmt.Printf("Generated search request: %+v\n", searchRequest)
}
```

### 嵌套结构体示例

```go
type ProductAttributes struct {
    Brand    string `filter:"brand,required"`
    Color    string `filter:"color"`
    Size     string `filter:"size,ignorezero"`
    Material string `filter:"material"`
}

type ProductFilter struct {
    Name       string            `filter:"name,op=CONTAINS"`
    Category   string            `filter:"category,required"`
    Price      float64           `filter:"price,op=>="`
    InStock    bool              `filter:"in_stock,op=="`
    CreatedAt  time.Time         `filter:"created_at,op=>=,datefmt=2006-01-02"`
    Tags       []string          `filter:"tags,op=CONTAINS"`
    Attributes ProductAttributes `filter:"attributes,nested"`
}

productFilter := ProductFilter{
    Name:      "Smartphone",
    Category:  "electronics",
    Price:     299.99,
    InStock:   true,
    CreatedAt: time.Now().AddDate(-2, 0, 0), // 两年前
    Tags:      []string{"smartphone", "android"},
    Attributes: ProductAttributes{
        Brand:    "TechCorp",
        Color:    "black",
        Size:     "6.5 inch",
        Material: "glass",
    },
}

builder.AddFilterStruct(productFilter)
```

### 类型安全数组过滤示例

```go
builder := meilisearch.NewBuilder()

// 使用 []string 类型
builder.AddFilter("tags", "IN", []string{"golang", "rust", "python"})

// 使用 []int 类型
builder.AddFilter("ids", "IN", []int{1, 2, 3, 4, 5})

// 使用 []float64 类型
builder.AddFilter("prices", "IN", []float64{9.99, 19.99, 29.99})

searchRequest := builder.Build()
```

### Client 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/store/meilisearch"
    "github.com/meilisearch/meilisearch-go"
)

func main() {
    // 创建客户端
    client, err := meilisearch.StdConfig("default").Singleton()
    if err != nil {
        panic(err)
    }

    // 使用 Builder 构建搜索请求
    builder := meilisearch.NewBuilder()
    builder.SetQuery("smartphone").
        SetPage(1).
        SetHitsPerPage(20).
        AddFilter("category", "=", "electronics").
        AddFilter("price", "<=", 500).
        AddSort("price", true)

    // 执行搜索
    resp, err := client.Search("products", builder.Build())
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d results\n", resp.TotalHits)

    // 添加文档
    docs := []map[string]any{
        {"id": "1", "name": "iPhone 15", "price": 999.99, "category": "electronics"},
        {"id": "2", "name": "Samsung Galaxy", "price": 799.99, "category": "electronics"},
    }
    task, err := client.AddDocuments("products", docs, nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Task UID: %d\n", task.TaskUID)

    // 获取文档
    var result meilisearch.DocumentsResult
    err = client.GetDocuments("products", &meilisearch.DocumentsQuery{
        Limit: 10,
    }, &result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Total documents: %d\n", result.Total)
}
```

## 最佳实践

1. **使用结构体过滤器**：对于复杂的查询，使用结构体过滤器可以使代码更清晰、更易于维护。

2. **合理使用标签**：
   - 使用 `required` 标记必填字段
   - 使用 `ignorezero` 忽略零值字段
   - 使用 `nested` 处理嵌套结构体

3. **自定义验证和转换**：
   - 对于需要特殊处理的字段，注册自定义验证器和转换器
   - 例如，邮箱转换为小写、年龄范围验证等

4. **分组过滤**：
   - 对于相关的过滤条件，使用分组功能
   - 可以根据需要选择性应用分组

5. **类型安全**：
   - 使用类型安全的数组过滤（`[]string`、`[]int`、`[]float64` 等）
   - 避免使用 `[]any` 除非必要

6. **性能考虑**：
   - 避免创建过于复杂的嵌套结构
   - 合理使用索引字段进行过滤

## 更新日志

### v1.1.0
- 新增 `appendFilter` 辅助方法，消除重复代码
- `isValidOperator` 性能优化，使用 map 替代切片遍历
- `formatFilter` 类型安全处理，支持 `[]string`、`[]int`、`[]int64`、`[]float64` 类型
- `toString` 函数扩展，支持更多数值类型和 `fmt.Stringer` 接口
- 使用框架日志组件替代 `fmt.Printf`
- `transport.go` 请求体复制逻辑优化
- Client 新增便捷方法：`Search`、`GetDocuments`、`GetDocument`、`AddDocuments`、`UpdateDocuments`、`DeleteDocument`、`DeleteAllDocuments`

## 总结

MeiliSearch 的增强过滤器功能提供了强大而灵活的方式来构建复杂的搜索查询。通过结构体过滤器，可以大大简化代码并提高可维护性。结合自定义验证器和转换器，可以满足各种复杂的业务需求。新增的类型安全数组过滤和 Client 便捷方法进一步提升了开发体验。
