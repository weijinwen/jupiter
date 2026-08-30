# Jupiter 核心模块

本文档介绍 Jupiter 框架的核心模块，包括应用管理、生命周期钩子、信号处理、单例模式、错误码和指标采集等。

## 概述

Jupiter 核心模块提供了微服务应用的基础能力，包括应用生命周期管理、优雅停机、信号处理、单例管理、错误码定义和指标采集等功能。

## 应用管理 (Application)

Application 是 Jupiter 框架的核心入口，负责管理服务器、Worker 和应用生命周期。

### 基本使用

```go
package main

import (
    "context"
    "github.com/weijinwen/jupiter/pkg/core/application"
)

func main() {
    // 创建应用实例
    app := application.DefaultApp()

    // 启动应用
    err := app.Startup()
    if err != nil {
        panic(err)
    }

    // 注册服务器
    // app.Run(server1, server2, ...)

    // 运行应用（阻塞直到收到退出信号）
    err = app.Run()
    if err != nil {
        panic(err)
    }
}
```

### 配置选项

```go
// 创建应用时设置选项
app, err := application.New(
    func() error {
        // 初始化逻辑
        return nil
    },
)

// 隐藏启动 Banner
app.HideBanner = true
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `New(fns ...func() error)` | 创建并初始化应用实例 |
| `DefaultApp()` | 创建默认应用实例 |
| `Startup(fns ...func() error)` | 启动应用，执行初始化函数 |
| `Serve(s ...server.Server)` | 注册服务器 |
| `Run(servers ...server.Server)` | 运行应用（阻塞） |
| `Stop()` | 立即停止应用 |
| `GracefulStop(ctx context.Context)` | 优雅停止应用 |
| `RegisterHooks(stage hooks.Stage, fns ...func())` | 注册生命周期钩子 |
| `RegisterWorker(name string, fn func(ctx context.Context) error)` | 注册 Worker |

### 应用选项

```go
// Disable 选项
type Disable int

const (
    DisableParserFlag Disable = iota  // 禁用命令行参数解析
)
```

## 生命周期钩子 (Hooks)

Jupiter 支持在应用生命周期的不同阶段执行钩子函数。

### 钩子阶段

```go
const (
    Stage_BeforeRun  hooks.Stage = "before_run"   // 应用启动前
    Stage_AfterStop  hooks.Stage = "after_stop"   // 应用停止后
    StageBeforeStop  hooks.Stage = "before_stop"  // 应用停止前（已废弃，使用 Stage_BeforeStop）
    StageAfterStop   hooks.Stage = "after_stop"   // 应用停止后（已废弃，使用 Stage_AfterStop）
)
```

### 注册钩子

```go
app := application.DefaultApp()

// 注册启动前钩子
app.RegisterHooks(hooks.Stage_BeforeRun, func() {
    log.Println("应用即将启动...")
})

// 注册停止后钩子
app.RegisterHooks(hooks.Stage_AfterStop, func() {
    log.Println("应用已停止")
})
```

## 信号处理 (Signals)

Jupiter 自动监听系统信号，实现优雅停机。

### 支持的信号

| 信号 | 行为 |
|------|------|
| `SIGINT` | 优雅停机 |
| `SIGTERM` | 优雅停机 |
| `SIGQUIT` | 立即停止 |

### 自定义信号处理

```go
import (
    "github.com/weijinwen/jupiter/pkg/core/signals"
)

// 自定义信号处理
signals.Shutdown(func(grace bool) {
    if grace {
        // 优雅停机
        log.Println("收到优雅停机信号")
    } else {
        // 立即停止
        log.Println("收到立即停止信号")
    }
})
```

## 单例管理 (Singleton)

Jupiter 提供了单例管理器，用于管理全局单例实例。

### 基本使用

```go
import (
    "github.com/weijinwen/jupiter/pkg/core/singleton"
)

// 存储单例
singleton.Store("module", "name", instance)

