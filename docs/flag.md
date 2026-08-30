# Jupiter 命令行参数

本文档介绍 Jupiter 框架的命令行参数解析模块，包括参数注册、解析和使用方法。

## 概述

Jupiter 提供了强大的命令行参数解析功能，支持多种参数类型，支持环境变量绑定，支持参数验证和钩子函数。

## 基本使用

### 注册参数

```go
package main

import (
    "github.com/weijinwen/jupiter/pkg/flag"
)

func main() {
    // 注册参数
    flag.Register(
        &flag.StringFlag{
            Name:    "config",
            Usage:   "配置文件路径",
            EnvVar:  "APP_CONFIG",
            Default: "config.toml",
        },
        &flag.BoolFlag{
            Name:    "debug",
            Usage:   "调试模式",
            EnvVar:  "APP_DEBUG",
            Default: false,
        },
        &flag.IntFlag{
            Name:    "port",
            Usage:   "服务端口",
            Default: 8080,
        },
    )

    // 解析参数
    err := flag.Parse()
    if err != nil {
        panic(err)
    }

    // 获取参数值
    config := flag.String("config")
    debug := flag.Bool("debug")
    port := flag.Int("port")
}
```

### 使用默认参数

Jupiter 内置了一些常用的默认参数：

```go
// 内置参数
flag.String("host")      // 服务主机地址
flag.String("config")    // 配置文件路径
flag.Bool("debug")       // 调试模式
```

## 参数类型

### StringFlag（字符串参数）

```go
&flag.StringFlag{
    Name:     "name",           // 参数名（支持逗号分隔多个别名）
    Usage:    "用户名称",        // 使用说明
    EnvVar:   "APP_NAME",       // 环境变量名
    Default:  "default",        // 默认值
    Variable: &name,            // 绑定变量（可选）
    Action: func(name string, fs *flag.FlagSet) {
        // 解析后的回调函数
    },
}
```

### BoolFlag（布尔参数）

```go
&flag.BoolFlag{
    Name:     "verbose",
    Usage:    "详细输出",
    EnvVar:   "APP_VERBOSE",
    Default:  false,
    Variable: &verbose,
    Action: func(name string, fs *flag.FlagSet) {
        // 回调函数
    },
}
```

### IntFlag（整数参数）

```go
&flag.IntFlag{
    Name:     "port",
    Usage:    "服务端口",
    Default:  8080,
    Variable: &port,
    Action: func(name string, fs *flag.FlagSet) {
        // 回调函数
    },
}
```

### UintFlag（无符号整数参数）

```go
&flag.UintFlag{
    Name:     "max-conn",
    Usage:    "最大连接数",
    Default:  100,
    Variable: &maxConn,
}
```

### Float64Flag（浮点数参数）

```go
&flag.Float64Flag{
    Name:     "ratio",
    Usage:    "比率",
    Default:  1.0,
    Variable: &ratio,
}
```

## 获取参数值

### 带错误返回的获取方式

```go
// 获取字符串参数
name, err := flag.StringE("name")
if err != nil {
    log.Printf("参数 %s 未定义", "name")
}

// 获取布尔参数
debug, err := flag.BoolE("debug")
if err != nil {
    log.Printf("参数 %s 未定义", "debug")
}

// 获取整数参数
port, err := flag.IntE("port")
if err != nil {
    log.Printf("参数 %s 未定义", "port")
}

// 获取无符号整数参数
maxConn, err := flag.UintE("max-conn")
if err != nil {
    log.Printf("参数 %s 未定义", "max-conn")
}

// 获取浮点数参数
ratio, err := flag.Float64E("ratio")
if err != nil {
    log.Printf("参数 %s 未定义", "ratio")
}
```

### 简化获取方式（忽略错误）

```go
name := flag.String("name")
debug := flag.Bool("debug")
port := flag.Int("port")
maxConn := flag.Uint("max-conn")
ratio := flag.Float64("ratio")
```

## 环境变量绑定

Jupiter 支持将参数与环境变量绑定，优先级为：命令行参数 > 默认值 > 环境变量。

### 使用方法

```go
&flag.StringFlag{
    Name:    "database-url",
    Usage:   "数据库连接地址",
    EnvVar:  "DATABASE_URL",  // 环境变量名
    Default: "localhost:5432",
}
```

### 优先级说明

```bash
# 1. 命令行参数优先级最高
./app --database-url="remote-db:5432"

# 2. 其次是默认值
# Default: "localhost:5432"

# 3. 最后是环境变量
export DATABASE_URL="env-db:5432"
./app
```

## 参数别名

