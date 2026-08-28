# XLog 日志组件

XLog 是 Jupiter 框架的日志组件，基于 [Zap](https://github.com/uber-go/zap) 日志库封装，提供了高性能的结构化日志功能，同时简化了使用难度。

## 功能特性

- 基于 Zap 的高性能结构化日志
- 支持多种日志级别（Debug、Info、Warn、Error、DPanic、Panic、Fatal）
- 支持日志文件自动轮转
- 支持异步日志写入
- 支持动态调整日志级别
- 支持上下文传递日志
- 内置指标采集
- 支持多日志实例

## 配置说明

### 基本配置

XLog 组件的配置项如下：

```toml
[jupiter.logger.default]
# 日志输出目录
dir = "/tmp/log"
# 日志文件名称
name = "app.json"
# 日志级别: debug, info, warn, error, dpanic, panic, fatal
level = "info"
# 是否添加调用者信息
addCaller = true
# 调用者跳过层级
callerSkip = 0
# 日志文件最大大小（MB），超过则轮转
maxSize = 500
# 日志文件最大保存天数
maxAge = 7
# 最多保留的备份文件数
maxBackup = 10
# 日志刷盘间隔
interval = "24h"
# 是否启用异步写入
async = true
# 是否启用队列模式
queue = false
# 队列休眠时间
queueSleep = "100ms"
# 是否启用调试模式（输出到控制台）
debug = false
```

### 框架日志配置

```toml
[jupiter.logger.jupiter]
# 框架日志配置，用于记录 Jupiter 框架内部日志
name = "jupiter.sys"
level = "info"
```

### 完整配置示例

```toml
[jupiter.logger.default]
dir = "/var/log/myapp"
name = "application.json"
level = "info"
addCaller = true
maxSize = 500
maxAge = 7
maxBackup = 10
interval = "24h"
async = true
debug = false

[jupiter.logger.jupiter]
name = "jupiter.sys"
level = "warn"
```

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/weijinwen/jupiter/pkg/xlog"
)

func main() {
    // 使用默认日志
    xlog.Info("application started", xlog.String("version", "1.0.0"))
    xlog.Debug("debug message", xlog.Int("count", 100))
    xlog.Error("error occurred", xlog.FieldErr(err))
}
```

### 创建自定义日志

```go
package main

import (
    "github.com/weijinwen/jupiter/pkg/xlog"
)

func main() {
    // 使用配置文件创建日志
    logger := xlog.StdConfig("jupiter", "mylog").Build()
    logger.Info("custom logger", xlog.String("module", "user"))
    
    // 使用代码配置创建日志
    config := xlog.Config{
        Name:  "custom.log",
        Dir:   "/tmp/logs",
        Level: "debug",
    }
    logger := config.Build()
    logger.Debug("debug message")
}
```

## 日志级别

XLog 支持以下日志级别，按严重程度从低到高排列：

| 级别 | 说明 |
|------|------|
| `debug` | 调试信息，最详细的日志 |
| `info` | 普通信息，记录正常流程 |
| `warn` | 警告信息，可能存在问题 |
| `error` | 错误信息，功能异常 |
| `dpanic` | 开发环境 panic，生产环境仅记录 |
| `panic` | 严重错误，记录后 panic |
| `fatal` | 致命错误，记录后退出程序 |

### 动态设置日志级别

```toml
[jupiter.logger.default]
level = "error"
```

## 日志方法

### 标准日志方法

```go
// 使用字段
xlog.Debug("message", xlog.String("key", "value"))
xlog.Info("message", xlog.Int("count", 100))
xlog.Warn("message", xlog.FieldErr(err))
xlog.Error("message", xlog.Any("data", obj))

// 使用格式化
xlog.Debugf("debug %s", "message")
xlog.Infof("info %s", "message")
xlog.Warnf("warn %s", "message")
xlog.Errorf("error %s", "message")

// 使用键值对
xlog.Debugw("debug", "key", "value")
xlog.Infow("info", "key", "value")
xlog.Warnw("warn", "key", "value")
xlog.Errorw("error", "key", "value")
```

### 带上下文的日志

```go
package main

import (
    "context"
    "github.com/weijinwen/jupiter/pkg/xlog"
)

