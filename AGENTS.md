# AGENTS.md

本仓库开发与协作的快速指引。文档与源码不一致时一律以源码和 `go.mod` 为准。

## 项目概览

Jupiter 是 Go 微服务框架库，模块路径 `github.com/weijinwen/jupiter`，`go.mod` 声明 `go 1.25.0`（README/文档中的版本号可能滞后，以 `go.mod` 为准）。根包 `jupiter.go` / `jupiter_option.go` 仅薄封装 `pkg/core/application`，真正实现都在 `pkg/` 下。

主要目录：

- `pkg/core/` 应用生命周期、hook、信号、指标、全局单例
- `pkg/conf/` 全局配置、数据源、环境变量、反序列化
- `pkg/flag/` 命令行 flag（`--config`、`--version` 等）
- `pkg/server/` 服务接口与实现：`governor`、`xgin`、`xhertz`（Hertz，依赖 `cloudwego/hertz`）
- `pkg/client/` 客户端：`redis`、`resty`
- `pkg/store/` 存储封装：`gorm`、`meilisearch`、`mongox`
- `pkg/cache/` 缓存：`xfreecache`、`xgolanglru`
- `pkg/xlog/` 基于 zap 的日志
- `pkg/util/` 通用工具；`pkg/worker/` worker 命令注册
- `docs/` 组件文档

## 常用命令

```bash
# 测试
make unit-test                       # = go test -race -coverprofile=coverage.txt -covermode=atomic ./...
go test ./pkg/conf                   # 单包
go test ./pkg/conf -run TestLoadFromReader -v   # 单函数
make covsh-unit                      # 查看覆盖率报告

# Lint（仓库无 .golangci.yml，使用默认配置）
golangci-lint run -v                 # 或 make golint
```

## 重要注意事项（改动前必读）

- **vendor 与 go.sum 被忽略**：`vendor/` 与 `go.sum` 在本地存在但被 `.gitignore` 忽略、未纳入版本控制。干净 checkout 后先 `go mod download`；本地有 `vendor/` 时 go 默认以 vendor 模式构建。
- **多个 Makefile 目标当前不可用**：`api/`、`proto/`、`tools/`、`test/e2e` 目录不在仓库中，因此 `make lint`/`generate`/`breaking`/`test-proto2http`/`validate`/`serve`/`update`/`e2e-test`/`init` 等 buf/swagger 目标都会失败。
- **不要用 `make all` 做通用验证**：它依赖未定义的 `build`/`fmt`/`errcheck` 目标。默认目标是 `make help`。
- **跨平台**：Makefile 使用 `find`/`grep`/`awk`，Windows 上优先用 Git Bash 或直接运行对应 `go` 命令。
- **配置加载方式**：通过 `--config=config.toml` 加载（按扩展名选 TOML/YAML/JSON 解析器）；不存在 `jupiter.WithConfigFile` Option。组件配置统一用 `StdConfig(name)` 约定键名，如 `server.gin`、`redis.default`、`database.default`、`jupiter.logger.default`。
- **改文档前先用源码验证**：README/docs 中部分命令、示例路径、配置 API 可能滞后于源码（新增服务器组件时同步更新 `docs/servers.md`）。
- **较大改动流程**：按 plan → spec → 实现记录在 `docs/superpowers/`。

## 详细参考

- [docs/README.md](docs/README.md)：组件文档入口。
