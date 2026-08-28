# Jupiter 缓存组件

本文档介绍 Jupiter 框架中的各种缓存组件，包括 FreeCache、GolangLRU 等，以及它们的使用方法和配置选项。

## 概述

Jupiter 框架提供了丰富的缓存组件，用于提高应用程序的性能和响应速度。这些缓存组件支持不同的缓存策略和存储后端，可以根据应用需求选择合适的缓存方案。

## FreeCache 缓存

### 基本使用

```go
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
[cache.freecache.default]
size = 256  # 缓存大小，单位MB
expire = "2m"  # 过期时间
disableMetric = false  # 是否禁用指标上报
```

## GolangLRU 缓存

### 基本使用

```go
// 使用标准配置创建 LRU 缓存
// 通过泛型支持指定键和值的类型
cache := xgolanglru.StdNew[int, *User]("default")

// 获取并设置缓存数据
user, err := cache.GetAndSetCacheData("user", 1, func() (*User, error) {
    // 从数据库加载用户数据
    return getUserFromDB(1)
})
```

### 配置选项

```toml
# GolangLRU 配置
[cache.golanglru.default]
size = 200000  # 缓存大小
expire = "2m"  # 过期时间
disableMetric = false  # 是否禁用指标上报
```

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
