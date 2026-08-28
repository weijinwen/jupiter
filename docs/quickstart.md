# 快速开始

本指南将帮助您快速上手 Jupiter 微服务框架，从安装到创建第一个应用。

## 环境要求

- Go 1.23.0 或更高版本
- Git

## 安装

### 安装 Jupiter

```bash
go get -u github.com/weijinwen/jupiter
```

### 安装开发工具

```bash
# 克隆项目
git clone https://github.com/weijinwen/jupiter.git
cd jupiter

# 安装开发工具
make init
```

## 创建第一个应用

### 1. 创建项目目录

```bash
mkdir my-jupiter-app && cd my-jupiter-app
```

### 2. 初始化 Go 模块

```bash
go mod init my-jupiter-app
```

### 3. 创建配置文件

创建 `config.toml` 文件：

```toml
[app]
name = "my-jupiter-app"
region = "default"
zone = "local"
env = "development"

[jupiter.server.http]
port = 8080
host = "0.0.0.0"
```

### 4. 创建主应用文件

创建 `main.go` 文件：

```go
package main

import (
    "github.com/weijinwen/jupiter"
    "github.com/weijinwen/jupiter/pkg/conf"
    "github.com/weijinwen/jupiter/pkg/server/xgin"
    "github.com/gin-gonic/gin"
)

func main() {
    // 加载配置
    if err := conf.LoadFromFile("config.toml"); err != nil {
        panic(err)
    }
    
    // 创建应用
    app := jupiter.New(jupiter.WithConfigFile("config.toml"))
    
    // 创建 Gin 服务器
    ginServer := xgin.Default()
    
    // 注册路由
    ginServer.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    ginServer.GET("/hello", func(c *gin.Context) {
        name := c.Query("name")
        if name == "" {
            name = "World"
        }
        c.JSON(200, gin.H{"message": "Hello, " + name + "!"})
    })
    
    // 将服务器注册到应用
    app.Serve(ginServer)
    
    // 启动应用
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### 5. 下载依赖

```bash
go mod tidy
```

### 6. 运行应用

```bash
go run main.go
```

### 7. 访问应用

打开浏览器或使用 curl 访问以下端点：

```bash
# 健康检查
curl http://localhost:8080/health

# 问候接口
curl http://localhost:8080/hello
curl http://localhost:8080/hello?name=Jupiter
```

## 下一步

现在您已经成功创建了第一个 Jupiter 应用，接下来可以：

- 学习 [核心概念](./concepts.md)
- 了解 [配置系统](./config.md)
- 探索 [服务器组件](./servers.md)
- 尝试 [客户端组件](./clients.md)
- 查看 [存储组件](./stores.md)

## 常见问题

### Q: 如何修改服务器端口？

A: 在 `config.toml` 文件中修改 `jupiter.server.http.port` 的值：

```toml
[jupiter.server.http]
port = 9090  # 修改为您需要的端口
```

### Q: 如何添加数据库支持？

A: 首先在 `config.toml` 中添加数据库配置：

```toml
[jupiter.mysql.default]
driverName = "mysql"
dsn = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

然后在代码中使用：

```go
import "github.com/weijinwen/jupiter/pkg/store/gorm"

// 获取数据库连接
db := gorm.StdConfig("default").MustOpen()
```

### Q: 如何添加 Redis 支持？

A: 首先在 `config.toml` 中添加 Redis 配置：

```toml
[jupiter.redis.default]
addr = "localhost:6379"
password = ""
db = 0
```

然后在代码中使用：

```go
import "github.com/weijinwen/jupiter/pkg/client/redis"

// 获取 Redis 客户端
client := redis.StdConfig("default").MustSingleton()
```

### Q: 如何添加 Kafka 支持？

A: 首先在 `config.toml` 中添加 Kafka 配置：

```toml
[client.kafka.default]
address = "localhost:9092"
client_id = "jupiter-kafka-client"
timeout = "30s"
compression_type = "none"
batch_size = 16384
batch_timeout = "10ms"
max_retries = 3
retry_backoff = "100ms"
```

然后在代码中使用：

```go
import "github.com/weijinwen/jupiter/pkg/client/kafka"

// 获取 Kafka 客户端
client := kafka.StdConfig("default").MustSingleton()

// 发送消息
msg, err := client.SendMessage("test-topic", []byte("key"), []byte("value"))
if err != nil {
    panic(err)
}
fmt.Printf("发送成功: topic=%s, partition=%d, offset=%d\n", msg.Topic, msg.Partition, msg.Offset)
```

### Q: 如何添加日志？

A: Jupiter 默认已经集成了日志组件，您可以直接使用：

```go
import "github.com/weijinwen/jupiter/pkg/xlog"

// 获取默认日志器
logger := xlog.Jupiter()

// 记录日志
logger.Info("应用启动成功")
logger.Error("发生错误", xlog.FieldErr(err))
```

## 更多资源

- [探索客户端组件](./clients.md)
- [学习存储组件](./stores.md)