func main() {
    ctx := context.Background()
    
    // 获取标准日志
    logger := xlog.L(ctx)
    logger.Info("standard log")
    
    // 获取 Jupiter 框架日志
    jupiterLogger := xlog.J(ctx)
    jupiterLogger.Info("jupiter log")
}
```

### 创建子日志

```go
// 添加固定字段
logger := xlog.With(xlog.String("service", "user"), xlog.String("version", "1.0.0"))
logger.Info("message") // 会自动包含 service 和 version 字段

// 添加选项
logger := xlog.WithOptions(zap.AddCallerSkip(1))

// 命名日志
logger := xlog.Named("user.service")
logger.Info("message") // logger 字段会显示 user.service
```

## 预定义字段

XLog 提供了丰富的预定义字段函数：

### 通用字段

```go
xlog.String("key", "value")      // 字符串字段
xlog.Int("key", 100)             // 整数字段
xlog.Int64("key", 1000)          // int64 字段
xlog.Int32("key", 100)           // int32 字段
xlog.Uint("key", 100)            // uint 字段
xlog.Any("key", value)           // 任意类型字段
xlog.Duration("key", duration)   // 时间间隔字段
xlog.Reflect("key", obj)         // 反射字段
xlog.ByteString("key", []byte)   // 字节数组字段
```

### 业务字段

```go
// 应用标识
xlog.FieldAid("app-123")         // aid: 应用ID
xlog.FieldMod("user.service")    // mod: 模块名
xlog.FieldName("operation")      // name: 操作名

// 地址信息
xlog.FieldAddr("127.0.0.1:8080") // addr: 地址
xlog.FieldHost("localhost")      // host: 主机名
xlog.FieldIP("192.168.1.1")      // ip: IP地址

// 请求信息
xlog.FieldMethod("GET")          // meth: 方法
xlog.FieldCode(200)              // code: 状态码
xlog.FieldCost(time.Second)      // cost: 耗时（毫秒）
xlog.FieldReqAID("req-123")      // reqAid: 请求AID
xlog.FieldReqHost("api.example.com") // reqHost: 请求主机

// 错误信息
xlog.FieldErr(err)               // error: 错误对象
xlog.FieldStringErr("error msg") // err: 错误字符串
xlog.FieldErrKind("timeout")     // errKind: 错误类型

// 数据字段
xlog.FieldKey("user_id")         // key: 键
xlog.FieldValue("12345")         // value: 值
xlog.FieldExtMessage(obj)        // ext: 扩展信息
xlog.FieldStack(stack)           // stack: 堆栈信息

// 事件字段
xlog.FieldEvent("login")         // event: 事件
xlog.FieldColor("red")           // color: 颜色
xlog.FieldStdMethod("POST")      // stdMeth: 标准方法

// 消息队列字段
xlog.FieldTID("topic-1")         // tid: Topic ID
xlog.FieldTopic("orders")        // topic: Topic名称
xlog.FieldPartition(0)           // partition: 分区
xlog.FieldOffset(1000)           // offset: 偏移量
```

## 上下文传递

### 创建带日志的上下文

```go
package main

import (
    "context"
    "github.com/weijinwen/jupiter/pkg/xlog"
)

func main() {
    // 创建带日志的上下文
    logger := xlog.Default()
    ctx := xlog.NewContext(context.Background(), logger, "trace-123")
    
    // 在函数间传递上下文
    processRequest(ctx)
}

func processRequest(ctx context.Context) {
    // 从上下文获取日志
    logger := xlog.L(ctx)
    logger.Info("processing request") // 会自动包含 trace_id 字段
}
```

## 日志轮转

XLog 内置了日志轮转功能，支持按文件大小和日期轮转：

### 轮转配置

```toml
[jupiter.logger.default]
# 单个日志文件最大大小（MB）
maxSize = 500
# 日志文件保留天数
maxAge = 7
# 最多保留的备份文件数
maxBackup = 10
# 日志刷盘间隔
interval = "24h"
```

### 轮转规则

- 当单个日志文件达到 `maxSize` 时，自动创建新文件
- 超过 `maxAge` 天的日志文件会被自动删除
- 备份文件数量不超过 `maxBackup`
- 支持按时间间隔自动刷盘

## 异步日志

XLog 支持异步日志写入，提高日志性能：

```toml
[jupiter.logger.default]
# 启用异步写入
async = true
```

异步日志会将日志先写入内存缓冲区，然后定期刷盘。在应用关闭时，会自动刷新缓冲区。

## 调试模式

在开发环境中，可以启用调试模式将日志输出到控制台：

```toml
[jupiter.logger.default]
debug = true
```

或者在代码中：

```go
config := xlog.Config{
    Name:  "app",
    Debug: true,
}
logger := config.Build()
```

调试模式下，日志会以彩色格式输出到控制台，便于开发调试。

## 多日志实例

XLog 支持创建多个独立的日志实例：

```go
// 业务日志
bizConfig := xlog.Config{
    Name:  "business.json",
    Dir:   "/var/log/myapp",
    Level: "info",
}
bizLogger := bizConfig.Build()

