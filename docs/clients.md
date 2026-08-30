# Jupiter 客户端组件

本文档介绍 Jupiter 框架中的各种客户端组件，包括 HTTP 客户端、Redis 客户端、Kafka 客户端等，以及它们的使用方法和配置选项。

## 概述

Jupiter 框架提供了丰富的客户端组件，用于与外部服务进行通信。这些客户端组件封装了连接管理、负载均衡、重试机制等功能，简化了与外部服务的交互。框架提供了统一的配置和接口，使开发者可以轻松地使用各种客户端组件。

## HTTP 客户端

### 基本使用

```go
import (
    "github.com/weijinwen/jupiter/pkg/client/resty"
)

// 创建 HTTP 客户端
client := resty.StdConfig("example").MustSingleton()

// GET 请求
resp, err := client.R().Get("https://api.example.com/users")
if err != nil {
    panic(err)
}
fmt.Printf("响应: %s\n", resp.String())

// POST 请求
resp, err = client.R().
    SetBody(map[string]interface{}{
        "name": "John",
        "age":  30,
    }).
    Post("https://api.example.com/users")
if err != nil {
    panic(err)
}
fmt.Printf("响应: %s\n", resp.String())
```

### 配置选项

```toml
# HTTP 客户端配置
[resty.example]
# 目标服务地址（必填）
addr = "https://api.example.com"
# 请求超时时间，默认 3000ms
timeout = "3000ms"
# 失败重试次数，默认 0（不重试）
retryCount = 0
# 失败重试间隔时间，默认 100ms
retryWaitTime = "100ms"
# 失败重试最大等待时间，默认 100ms
retryMaxWaitTime = "100ms"
# 收到响应后是否立即关闭连接
closeConnection = false
# 调试开关
debug = false
# 指标采集开关，默认开启
enableMetric = true
# 访问日志开关
enableAccessLog = false
# 慢日志阈值，默认 500ms
slowThreshold = "500ms"

# 默认请求头
[resty.example.headers]
Content-Type = "application/json"
Accept = "application/json"

# 认证配置（AuthToken 和 Username/Password 二选一，AuthToken 优先）
authToken = ""
username = ""
password = ""
```

### 请求和响应处理

```go
// 设置请求头
client.R().SetHeader("Content-Type", "application/json")

// 设置查询参数
client.R().SetQueryParam("page", "1").SetQueryParam("size", "10")

// 设置表单数据
client.R().SetFormData(map[string]string{
    "username": "john",
    "password": "password",
})

// 设置文件
client.R().SetFile("file", "./example.txt")

// 解析响应
var user User
err := client.R().SetResult(&user).Get("https://api.example.com/users/1")
if err != nil {
    panic(err)
}
fmt.Printf("用户: %s\n", user.Name)
```

### 重试配置

```go
import (
    "github.com/go-resty/resty/v2"
    "github.com/weijinwen/jupiter/pkg/client/resty"
)

// 创建客户端
config := resty.StdConfig("example")
client := config.MustSingleton()

// 设置重试条件
client.SetRetryCondition(func(r *resty.Response, err error) bool {
    // 返回 true 表示需要重试
    return r.StatusCode() >= 500
})

// 或者通过配置设置
config.RetryCondition = func(r *resty.Response, err error) bool {
    return r.StatusCode() >= 500
}
```

## Redis 客户端

### 基本使用

```go
import (
    "github.com/weijinwen/jupiter/pkg/client/redis"
)

// 创建 Redis 客户端
client := redis.StdConfig("default").MustSingleton()

// 字符串操作
err := client.Set(context.Background(), "key", "value", 10*time.Minute).Err()
if err != nil {
    panic(err)
}
value, err := client.Get(context.Background(), "key").Result()
if err != nil {
    panic(err)
}
fmt.Printf("值: %s\n", value)

// 哈希操作
err = client.HSet(context.Background(), "user:1", "name", "John", "age", 30).Err()
if err != nil {
    panic(err)
}
name, err := client.HGet(context.Background(), "user:1", "name").Result()
if err != nil {
    panic(err)
}
fmt.Printf("用户名: %s\n", name)
```

### 配置选项

