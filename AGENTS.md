# AGENTS.md

本仓库开发与协作的快速指引。文档与源码不一致时一律以源码和 `go.mod` 为准。

## 项目概览

Jupiter 是 Go 微服务框架库，参考了 [douyu/jupiter](https://github.com/douyu/jupiter) 项目，在其设计思路上精简重构，去除了不需要的功能，仅保留核心组件。模块路径 `github.com/weijinwen/jupiter`，`go.mod` 声明 `go 1.25.0`（README/文档中的版本号可能滞后，以 `go.mod` 为准）。根包 `jupiter.go` / `jupiter_option.go` 仅薄封装 `pkg/core/application`，真正实现都在 `pkg/` 下。

主要目录：

- `pkg/core/` 应用核心：`application/`（Application 生命周期）、`hooks/`（阶段 hook）、`signals/`（信号）、`metric/`、`singleton/`、`constant/`（含配置键前缀 `ConfigKey`）
- `pkg/conf/` 全局配置：`LoadFromReader`/`LoadFromDataSource` 加载、`UnmarshalKey` 读取、`datasource/`（file/http）、watch 热更新
- `pkg/flag/` 命令行 flag（`--config`、`--config-namespace`、`--watch`、`--version` 等）
- `pkg/server/` 服务接口与实现：`governor`（运维端口）、`xgin`（Gin）、`xhertz`（Hertz，依赖 `cloudwego/hertz`，默认 netpoll 传输）
- `pkg/client/` 客户端：`redis`、`resty`
- `pkg/store/` 存储封装：`elasticsearch`、`gorm`、`meilisearch`、`mongox`
- `pkg/cache/` 缓存：`xfreecache`、`xgolanglru`
- `pkg/xlog/` 基于 zap 的日志（Hertz 引擎日志 hlog 也接入此管道）
- `pkg/util/` 通用工具（xgo、xmap、xcast 等）；`pkg/worker/` worker 命令注册与执行
- `pkg/env.go` 运行环境信息（AppMode/AppLogDir 等，读 `JUPITER_*` 环境变量）
- `examples/xgin/`、`examples/xhertz/` 可运行示例（main.go + config.toml）
- `docs/` 组件文档

## 架构与生命周期

- 入口：根包 `jupiter.New(fns...)` / `DefaultApp()`（`pkg/core/application` 的薄封装）。可用 Option 仅 `WithConfigParser`、`WithDisable`；**不存在 `WithConfigFile`**（docs/quickstart.md 中的该写法已失效）。
- 典型流程：`New` → `Startup`（注册 flags、加载配置）→ `app.Serve(server)` → `app.Run()`（阻塞，监听 SIGINT/SIGTERM）→ `GracefulStop` 优雅停止。
- Hook 阶段（`pkg/core/hooks`，`hooks.Register(stage, fn)` 或 `app.RegisterHooks`）：`Stage_BeforeLoadConfig` → `Stage_AfterLoadConfig` → `Stage_BeforeRun` → `Stage_BeforeStop` → `Stage_AfterStop`。
- `WithDisable` 可关：`DisableParserFlag`、`DisableLoadConfig`、`DisableDefaultGovernor`。
- Server 组件统一模式：`StdConfig(name)` → `MustBuild()`（内置 recover、访问日志/慢查询、Prometheus 指标中间件）→ 注册路由 → `app.Serve(...)`。
- Worker：`worker.Register(cmd)` / `worker.RegisterFunc(name, fn)` 或 `app.RegisterWorker(name, fn)`，按名执行（默认 30 分钟超时）。

## 常用命令

```bash
# 编译与静态检查
go build ./... && go vet ./...

# 测试
make unit-test                       # = go test -race -coverprofile=coverage.txt -covermode=atomic ./...
go test ./pkg/conf                   # 单包
go test ./pkg/conf -run TestLoadFromReader -v   # 单函数
make covsh-unit                      # 查看覆盖率报告（需先安装 gocovsh）

# Lint（需先安装 golangci-lint，使用默认配置）
golangci-lint run -v                 # 或 make golint
```

## 配置系统（易踩坑）

- 通过 `--config=config.toml` 加载，按扩展名选 TOML/YAML/JSON 解析器；`--watch`（或环境变量 `JUPITER_CONFIG_WATCH`）开启热更新。
- **配置键必须带 `jupiter.` 前缀**：组件 `StdConfig(name)` 内部经 `pkg/core/constant/config.go` 的 `ConfigKey` 拼前缀（默认 `"jupiter"`），实际查找 `jupiter.server.hertz` 这类键。配置文件写 `[server.hertz]`（无前缀）会**静默回落默认值**（已实测：配置 `port = 1111`，`xhertz.StdConfig("hertz")` 仍得默认 9092）。正确写法 `[jupiter.server.hertz]`；或代码里 `constant.SetConfigPrefix("")`。
  - 例外：`xlog.StdConfig(prefix, name)` 自行传前缀，键为 `{prefix}.logger.{name}`。
  - `--config-namespace` flag 只改 `conf.GetOptions.Namespace`，键查找并未使用它，改不了上述前缀。
  - 现状：`examples/*/config.toml` 与 docs 中多数示例键名（`[server.hertz]`、`[server.http]`）缺前缀，照抄不生效，属已知文档滞后。
- **键名用 camelCase**：conf 按字段名大小写不敏感匹配、不处理下划线，`slowQueryThresholdInMilli`、`disableMetric` 这类多词键必须 camelCase。
- 代码读取：`conf.UnmarshalKey(key, &v)` 按字面键查找（无自动前缀），缺失返回 `ErrInvalidKey`（StdConfig 据此回落默认值）。

## 重要注意事项（改动前必读）

- **vendor 与 go.sum 被忽略**：`vendor/` 与 `go.sum` 在本地存在但被 `.gitignore` 忽略、未纳入版本控制。干净 checkout 后先 `go mod download`；本地有 `vendor/` 时 go 默认以 vendor 模式构建。
- **跨平台**：Makefile 使用 `find`/`grep`/`awk`，Windows 上优先用 Git Bash 或直接运行对应 `go` 命令。
- **xhertz 端口占用即 panic**：Hertz 由 netpoll 自管 listener，端口被占时进程在启动阶段直接 panic 退出（不同于 xgin 在构建期报错），部署与集成测试需确保端口可用。
- **改文档前先用源码验证**：README/docs 中部分命令、示例路径、配置 API 可能滞后于源码（如 docs 声称 Go 1.23/1.24，`go.mod` 实为 1.25.0；quickstart 的 `WithConfigFile` 已不存在）。新增服务器组件时同步更新 `docs/servers.md`。

## 详细参考

- [docs/README.md](docs/README.md)：组件文档入口（quickstart / concepts / config / servers / clients / stores / caches / utils / xlog / worker / meilisearch）。
- [docs/servers.md](docs/servers.md)：xgin / xhertz / governor 的配置项、CORS、WebSocket、超时与 TLS 说明。
- [pkg/conf/README.md](pkg/conf/README.md)、[pkg/xlog/README.md](pkg/xlog/README.md)：包级说明。
