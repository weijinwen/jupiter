# Worker 命令执行模块

本文档介绍 Jupiter 框架的 Worker 模块，用于执行一次性命令行任务。

## 概述

Worker 模块采用**命令模式（Command Pattern）**设计，允许开发者注册可执行的命令，并通过命令行参数触发执行。主要用于数据迁移、定时任务、数据清理等一次性任务场景。

## 核心概念

### Command 接口

```go
type Command interface {
    Name() string                    // 命令名称
    Run(ctx context.Context) error   // 执行命令
}
```

### 命令注册表

Worker 使用全局注册表管理命令，支持线程安全的注册和获取。

## 使用方法

### 1. 注册函数命令

```go
package main

import (
    "context"
    "fmt"

    "github.com/weijinwen/jupiter"
    "github.com/weijinwen/jupiter/pkg/worker"
)

func main() {
    app := jupiter.New()

    // 注册命令
    worker.RegisterFunc("send-email", func(ctx context.Context) error {
        fmt.Println("发送邮件...")
        // 执行邮件发送逻辑
        return nil
    })

    app.Run()
}
```

### 2. 实现自定义命令

```go
type DataSyncCommand struct {
    sourceDB *gorm.DB
    targetDB *gorm.DB
}

func (c *DataSyncCommand) Name() string {
    return "data-sync"
}

func (c *DataSyncCommand) Run(ctx context.Context) error {
    // 执行数据同步
    return c.syncData(ctx)
}

// 注册
worker.Register(&DataSyncCommand{
    sourceDB: db1,
    targetDB: db2,
})
```

### 3. 命令行执行

```bash
# 执行命令（默认30分钟超时）
go run main.go --cmd=send-email

# 指定超时时间
go run main.go --cmd=send-email --cmd-timeout=5m

# 支持的格式：ns, us, ms, s, m, h
go run main.go --cmd=send-email --cmd-timeout=300s
```

## 增强功能

### 超时控制

使用 `WithTimeout` 包装命令：

```go
cmd := &DataSyncCommand{}
timeoutCmd := worker.WithTimeout(cmd, 10*time.Minute)
worker.Register(timeoutCmd)
```

### 命令链

多个命令按顺序执行：

```go
// 创建多个命令
cmd1 := worker.CommandFunc(func(ctx context.Context) error {
    fmt.Println("步骤1: 备份数据")
    return nil
})

cmd2 := worker.CommandFunc(func(ctx context.Context) error {
    fmt.Println("步骤2: 执行迁移")
    return nil
})

cmd3 := worker.CommandFunc(func(ctx context.Context) error {
    fmt.Println("步骤3: 验证结果")
    return nil
})

// 创建命令链
chain := worker.Chain(cmd1, cmd2, cmd3)
worker.Register(chain)
```

执行命令链：
```bash
go run main.go --cmd=chain
```

### 可中断的命令链

使用 `ChainWithStop` 创建可在中途停止的链：

```go
chain := worker.ChainWithStop(
    worker.CommandFunc(step1),
    worker.CommandFunc(step2),
    worker.CommandFunc(step3),
)
worker.Register(chain)
```

在命令中返回 `worker.ErrStopChain` 可提前终止链式执行（不报错）：

```go
func step2(ctx context.Context) error {
    // 检查是否需要停止
    if shouldStop() {
        return worker.ErrStopChain  // 正常停止，不报错
    }
    return nil
}
```

## 配置选项

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--cmd` | 要执行的命令名称 | - |
| `--cmd-timeout` | 命令执行超时时间 | 30m |

### 示例配置

```toml
# config.toml 中不需要特殊配置
# 通过命令行参数控制
```

## 典型应用场景

### 1. 数据迁移

```go
func main() {
    app := jupiter.New()

    app.RegisterWorker("migrate-users", func(ctx context.Context) error {
        // 从旧表迁移用户数据到新表
        return migrateUsers(ctx)
    })

    app.Run()
}
```

### 2. 定时任务替代方案

```go
func main() {
    app := jupiter.New()

    // 注册清理任务
    app.RegisterWorker("cleanup-logs", func(ctx context.Context) error {
        // 清理7天前的日志
        cutoff := time.Now().AddDate(0, 0, -7)
        return deleteOldLogs(ctx, cutoff)
    })

    // 注册报表生成任务
    app.RegisterWorker("generate-report", func(ctx context.Context) error {
        return generateDailyReport(ctx)
    })

    app.Run()
}
```

通过 cron 定时调用：
```bash
# 添加到 crontab
0 2 * * * cd /app && ./myapp --cmd=cleanup-logs
0 6 * * * cd /app && ./myapp --cmd=generate-report
```

### 3. 数据库维护

```go
app.RegisterWorker("optimize-db", func(ctx context.Context) error {
    // 执行数据库优化
    return db.Exec("OPTIMIZE TABLE users").Error
})
```

## 注意事项

1. **命令执行完成后会退出程序**，仅用于一次性任务
2. **未指定 `--cmd` 时正常运行应用**，不会触发命令执行
3. **命令名称唯一**，后注册的同名命令会覆盖之前的
4. **超时时间从配置加载后开始计算**

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/weijinwen/jupiter"
    "github.com/weijinwen/jupiter/pkg/worker"
)

func main() {
    app := jupiter.New()

    // 注册数据导出命令
    worker.RegisterFunc("export-data", func(ctx context.Context) error {
        fmt.Println("开始导出数据...")

        // 检查上下文是否已取消
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // 模拟导出操作
        time.Sleep(2 * time.Second)

        fmt.Println("数据导出完成")
        return nil
    })

    // 注册带超时的命令
    heavyCmd := worker.CommandFunc(func(ctx context.Context) error {
        fmt.Println("执行耗时操作...")

        // 模拟长时间运行
        for i := 0; i < 10; i++ {
            select {
            case <-ctx.Done():
                return fmt.Errorf("操作超时或取消")
            case <-time.After(1 * time.Second):
                fmt.Printf("进度: %d/10\n", i+1)
            }
        }

        return nil
    })

    worker.Register(worker.WithTimeout(heavyCmd, 5*time.Second))

    app.Run()
}
```

运行方式：
```bash
# 正常启动应用
go run main.go

# 执行导出命令（默认30分钟超时）
go run main.go --cmd=export-data

# 执行带超时的命令
go run main.go --cmd=anonymous --cmd-timeout=10s
```

## 下一步

- 了解 [工具函数](./utils.md)
- 学习 [日志组件](./xlog.md)
- 探索 [配置系统](./config.md)
- 查看 [快速开始](./quickstart.md)