```toml
# Redis 客户端配置
[client.redis.default]
# 连接地址（支持集群模式，配置多个地址）
addr = ["localhost:6379"]
# 用户名（Redis 6.0+ ACL）
username = ""
# 密码
password = ""
# 数据库编号，默认 0
db = 0
# 连接池大小（每个节点），默认 200
poolSize = 200
# 最小空闲连接数，默认 20
minIdleConns = 20
# 最大重试次数，-1 表示不重试，默认 3
maxRetries = 3
# 连接超时时间，默认 3s
dialTimeout = "3s"
# 读取超时时间，默认 1s
readTimeout = "1s"
# 写入超时时间，默认 1s
writeTimeout = "1s"
# 空闲连接超时时间，默认 60s
idleTimeout = "60s"

# 调试开关
debug = false
# 慢日志阈值（请求耗时超过此值会被记录）
slowThreshold = "0s"
# 指标采集开关
enableMetric = false
# 访问日志开关
enableAccessLog = false
# 连接失败时的错误级别：panic 或 error
level = "error"

# 分布式锁配置
# 锁默认过期时间，默认 30s
lockDefaultExpiration = "30s"
# 锁默认重试间隔，默认 100ms
lockDefaultRetryDelay = "100ms"
# 锁默认最大重试次数，-1 表示无限重试
lockDefaultMaxRetries = -1
# 锁默认获取超时时间，0 表示不超时
lockDefaultTimeout = "0s"
```

### 使用示例

```go
// 从配置创建客户端
client := redis.StdConfig("default").MustSingleton()

// 使用快速创建函数
client := redis.NewRedis("default")

// 分布式锁操作
lock := client.GetLock(
    redis.WithLockExpiration(30 * time.Second),
    redis.WithLockRetryDelay(100 * time.Millisecond),
)

// 获取锁
ctx := context.Background()
err := lock.Lock(ctx, "my-resource")
if err != nil {
    panic(err)
}
defer lock.Unlock(ctx)
```

### 分布式锁配置

```toml
[client.redis.default]
# 分布式锁默认过期时间
lockDefaultExpiration = "30s"
# 分布式锁重试间隔
lockDefaultRetryDelay = "100ms"
# 分布式锁最大重试次数，-1 表示无限重试
lockDefaultMaxRetries = -1
# 分布式锁获取超时时间，0 表示不超时
lockDefaultTimeout = "0s"
```

### 事务

```go
// 使用事务
pipe := client.TxPipeline()
pipe.Incr(ctx, "counter")
pipe.Expire(ctx, "counter", time.Hour)
_, err := pipe.Exec(ctx)
if err != nil {
    panic(err)
}
```

## 客户端最佳实践

### 1. 连接池配置

合理配置连接池大小：

```go
// 根据业务场景配置连接池
client := redis.New(
    redis.WithConfig("default"),
    redis.WithPoolSize(100),        // 连接池大小
    redis.WithMinIdleConns(10),     // 最小空闲连接
    redis.WithMaxRetries(3),        // 最大重试次数
    redis.WithDialTimeout(5*time.Second),  // 连接超时
    redis.WithReadTimeout(3*time.Second),  // 读取超时
    redis.WithWriteTimeout(3*time.Second), // 写入超时
)
```

### 2. 超时设置

设置合理的超时时间：

```go
// 设置请求超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.GetUser(ctx, &userpb.GetUserRequest{
    Id: 1,
})
if err != nil {
    panic(err)
}
```

### 3. 错误处理

统一错误处理：

```go
// 定义错误处理函数
func handleHTTPError(err error) error {
    if err == nil {
        return nil
    }

    // 根据错误类型进行处理
    if errors.Is(err, context.DeadlineExceeded) {
        return xerror.New(xcode.Timeout, "请求超时")
    }

    return xerror.Wrap(xcode.Internal, "调用服务失败", err)
}

// 使用错误处理函数
resp, err := client.R().Get("https://api.example.com/users/1")
if err != nil {
    return handleHTTPError(err)
}
```

### 4. 资源管理

正确管理客户端资源：

```go
// 创建客户端
client := redis.StdConfig("default").Build()

// 使用客户端
defer client.Close()

// 使用客户端进行操作
err := client.Set("key", "value", 10*time.Minute).Err()
if err != nil {
    panic(err)
}
```

## 总结

Jupiter 框架提供了丰富的客户端组件，支持多种协议和服务。通过统一的配置和接口，简化了与外部服务的交互。合理使用这些客户端组件，可以构建高性能、高可用的微服务应用。

## 下一步

- 了解 [存储组件](./stores.md)
- 学习 [缓存组件](./caches.md)
- 探索 [日志组件](./xlog.md)
- 查看 [配置系统](./config.md)