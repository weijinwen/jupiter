# 我的 Jupiter 微服务项目

![License](https://img.shields.io/badge/license-Apache%202-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.25.0+-00ADD8?style=flat&logo=go)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)

这是基于 Jupiter 企业级微服务框架构建的个人项目，提供了完整的微服务开发解决方案，包括服务器、配置管理、日志、工作队列等核心组件。

## 项目特性

- **微服务架构支持**：提供完整的微服务开发工具链
- **多服务器支持**：内置 Gin（`xgin`）、Hertz（`xhertz`）HTTP 服务器与治理服务器（`governor`）
- **配置管理**：支持 TOML/YAML/JSON 多格式，通过 `file://`、`http://` 数据源加载，统一 `StdConfig(name)` 约定键名
- **监控指标**：集成 Prometheus 指标收集，governor 内置 `/metrics` 端点
- **日志管理**：基于 zap 的结构化日志系统
- **缓存支持**：内置 FreeCache、Golang LRU 多种缓存实现
- **数据库支持**：支持 MySQL、PostgreSQL（GORM）、MongoDB（`mongox`）、MeiliSearch 等多种数据存储
- **客户端支持**：内置 Redis 客户端与基于 Resty 的 HTTP 客户端
- **WebSocket 支持**：`xgin`/`xhertz` 均提供 WebSocket 升级能力
- **优雅停止**：基于信号的优雅停止（`GracefulStop`），支持停止排空超时配置
- **TLS/HTTPS**：HTTP 服务器支持配置证书启用 HTTPS

## 快速开始

### 环境要求

- Go 1.25.0 或更高版本
- Git

### 安装

```bash
go get -u github.com/weijinwen/jupiter
```

### 创建项目

```bash
# 创建项目目录
mkdir my-jupiter-app && cd my-jupiter-app

# 初始化 Go 模块
go mod init my-jupiter-app

# 创建 main.go 文件
cat > main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"

    "github.com/weijinwen/jupiter"
    "github.com/weijinwen/jupiter/pkg/server/xgin"
)

func main() {
    app, err := jupiter.New()
    if err != nil {
        panic(err)
    }

    // 创建 Gin 服务器（默认 host 取 --host flag，port 9091，release 模式）
    ginServer := xgin.DefaultConfig().
        WithHost("0.0.0.0").
        WithPort(8080).
        MustBuild()

    // 注册路由
    ginServer.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 将服务器注册到应用
    app.Serve(ginServer)

    // 启动应用（收到 SIGINT/SIGTERM 时优雅停止）
    if err := app.Run(); err != nil {
        panic(err)
    }
}
EOF

# 下载依赖
go mod tidy

# 运行应用
go run main.go
```

### 访问应用

打开浏览器访问 http://localhost:8080/health，你将看到返回的 JSON 响应。

## 项目文档

- [文档总览](docs/README.md)
- [快速开始](docs/quickstart.md)
- [核心概念](docs/concepts.md)
- [配置系统](docs/config.md)
- [服务器组件](docs/servers.md)
- [客户端组件](docs/clients.md)
- [存储组件](docs/stores.md)
- [缓存组件](docs/caches.md)
- [日志组件](docs/xlog.md)
- [MeiliSearch 存储](docs/meilisearch.md)
- [Worker 命令执行](docs/worker.md)
- [工具函数](docs/utils.md)

## 组件配置控制功能

本项目基于 Jupiter 框架的统一组件配置控制功能，允许通过配置文件控制组件的行为和功能开关。

### 核心特性

- **统一命名规范**：所有组件的功能开关采用一致的命名规范（`enableMetric`、`enableAccessLog`、`slowThreshold`、`debug` 等）
- **默认值策略**：每个功能开关都有合理的默认值，开箱即用
- **灵活配置**：支持通过配置文件和代码两种方式进行配置

### 支持的组件

- **Redis 客户端**：支持调试模式、指标收集、访问日志和慢查询监控
- **GORM 存储组件**：支持调试模式、指标收集、数据库日志和慢查询监控
- **MeiliSearch 存储组件**：支持调试模式、指标收集、访问日志和慢查询监控
- **Resty HTTP 客户端**：支持调试模式、指标收集、访问日志和慢查询监控

## 项目结构

```
jupiter/
├── pkg/                    # 核心包
│   ├── cache/             # 缓存实现
│   │   ├── xfreecache/    # FreeCache 缓存
│   │   └── xgolanglru/    # GolangLRU 缓存
│   ├── client/            # 客户端实现
│   │   ├── redis/         # Redis 客户端
│   │   └── resty/         # HTTP 客户端
│   ├── conf/              # 配置管理
│   ├── core/              # 核心组件（应用生命周期、hook、指标、单例）
│   ├── flag/              # 命令行 flag（--config、--version 等）
│   ├── server/            # 服务器实现
│   │   ├── governor/      # 治理服务器（/metrics、健康检查）
│   │   ├── xgin/          # Gin 服务器
│   │   └── xhertz/        # Hertz 服务器
│   ├── store/             # 存储实现
│   │   ├── gorm/          # GORM 数据库 ORM
│   │   ├── meilisearch/   # MeiliSearch 搜索引擎
│   │   └── mongox/        # MongoDB 扩展
│   ├── util/              # 通用工具
│   ├── worker/            # Worker 命令注册
│   └── xlog/              # 日志组件
├── docs/                  # 文档
└── examples/              # 示例代码（xhertz）
```