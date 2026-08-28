# Jupiter 客户端组件

本文档介绍 Jupiter 框架中的各种客户端组件，包括 HTTP 客户端、Redis 客户端、Kafka 客户端等，以及它们的使用方法和配置选项。

## 概述

Jupiter 框架提供了丰富的客户端组件，用于与外部服务进行通信。这些客户端组件封装了连接管理、负载均衡、重试机制等功能，简化了与外部服务的交互。框架提供了统一的配置和接口，使开发者可以轻松地使用各种客户端组件。

## HTTP 客户端

### 基本使用

```go
// 创建 HTTP 客户端
client := resty.StdConfig("example").Build()

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
[client.http.example]
base_url = "https://api.example.com"
timeout = "30s"
retry_count = 3
retry_wait_time = "100ms"
debug = false

# TLS 配置
[client.http.example.tls]
enable = false
cert_file = ""
key_file = ""
insecure_skip_verify = false

# 代理配置
[client.http.example.proxy]
enable = false
url = ""
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

## Redis 客户端

### 基本使用

```go
// 创建 Redis 客户端
client := redis.StdConfig("default").Build()

// 字符串操作
err := client.Set("key", "value", 10*time.Minute).Err()
if err != nil {
    panic(err)
}
value, err := client.Get("key").Result()
if err != nil {
    panic(err)
}
fmt.Printf("值: %s\n", value)

// 哈希操作
err = client.HSet("user:1", "name", "John", "age", 30).Err()
if err != nil {
    panic(err)
}
name, err := client.HGet("user:1", "name").Result()
if err != nil {
    panic(err)
}
fmt.Printf("用户名: %s\n", name)
```

### 配置选项

```toml
# Redis 客户端配置
[client.redis.default]
address = "localhost:6379"
password = ""
db = 0
pool_size = 10
min_idle_conns = 5
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"
idle_timeout = "5m"
idle_check_frequency = "1m"

# 集群配置
[client.redis.cluster]
addresses = ["redis1:6379", "redis2:6379", "redis3:6379"]
password = ""
pool_size = 10
min_idle_conns = 5
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"
idle_timeout = "5m"
idle_check_frequency = "1m"

# 哨兵配置
[client.redis.sentinel]
master_name = "mymaster"
sentinel_addresses = ["sentinel1:26379", "sentinel2:26379", "sentinel3:26379"]
password = ""
pool_size = 10
min_idle_conns = 5
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"
idle_timeout = "5m"
idle_check_frequency = "1m"
```

### 连接池

Jupiter 支持 Redis 连接池：

```go
// 创建连接池
client := redis.New(
    redis.WithConfig("default"),
    redis.WithPoolSize(10),
    redis.WithMinIdleConns(5),
    redis.WithMaxRetries(3),
)
```

### 事务

```go
// 使用事务
pipe := client.TxPipeline()
pipe.Incr("counter")
pipe.Expire("counter", time.Hour)
_, err := pipe.Exec()
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