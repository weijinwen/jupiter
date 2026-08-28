# CLAUDE.md

This file provides guidance to AI coding agents when working with code in this repository.

## 项目定位

Jupiter 是一个 Go 微服务框架库，模块路径为 `github.com/weijinwen/jupiter`。根包只暴露少量入口别名，主要实现位于 `pkg/` 下：应用生命周期、配置加载、服务器抽象、客户端、存储、缓存、日志和 worker 命令注册。

当前 `go.mod` 声明 `go 1.25.0`。README 和部分文档中的 Go 版本或示例可能滞后，改动前以源码和 `go.mod` 为准。

## 常用命令

### 测试

```bash
# 运行全部测试
make unit-test
# 等价核心命令：go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# 快速运行全部测试（无 race/覆盖率）
go test ./...

# 运行单个包
go test ./pkg/conf

# 运行单个测试函数
go test ./pkg/conf -run TestLoadFromReader -v

# 查看 unit-test 生成的覆盖率报告
make covsh-unit
```

### 代码检查与生成

```bash
# Go lint
golangci-lint run -v
# 或 make golint

# Protobuf lint（不是 Go lint）
make lint

# 生成 Protobuf 代码
make generate

# Protobuf 兼容性检查
make breaking

# 测试 proto 到 HTTP 相关代码
make test-proto2http
```

### 文档与 OpenAPI

```bash
# 校验 OpenAPI 文档
make validate

# 启动 OpenAPI 文档服务
make serve
```

### 注意事项

- `make help` 是默认目标，可查看 Makefile 中声明的目标。
- `make all` 当前依赖未定义的 `build`、`fmt`、`errcheck` 目标，不要作为通用验证命令。
- `make e2e-test` 依赖 `test/e2e` 目录；当前仓库未跟踪该目录时该命令不可用。
- `make init` 期望存在 `tools/tools.go`；运行前先确认 `tools/` 是否存在。
- `api/`、`proto/`、`tools/`、`test/e2e` 目录均不在仓库中，`make lint/generate/breaking/test-proto2http/validate/serve/update` 等 buf/swagger 目标当前不可用。
- `vendor/` 与 `go.sum` 在本地磁盘存在但被 `.gitignore` 忽略、未纳入版本控制；干净 checkout 后先执行 `go mod download`。本地存在 `vendor/` 时 go 默认以 vendor 模式构建。
- 仓库无 `.golangci.yml`，`make golint` 使用 golangci-lint 默认配置。
- Makefile 使用 `find`、`grep`、`awk` 等 Unix 工具；在 Windows 上优先使用 Git Bash/兼容 shell，或直接运行对应的 `go` 命令。

## 架构概览

### 根包入口

[jupiter.go](jupiter.go) 和 [jupiter_option.go](jupiter_option.go) 是对 [pkg/core/application/](pkg/core/application/) 的薄封装：

- `jupiter.New`、`jupiter.DefaultApp` 转发自 `application.New`、`application.DefaultApp`。
- `jupiter.Option`、`WithConfigParser`、`WithDisable` 等转发自 application 包。
- 框架实例类型 `jupiter.Application` 是 `application.Application` 的类型别名。

### Application 生命周期

核心实现位于 [pkg/core/application/application.go](pkg/core/application/application.go)：

1. `New(fns...)` 创建 `Application` 并执行 `Startup(fns...)`；`DefaultApp()` 只初始化实例。
2. 初始化阶段创建生命周期控制器、server 列表、默认 logger、TOML 配置解析器、组件列表，并解析已注册 flags。
3. `Serve(servers...)` 只注册实现 [pkg/server/server.go](pkg/server/server.go) 中 `Server` 接口的服务。
4. `Run(servers...)` 注册额外 server，触发 `hooks.Stage_BeforeRun`，启动信号监听，再并发执行各 server 的 `Serve()`。
5. 退出时根据信号调用 `Stop()` 或 `GracefulStop(ctx)`，并触发 `Stage_BeforeStop`、`Stage_AfterStop`。

