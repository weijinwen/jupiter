# Jupiter 缓存组件

本文档介绍 Jupiter 框架中的各种缓存组件，包括 FreeCache、GolangLRU 等，以及它们的使用方法和配置选项。

## 概述

Jupiter 框架提供了丰富的缓存组件，用于提高应用程序的性能和响应速度。这些缓存组件支持不同的缓存策略和存储后端，可以根据应用需求选择合适的缓存方案。

## FreeCache 缓存

### 基本使用

```go
import (
    "github.com/weijinwen/jupiter/pkg/cache/xfreecache"
)

// 创建 FreeCache 配置
cfg := xfreecache.Config{
    Size:   256 * xfreecache.MB,  // 缓存容量，最小512*1024
    Expire: 2 * time.Minute,      // 失效时间
    Name:   "default",            // 缓存名称
}

// 构建缓存实例
cache := cfg.Build()

// 使用缓存（通过泛型 Cache 接口）
type User struct {
    ID   int
    Name string
    Age  int
}

// 获取并设置缓存数据
user, err := cache.GetAndSetCacheData("user", 1, func() (User, error) {
    // 从数据库加载用户数据
    return getUserFromDB(1)
})
```

### 配置选项

```toml
# FreeCache 配置
[xfreecache.default]
# 缓存容量，最小 512KB，最大 8GB（必填）
# 可以使用 Size 类型：Size = 256 * MB
size = 256MB
# 失效时间（必填）
expire = "2m"
# metric 上报开关，false 开启，true 关闭，默认开启
disableMetric = false
# 本地缓存名称，用于日志标识和 metric 上报（选填）
name = "default"
```

### Size 常量

```go
const (
    B  Size = 1
    KB Size = 1024 * B
    MB Size = 1024 * KB
    GB Size = 1024 * MB
)
```

### 注意事项

- 缓存容量最小为 512KB
- 单个条目大小必须小于缓存容量的 1/1024
- 失效时间必须大于 0

## GolangLRU 缓存

### 基本使用

```go
import (
    "github.com/weijinwen/jupiter/pkg/cache/xgolanglru"
)

// 使用标准配置创建 LRU 缓存
// 通过泛型支持指定键和值的类型
cache := xgolanglru.StdNew[int, *User]("default")

// 获取并设置缓存数据
user, err := cache.GetAndSetCacheData("user", 1, func() (*User, error) {
    // 从数据库加载用户数据
    return getUserFromDB(1)
})

// 使用自定义配置创建缓存
config := &xgolanglru.Config{
    Size:     200000,           // 缓存大小，默认 200000
    Expire:   5 * time.Minute,  // 失效时间（必填）
    Name:     "custom",         // 缓存名称
}
cache = xgolanglru.New[int, *User](config)
```

### 配置选项

```toml
# GolangLRU 配置
[jupiter.xgolanglru.default]
# 缓存大小，默认 200000
size = 200000
# 失效时间（必填）
expire = "2m"
# metric 上报开关，false 开启，true 关闭，默认开启
disableMetric = false
```

### 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `size` | int | 200000 | 缓存条目数量上限 |
| `expire` | duration | - | 失效时间（必填） |
| `disableMetric` | bool | false | 是否禁用指标上报 |
| `name` | string | "default" | 缓存名称（代码设置） |

### 注意事项

- 失效时间必须大于 0
- 缓存大小默认为 200000 条目
- 支持泛型，可指定键和值的类型

## 缓存最佳实践

### 1. 缓存键设计

合理设计缓存键：

```go
// 使用命名空间
userKey := fmt.Sprintf("user:%d", userID)
productKey := fmt.Sprintf("product:%d", productID)

// 使用层级结构
userProfileKey := fmt.Sprintf("user:%d:profile", userID)
userOrdersKey := fmt.Sprintf("user:%d:orders", userID)

// 使用版本号
userKeyV1 := fmt.Sprintf("user:v1:%d", userID)
userKeyV2 := fmt.Sprintf("user:v2:%d", userID)
```

### 2. 缓存过期策略

合理设置缓存过期时间：

```go
// 短期缓存，适用于频繁变化的数据
cache.Set(key, value, 5*time.Minute)

// 中期缓存，适用于不经常变化的数据
cache.Set(key, value, 1*time.Hour)

// 长期缓存，适用于几乎不变的数据
cache.Set(key, value, 24*time.Hour)

// 随机过期，防止缓存雪崩
randomExpire := time.Duration(rand.Intn(300)+300) * time.Second // 5-10分钟
cache.Set(key, value, randomExpire)
```

### 3. 缓存穿透防护

防止缓存穿透：

```go
// 缓存空值
func getUserByID(userID int) (*User, error) {
    key := fmt.Sprintf("user:%d", userID)

    // 尝试从缓存获取
    user, err := cache.GetAndSetCacheData("user", userID, func() (*User, error) {
        // 从数据库获取
        user, err := getUserFromDB(userID)
        if err != nil {
            return nil, err
        }
        return user, nil
    })
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

## 总结

Jupiter 框架提供了丰富的缓存组件，支持不同的缓存策略和存储后端。合理使用这些缓存组件，可以显著提高应用程序的性能和响应速度。

## 下一步

- 了解 [工具函数](./utils.md)
- 学习 [Worker 命令执行](./worker.md)
- 探索 [日志组件](./xlog.md)
- 查看 [配置系统](./config.md)