// 访问日志
accessConfig := xlog.Config{
    Name:  "access.json",
    Dir:   "/var/log/myapp",
    Level: "info",
}
accessLogger := accessConfig.Build()

// 错误日志
errorConfig := xlog.Config{
    Name:  "error.json",
    Dir:   "/var/log/myapp",
    Level: "error",
}
errorLogger := errorConfig.Build()
```

## 指标采集

XLog 内置了日志指标采集功能，会自动记录各级别日志的数量：

```go
// 指标名称: log_level_total
// 标签: name（日志名称）, level（日志级别）
```

可以通过 Prometheus 采集这些指标，用于监控日志输出情况。

## 最佳实践

### 1. 合理使用日志级别

```go
// Debug：开发调试使用，生产环境可以关闭
xlog.Debug("detailed debug info", xlog.Any("data", data))

// Info：记录正常业务流程
xlog.Info("user login", xlog.String("user_id", userID))

// Warn：记录可能的问题
xlog.Warn("slow query", xlog.FieldCost(duration), xlog.String("sql", sql))

// Error：记录错误信息
xlog.Error("database connection failed", xlog.FieldErr(err))
```

### 2. 使用结构化字段

```go
// 推荐：使用结构化字段
xlog.Info("user login",
    xlog.String("user_id", userID),
    xlog.String("ip", clientIP),
    xlog.Int64("timestamp", time.Now().Unix()),
)

// 不推荐：使用格式化字符串
xlog.Infof("user %s login from %s at %d", userID, clientIP, time.Now().Unix())
```

### 3. 错误日志记录

```go
// 推荐：同时记录错误和上下文
xlog.Error("operation failed",
    xlog.FieldErr(err),
    xlog.String("operation", "create_order"),
    xlog.String("order_id", orderID),
)

// 不推荐：只记录错误信息
xlog.Error(err.Error())
```

### 4. 上下文传递

```go
func Handler(ctx context.Context, req *Request) {
    // 将请求信息添加到上下文
    ctx = xlog.NewContext(ctx, xlog.Default(), req.TraceID)
    
    // 传递上下文
    process(ctx, req)
}

func process(ctx context.Context, req *Request) {
    // 使用上下文中的日志，自动包含 trace_id
    logger := xlog.L(ctx)
    logger.Info("processing request", xlog.String("user_id", req.UserID))
}
```

### 5. 日志配置建议

```toml
# 生产环境配置
[jupiter.logger.default]
level = "info"        # 只记录 info 及以上级别
async = true          # 启用异步写入
maxSize = 500         # 单文件最大 500MB
maxAge = 7            # 保留 7 天
maxBackup = 10        # 最多 10 个备份

# 开发环境配置
[jupiter.logger.default]
level = "debug"       # 记录所有级别
debug = true          # 输出到控制台
async = false         # 同步写入，便于调试
```

## 注意事项

1. **日志级别设置**：生产环境建议设置为 `info` 或更高级别，避免产生过多日志
2. **异步日志**：启用异步日志可以提高性能，但在应用崩溃时可能丢失部分日志
3. **日志轮转**：合理设置 `maxSize` 和 `maxAge`，避免磁盘空间不足
4. **敏感信息**：不要在日志中记录密码、密钥等敏感信息
5. **错误处理**：使用 `xlog.FieldErr(err)` 记录错误时，确保错误不为 nil

## 下一步

- 了解 [配置系统](./config.md)
- 探索 [工具函数](./utils.md)
- 查看 [MeiliSearch 存储](./meilisearch.md)
- 学习 [Worker 命令执行](./worker.md)