// 加载单例
if val, ok := singleton.Load("module", "name"); ok {
    instance := val.(*MyType)
    // 使用 instance
}
```

### 使用场景

- 数据库连接池
- Redis 客户端
- 配置管理器
- 日志实例

## 错误码 (ECode)

Jupiter 提供了统一的错误码管理。

### 模块定义

```go
// pkg/core/ecode/unified.go
const (
    ModApp           = "app"           // 应用模块
    ModClientRedis   = "client.redis"  // Redis 客户端
    ModeClientResty  = "client.resty"  // HTTP 客户端
    ModStoreGorm     = "store.gorm"    // GORM 存储
    ModStoreMongoDB  = "store.mongodb" // MongoDB 存储
)
```

## 指标采集 (Metric)

Jupiter 内置了 Prometheus 指标采集能力。

### 指标类型

#### Counter（计数器）

```go
import (
    "github.com/weijinwen/jupiter/pkg/core/metric"
)

// 创建 Counter
counter := metric.NewCounter(
    metric.WithNamespace("jupiter"),
    metric.WithSubsystem("http"),
    metric.WithName("requests_total"),
    metric.WithHelp("Total number of HTTP requests"),
    metric.WithLabels("method", "path"),
)

// 使用 Counter
counter.WithLabels("GET", "/api/users").Add(1)
```

#### Gauge（仪表盘）

```go
// 创建 Gauge
gauge := metric.NewGauge(
    metric.WithNamespace("jupiter"),
    metric.WithSubsystem("http"),
    metric.WithName("active_connections"),
    metric.WithHelp("Number of active connections"),
)

// 使用 Gauge
gauge.Set(100)
gauge.Inc()
gauge.Dec()
```

#### Histogram（直方图）

```go
// 创建 Histogram
histogram := metric.NewHistogram(
    metric.WithNamespace("jupiter"),
    metric.WithSubsystem("http"),
    metric.WithName("request_duration_seconds"),
    metric.WithHelp("HTTP request duration"),
    metric.WithBuckets(0.01, 0.05, 0.1, 0.5, 1, 5),
)

// 使用 Histogram
histogram.Observe(0.15)
```

#### Summary（摘要）

```go
// 创建 Summary
summary := metric.NewSummary(
    metric.WithNamespace("jupiter"),
    metric.WithSubsystem("http"),
    metric.WithName("request_duration_seconds"),
    metric.WithHelp("HTTP request duration"),
    metric.WithObjectives(0.5, 0.9, 0.99),
)

// 使用 Summary
summary.Observe(0.15)
```

## 自动进程管理 (AutoProc)

Jupiter 自动管理进程的 CPU 亲和性和 GOMAXPROCS 设置。

### 功能

- 自动设置 GOMAXPROCS
- 自动调整 CPU 亲和性
- 容器环境适配

### 使用

```go
import (
    _ "github.com/weijinwen/jupiter/pkg/core/autoproc" // 自动导入以启用
)
```

## 常量定义 (Constant)

Jupiter 定义了各种常量，用于配置键名、模块名等。

### 配置键名

```go
// pkg/core/constant/config.go
func ConfigKey(parts ...string) string {
    return "jupiter." + strings.Join(parts, ".")
}
```

### 模块名

```go
// pkg/core/constant/module.go
const (
    ModuleClientRedis   = "client.redis"
    ModuleClientResty   = "client.resty"
    ModuleStoreGorm     = "store.gorm"
    ModuleStoreMongoDB  = "store.mongodb"
    ModuleStoreMeili    = "store.meilisearch"
)
```

### 环境变量

```go
// pkg/core/constant/env.go
const (
    EnvRunMode   = "JUPITER_RUN_MODE"
    EnvAppMode   = "JUPITER_APP_MODE"
    EnvLogLevel  = "JUPITER_LOG_LEVEL"
)
```

## 最佳实践

### 1. 应用初始化顺序

```go
func main() {
    // 1. 创建应用
    app := application.DefaultApp()

    // 2. 注册钩子
    app.RegisterHooks(hooks.Stage_BeforeRun, func() {
        // 初始化依赖
    })

    // 3. 注册服务器
    app.Run(httpServer, grpcServer)
}
```

### 2. 优雅停机

```go
// 注册停止前钩子，释放资源
app.RegisterHooks(hooks.Stage_BeforeStop, func() {
    // 关闭数据库连接
    // 关闭 Redis 连接
    // 等待请求处理完成
})
```

### 3. 错误处理

```go
// 使用统一错误码
err := someOperation()
if err != nil {
    return xerror.Wrap(ecode.Internal, "操作失败", err)
}
```

## 下一步

- 了解 [配置系统](./config.md)
- 学习 [服务器组件](./servers.md)
- 探索 [客户端组件](./clients.md)
- 查看 [工具函数](./utils.md)