支持为参数定义多个别名，使用逗号分隔：

```go
&flag.StringFlag{
    Name:    "config, c",  // 支持 --config 或 -c
    Usage:   "配置文件路径",
    Default: "config.toml",
}
```

## 回调函数 (Action)

在参数解析完成后执行回调函数：

```go
&flag.StringFlag{
    Name:   "log-level",
    Usage:  "日志级别",
    Default: "info",
    Action: func(name string, fs *flag.FlagSet) {
        // 参数解析完成后执行
        level := fs.String("log-level")
        xlog.SetLevel(level)
    },
}
```

## 自定义 FlagSet

### 创建自定义 FlagSet

```go
import (
    "github.com/weijinwen/jupiter/pkg/flag"
)

// 创建自定义 FlagSet
fs := &flag.FlagSet{}

// 注册参数
fs.Register(
    &flag.StringFlag{
        Name:    "name",
        Usage:   "名称",
        Default: "default",
    },
)

// 解析参数
err := fs.Parse()
if err != nil {
    panic(err)
}

// 获取参数值
name := fs.String("name")
```

## 在 Application 中使用

Jupiter Application 自动集成命令行参数解析：

```go
package main

import (
    "github.com/weijinwen/jupiter/pkg/core/application"
    "github.com/weijinwen/jupiter/pkg/flag"
)

func main() {
    // 注册自定义参数
    flag.Register(
        &flag.StringFlag{
            Name:    "app-name",
            Usage:   "应用名称",
            Default: "my-app",
        },
    )

    // 创建应用（自动解析参数）
    app := application.DefaultApp()

    // 运行应用
    app.Run()
}
```

### 禁用参数解析

```go
app := application.DefaultApp()

// 禁用参数解析
app.disableMap[application.DisableParserFlag] = true
```

## 最佳实践

### 1. 参数命名规范

```go
// 使用 kebab-case 命名
&flag.StringFlag{
    Name:    "log-level",      // 推荐
    // Name:   "logLevel",      // 不推荐
    // Name:   "log_level",     // 不推荐
}
```

### 2. 提供有意义的 Usage

```go
&flag.StringFlag{
    Name:  "config",
    Usage: "配置文件路径，支持 TOML 格式",  // 清晰说明
}
```

### 3. 合理使用环境变量

```go
// 敏感信息使用环境变量
&flag.StringFlag{
    Name:   "db-password",
    Usage:  "数据库密码",
    EnvVar: "DB_PASSWORD",  // 敏感信息不设默认值
}
```

### 4. 参数验证

```go
&flag.IntFlag{
    Name:    "port",
    Usage:   "服务端口",
    Default: 8080,
    Action: func(name string, fs *flag.FlagSet) {
        port := fs.Int("port")
        if port < 1 || port > 65535 {
            log.Fatalf("端口号无效: %d", port)
        }
    },
}
```

## 示例

### 完整示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/weijinwen/jupiter/pkg/flag"
)

func main() {
    // 注册参数
    flag.Register(
        &flag.StringFlag{
            Name:    "config, c",
            Usage:   "配置文件路径",
            EnvVar:  "APP_CONFIG",
            Default: "config.toml",
        },
        &flag.BoolFlag{
            Name:    "debug, d",
            Usage:   "启用调试模式",
            EnvVar:  "APP_DEBUG",
            Default: false,
        },
        &flag.IntFlag{
            Name:    "port, p",
            Usage:   "HTTP 服务端口",
            Default: 8080,
            Action: func(name string, fs *flag.FlagSet) {
                port := fs.Int("port")
                if port < 1 || port > 65535 {
                    log.Fatalf("无效的端口号: %d", port)
                }
            },
        },
        &flag.StringFlag{
            Name:   "log-level",
            Usage:  "日志级别 (debug, info, warn, error)",
            EnvVar: "APP_LOG_LEVEL",
            Default: "info",
        },
    )

    // 解析参数
    err := flag.Parse()
    if err != nil {
        log.Fatalf("解析参数失败: %v", err)
    }

    // 使用参数
    config := flag.String("config")
    debug := flag.Bool("debug")
    port := flag.Int("port")
    logLevel := flag.String("log-level")

    fmt.Printf("配置文件: %s\n", config)
    fmt.Printf("调试模式: %v\n", debug)
    fmt.Printf("服务端口: %d\n", port)
    fmt.Printf("日志级别: %s\n", logLevel)
}
```

## 下一步

- 了解 [配置系统](./config.md)
- 学习 [核心模块](./core.md)
- 探索 [服务器组件](./servers.md)