Hook 定义在 [pkg/core/hooks/hooks.go](pkg/core/hooks/hooks.go)，当前阶段包括配置加载前后、运行前、停止前后。Hook 执行顺序是后注册先执行。

### 配置系统

配置全局入口在 [pkg/conf/](pkg/conf/)，命令行 flag 注册在 [pkg/conf/init.go](pkg/conf/init.go)：

- 配置通常通过 `--config=config.toml` 加载，而不是 `jupiter.WithConfigFile`（当前代码没有该 Option）。
- `--config` 根据扩展名选择 TOML、YAML 或 JSON 解析器，并通过 `conf.NewDataSource` 构造数据源。
- 内置 datasource 注册来自 application 包的空白导入，当前源码包含 `file://` 和 `http://`。
- 代码中也可直接使用 `conf.LoadFromDataSource(ds, unmarshaller)` 或 `conf.LoadFromReader(r, unmarshaller)`。
- 读取配置的常见方式是 `conf.UnmarshalKey(key, &cfg)`，组件通常封装为 `StdConfig(name)`。

常见组件配置键约定：

```go
server := xgin.StdConfig("gin").MustBuild()     // server.gin
redisClient := redis.StdConfig("default").MustSingleton() // redis.default
db := gorm.StdConfig("default").MustBuild()     // database.default
logger := xlog.StdConfig("jupiter", "default").Build() // jupiter.logger.default
```

### Server、Component 与 Worker

- Server 抽象在 [pkg/server/server.go](pkg/server/server.go)：实现 `Serve`、`Stop`、`GracefulStop`、`Info`、`Healthz` 即可被 `Application.Serve` 管理。
- 当前 server 实现位于 [pkg/server/governor/](pkg/server/governor/)、[pkg/server/xgin/](pkg/server/xgin/)、[pkg/server/xhertz/](pkg/server/xhertz/)；各自通过 `StdConfig`/`RawConfig` 构建。
- Worker 命令注册与执行在 [pkg/worker/worker.go](pkg/worker/worker.go)，`Application.RegisterWorker` 最终调用 `worker.RegisterFunc`。

### 主要包分层

- [pkg/core/](pkg/core/)：应用生命周期、hook、信号、指标、全局常量和单例工具。
- [pkg/flag/](pkg/flag/)：全局命令行 flag 注册与解析（`--config`、`--version` 等）。
- [pkg/conf/](pkg/conf/)：全局配置、数据源、环境变量加载和反序列化。
- [pkg/server/](pkg/server/)：服务接口与 HTTP/governor 等 server 实现。
- [pkg/client/](pkg/client/)：外部服务客户端，如 Redis、Resty。
- [pkg/store/](pkg/store/)：数据库/搜索存储封装，如 GORM、MongoDB、MeiliSearch。
- [pkg/cache/](pkg/cache/)：本地缓存实现和缓存接口。
- [pkg/xlog/](pkg/xlog/)：基于 zap 的日志配置和默认 logger。
- [pkg/util/](pkg/util/)：通用工具包，测试多为普通 Go 单测。
- [docs/](docs/)：组件文档；若文档与源码不一致，以源码为准。

## 文档和规则文件

- 未发现 `.cursor/rules/`、`.cursorrules` 或 `.github/copilot-instructions.md`。
- [README.md](README.md) 提供项目定位和功能清单，但其中部分命令、示例路径和配置 API 可能滞后；其引用的 `docs/component-configuration.md`、`docs/configuration-examples.md` 不存在，链接已失效。更新文档时应先用源码验证。
- [docs/README.md](docs/README.md) 是文档目录入口，具体组件说明分散在 `docs/*.md`。
- [docs/superpowers/](docs/superpowers/) 存放 plan/spec 设计文档（现有 meilisearch 分面搜索一例）；较大改动按 plan → spec → 实现的流程在此记录。
