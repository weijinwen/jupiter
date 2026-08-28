# Jupiter 配置管理

本文档详细介绍 Jupiter 框架的配置系统，包括配置结构、加载方式、环境覆盖、动态配置以及各组件的配置控制功能。

## 概述

Jupiter 框架提供统一的配置管理系统，支持多种数据源和格式，通过分层配置和环境覆盖机制，满足不同环境下的配置需求。配置系统是框架的核心组件之一，为其他组件提供配置支持。

### 功能特性

- **多数据源支持**：支持从文件、环境变量、etcd、HTTP 等多种数据源加载配置
- **环境覆盖**：支持通过环境变量和命令行参数覆盖配置
- **动态配置**：支持配置热更新和变更监听
- **组件配置控制**：各组件提供灵活的功能开关配置
- **配置加密**：支持敏感配置的加密存储

## 配置结构

### 基本结构

Jupiter 配置采用分层结构，支持环境覆盖：

```toml
# 基础配置
[app]
name = "example-service"
version = "1.0.0"
region = "default"
zone = "default"

# 开发环境配置
[app.dev]
env = "development"
debug = true

# 生产环境配置
[app.prod]
env = "production"
debug = false
```

### 常用配置项

Jupiter 框架定义了一些常用的配置项：

```toml
# 应用配置
[app]
name = "example-service"
version = "1.0.0"
region = "default"
zone = "default"
env = "development"
debug = true

# 服务器配置
[server.http]
host = "0.0.0.0"
port = 8080
mode = "debug"

[server.grpc]
host = "0.0.0.0"
port = 9090

# 数据库配置
[store.mysql]
driver = "mysql"
dsn = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = "1h"

# Redis 配置
[client.redis]
address = "localhost:6379"
password = ""
db = 0
pool_size = 10

# 日志配置
[logger]
level = "info"
format = "json"
output = ["stdout", "file"]
file_path = "./logs/app.log"
max_size = 100
max_age = 7
max_backups = 10

# 监控配置
[monitor]
enable = true
address = "localhost:9091"
path = "/metrics"
```

## 配置加载

### 从文件加载

最常用的配置加载方式是从文件加载：

```go
// 从 TOML 文件加载配置
conf.LoadFromFile("config.toml")

// 从 JSON 文件加载配置
conf.LoadFromFile("config.json")

// 从 YAML 文件加载配置
conf.LoadFromFile("config.yaml")
```

### 从环境变量加载

Jupiter 支持从环境变量加载配置：

```go
// 从环境变量加载配置
conf.LoadFromEnv()

// 设置环境变量前缀
conf.LoadFromEnv(conf.WithPrefix("JUPITER_"))
```

环境变量命名规则：
- 使用 `__` 分隔配置层级
- 使用 `_` 分隔单词
- 例如：`JUPITER_APP__NAME` 对应 `[app]name`

### 从 etcd 加载

Jupiter 支持从 etcd 加载配置，实现配置的集中管理和动态更新：

```go
// 从 etcd 加载配置
conf.LoadFromEtcd("config/key")

// 带选项的 etcd 加载
conf.LoadFromEtcd("config/key", 
    conf.WithEtcdEndpoints([]string{"http://localhost:2379"}),
    conf.WithEtcdDialTimeout(5*time.Second),
    conf.WithEtcdUsername("user"),
    conf.WithEtcdPassword("password"),
)
```

### 从 HTTP 加载

Jupiter 支持从 HTTP 接口加载配置：

```go
// 从 HTTP 接口加载配置
conf.LoadFromHTTP("http://config-server/api/config")

// 带选项的 HTTP 加载
conf.LoadFromHTTP("http://config-server/api/config",
    conf.WithHTTPTimeout(10*time.Second),
    conf.WithHTTPHeaders(map[string]string{
        "Authorization": "Bearer token",
    }),
)
```

### 多源加载

Jupiter 支持从多个数据源加载配置，按优先级合并：

```go
// 多源加载配置
conf.Load(
    conf.WithFile("config.toml"),
    conf.WithEnv(),
    conf.WithEtcd("config/key"),
)
```

## 配置解析

### 定义配置结构

定义 Go 结构体来映射配置：

```go
type Config struct {
    App    AppConfig    `toml:"app"`
    Server ServerConfig `toml:"server"`
    Store  StoreConfig  `toml:"store"`
}

type AppConfig struct {
    Name    string `toml:"name"`
    Version string `toml:"version"`
    Region  string `toml:"region"`
    Zone    string `toml:"zone"`
    Env     string `toml:"env"`
    Debug   bool   `toml:"debug"`
}

type ServerConfig struct {
    HTTP HTTPConfig `toml:"http"`
    GRPC GRPCConfig `toml:"grpc"`
}

type HTTPConfig struct {
    Host string `toml:"host"`
    Port int    `toml:"port"`
    Mode string `toml:"mode"`
}

type GRPCConfig struct {
    Host string `toml:"host"`
    Port int    `toml:"port"`
}

type StoreConfig struct {
    MySQL MySQLConfig `toml:"mysql"`
    Redis RedisConfig `toml:"redis"`
}

type MySQLConfig struct {
    Driver          string        `toml:"driver"`
    DSN             string        `toml:"dsn"`
    MaxOpenConns    int           `toml:"max_open_conns"`
    MaxIdleConns    int           `toml:"max_idle_conns"`
    ConnMaxLifetime time.Duration `toml:"conn_max_lifetime"`
}

type RedisConfig struct {
    Address  string `toml:"address"`
    Password string `toml:"password"`
    DB       int    `toml:"db"`
    PoolSize int    `toml:"pool_size"`
}
```

### 加载配置到结构体

将配置加载到定义的结构体：

```go
var cfg Config

// 加载配置
if err := conf.Load("config.toml", &cfg); err != nil {
    panic(err)
}

// 使用配置
fmt.Printf("应用名称: %s\n", cfg.App.Name)
fmt.Printf("HTTP 端口: %d\n", cfg.Server.HTTP.Port)
```

### 配置验证

Jupiter 支持配置验证：

```go
type Config struct {
    App    AppConfig    `toml:"app" validate:"required"`
    Server ServerConfig `toml:"server" validate:"required"`
}

type AppConfig struct {
    Name    string `toml:"name" validate:"required"`
    Version string `toml:"version" validate:"required,semver"`
    Region  string `toml:"region" validate:"required"`
    Zone    string `toml:"zone" validate:"required"`
    Env     string `toml:"env" validate:"required,oneof=development testing staging production"`
    Debug   bool   `toml:"debug"`
}

// 加载并验证配置
var cfg Config
if err := conf.Load("config.toml", &cfg, conf.WithValidate()); err != nil {
    panic(err)
}
```

## 环境覆盖

### 环境变量覆盖

Jupiter 支持通过环境变量覆盖配置：

```go
// 启用环境变量覆盖
conf.Load("config.toml", &cfg, conf.WithEnvOverride())
```

环境变量命名规则：
- 使用 `__` 分隔配置层级
- 使用 `_` 分隔单词
- 例如：`JUPITER_APP__NAME` 对应 `[app]name`

### 命令行参数覆盖

Jupiter 支持通过命令行参数覆盖配置：

```go
// 启用命令行参数覆盖
conf.Load("config.toml", &cfg, conf.WithFlagOverride())
```

命令行参数命名规则：
- 使用 `--` 前缀
- 使用 `.` 分隔配置层级
- 例如：`--app.name=new-name` 对应 `[app]name`

## 动态配置

### 配置监听

Jupiter 支持监听配置变化：

```go
// 监听配置变化
conf.Watch("app.name", func(oldValue, newValue interface{}) {
    fmt.Printf("配置变化: %s -> %s\n", oldValue, newValue)
})
```

### 配置热更新

Jupiter 支持配置热更新：

```go
// 启用配置热更新
conf.Load("config.toml", &cfg, conf.WithWatch())

// 监听配置变化
conf.OnChange(func() {
    fmt.Println("配置已更新")
})
```

## 配置加密

### 敏感配置加密

Jupiter 支持对敏感配置进行加密：

```go
// 加密配置
encrypted := conf.Encrypt("password", "secret-key")

// 解密配置
decrypted := conf.Decrypt(encrypted, "secret-key")
```

### 配置文件加密

Jupiter 支持对整个配置文件进行加密：

```go
// 加密配置文件
conf.EncryptFile("config.toml", "config.enc", "secret-key")

// 从加密文件加载配置
conf.LoadFromEncryptedFile("config.enc", "secret-key", &cfg)
```

## 组件配置控制

### 概述

Jupiter 框架中的组件提供了灵活的配置控制功能，允许开发者通过配置文件或代码来启用或禁用特定功能。这种设计提供了以下优势：

- **一致性**：所有组件使用统一的命名规范和默认值策略
- **灵活性**：可以根据环境和需求灵活启用或禁用功能
- **性能优化**：通过禁用不需要的功能来减少资源消耗
- **配置清晰**：明确的配置项名称和默认值

### 通用配置规范

#### 命名规范

所有功能开关配置项遵循以下命名规范：

- 使用 `Enable` 前缀表示功能开关
- 使用驼峰命名法
- 名称清晰表达功能含义

#### 默认值策略

所有功能开关的默认值为 `false`，即默认不启用任何额外功能，保持最小化资源消耗。

### Redis 组件配置

#### 功能开关

Redis 组件提供以下功能开关：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| EnableMetric | bool | false | 是否启用指标收集 |
| EnableTrace | bool | false | 是否启用链路追踪 |
| EnableAccessLog | bool | false | 是否启用访问日志 |
| EnableSentinel | bool | false | 是否启用哨兵模式 |

#### 配置示例

```toml
[redis.default]
addr = "127.0.0.1:6379"
password = ""
db = 0
maxIdle = 10
maxActive = 100
idleTimeout = "80s"
dialTimeout = "200ms"
readTimeout = "200ms"
writeTimeout = "200ms"
onDialError = "error"
slowLogThreshold = "100ms"
enableMetric = true
enableTrace = true
enableAccessLog = true
enableSentinel = false
```

#### 代码示例

```go
// 使用默认配置
client := redis.StdConfig("default").Build()

// 使用自定义配置
config := &redis.Config{
    Addr:               "127.0.0.1:6379",
    Password:           "",
    DB:                 0,
    MaxIdle:            10,
    MaxActive:          100,
    IdleTimeout:        80 * time.Second,
    DialTimeout:        200 * time.Millisecond,
    ReadTimeout:        200 * time.Millisecond,
    WriteTimeout:       200 * time.Millisecond,
    OnDialError:        "error",
    SlowLogThreshold:   100 * time.Millisecond,
    EnableMetric:       true,
    EnableTrace:        true,
    EnableAccessLog:    true,
    EnableSentinel:     false,
}
client := redis.Build(config)
```

### GORM 组件配置

#### 功能开关

GORM 组件提供以下功能开关：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| EnableMetric | bool | false | 是否启用指标收集 |
| EnableTrace | bool | false | 是否启用链路追踪 |
| EnableAccessLog | bool | false | 是否启用访问日志 |
| EnableDBLog | bool | false | 是否启用数据库日志 |

#### 配置示例

```toml
[database.mysql_example]
dsn = "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
debug = true
maxIdleConns = 10
maxOpenConns = 100
connMaxLifetime = "300s"
slowThreshold = "200ms"
disableForeignKeyConstraintWhenMigrating = false
enableMetric = true
enableTrace = true
enableAccessLog = true
enableDBLog = true
```

#### 代码示例

```go
// 使用默认配置
db := gorm.StdConfig("mysql_example").Build()

// 使用自定义配置
config := &gorm.Config{
    DSN:                                  "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    Debug:                                true,
    MaxIdleConns:                         10,
    MaxOpenConns:                         100,
    ConnMaxLifetime:                      300 * time.Second,
    SlowThreshold:                        200 * time.Millisecond,
    DisableForeignKeyConstraintWhenMigrating: false,
    EnableMetric:                         true,
    EnableTrace:                          true,
    EnableAccessLog:                      true,
    EnableDBLog:                          true,
}
db := gorm.Build(config)
```

### MeiliSearch 组件配置

#### 功能开关

MeiliSearch 组件提供以下功能开关：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Debug | bool | false | 是否启用调试模式 |
| EnableMetric | bool | false | 是否启用指标收集 |
| EnableAccessLog | bool | false | 是否启用访问日志 |
| SlowThreshold | string | "100ms" | 慢查询阈值，超过此阈值的查询将被记录 |

#### 配置示例

```toml
[meilisearch.default]
host = "http://localhost:7700"
apiKey = ""
debug = false
enableMetric = true
enableAccessLog = true
slowThreshold = "100ms"
```

#### 代码示例

```go
// 使用默认配置
client := meilisearch.StdConfig("default").Build()

// 使用自定义配置
config := &meilisearch.Config{
    Host:           "http://localhost:7700",
    APIKey:         "",
    Debug:          false,
    EnableMetric:   true,
    EnableAccessLog: true,
    SlowThreshold:  100 * time.Millisecond,
}
client := meilisearch.Build(config)
```

## 配置最佳实践

### 1. 配置分层

按照环境分层配置：

```toml
# 基础配置
[app]
name = "example-service"
version = "1.0.0"

# 开发环境
[app.dev]
env = "development"
debug = true

# 测试环境
[app.test]
env = "testing"
debug = true

# 生产环境
[app.prod]
env = "production"
debug = false
```

### 2. 敏感信息分离

将敏感信息分离到环境变量或加密配置：

```go
// 从环境变量加载敏感信息
conf.LoadFromEnv(conf.WithPrefix("JUPITER_SECRET_"))

// 或者使用加密配置
conf.LoadFromEncryptedFile("secret.enc", "secret-key")
```

### 3. 配置验证

始终对配置进行验证：

```go
type Config struct {
    App    AppConfig    `toml:"app" validate:"required"`
    Server ServerConfig `toml:"server" validate:"required"`
}

// 加载并验证配置
var cfg Config
if err := conf.Load("config.toml", &cfg, conf.WithValidate()); err != nil {
    panic(err)
}
```

### 4. 默认值

为配置提供合理的默认值：

```go
type ServerConfig struct {
    Host string `toml:"host" default:"0.0.0.0"`
    Port int    `toml:"port" default:"8080"`
    Mode string `toml:"mode" default:"release"`
}
```

### 5. 环境相关配置

根据不同环境配置不同的功能开关：

- **开发环境**：启用所有功能，便于调试和监控
- **测试环境**：启用必要的功能，如指标收集和访问日志
- **生产环境**：仅启用必要的功能，如指标收集，禁用调试功能

### 6. 性能优化

根据应用需求启用功能：

- **高并发场景**：禁用不必要的日志和追踪功能
- **调试场景**：启用详细的日志和追踪功能
- **监控场景**：启用指标收集功能

### 7. 配置管理

- 使用配置文件管理功能开关，便于环境切换
- 对于关键功能，可以在代码中添加额外检查
- 定期审查功能开关的使用情况，优化配置

## 完整配置示例

```toml
# 应用配置
[app]
name = "user-service"
version = "1.0.0"
region = "default"
zone = "default"
env = "development"
debug = true

# 服务器配置
[server.http]
host = "0.0.0.0"
port = 8080
mode = "debug"
read_timeout = "30s"
write_timeout = "30s"
idle_timeout = "60s"

[server.grpc]
host = "0.0.0.0"
port = 9090
timeout = "30s"

# 数据库配置
[store.mysql]
driver = "mysql"
dsn = "user:password@tcp(localhost:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = "1h"
enableMetric = true
enableTrace = false
enableAccessLog = false
enableDBLog = false

# Redis 配置
[client.redis]
address = "localhost:6379"
password = ""
db = 0
pool_size = 10
min_idle_conns = 5
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"
enableMetric = true
enableTrace = false
enableAccessLog = false
enableSentinel = false

# MeiliSearch 配置
[search.default]
addr = "http://localhost:7700"
apiKey = ""
timeout = "3000ms"
enableMetric = true
enableAccessLog = false
slowThreshold = "1s"

# 日志配置
[logger]
level = "info"
format = "json"
output = ["stdout", "file"]
file_path = "./logs/app.log"
max_size = 100
max_age = 7
max_backups = 10
compress = true

# 监控配置
[monitor]
enable = true
address = "localhost:9091"
path = "/metrics"
push_interval = "10s"
push_gateway = "http://pushgateway:9091"

# 追踪配置
[tracer]
enable = true
type = "jaeger"
endpoint = "http://jaeger:14268/api/traces"
sample_rate = 0.1

# 注册中心配置
[registry]
enable = true
type = "etcd"
endpoints = ["http://localhost:2379"]
timeout = "5s"
dial_timeout = "5s"
username = ""
password = ""
```

## 环境配置示例

### 开发环境配置

```toml
# Redis 开发环境配置
[redis.default]
addr = "127.0.0.1:6379"
enableMetric = true
enableTrace = true
enableAccessLog = true
enableSentinel = false

# GORM 开发环境配置
[database.mysql_example]
dsn = "mysql://user:password@tcp(localhost:3306)/dev_db?charset=utf8mb4&parseTime=True&loc=Local"
debug = true
enableMetric = true
enableTrace = true
enableAccessLog = true
enableDBLog = true

# MeiliSearch 开发环境配置
[search.default]
addr = "http://localhost:7700"
debug = true
enableMetric = true
enableAccessLog = true
```

### 生产环境配置

```toml
# Redis 生产环境配置
[redis.default]
addr = "redis-cluster:6379"
password = "prod_password"
enableMetric = true
enableTrace = false
enableAccessLog = false
enableSentinel = true

# GORM 生产环境配置
[database.mysql_example]
dsn = "mysql://user:password@tcp(mysql-cluster:3306)/prod_db?charset=utf8mb4&parseTime=True&loc=Local"
debug = false
enableMetric = true
enableTrace = false
enableAccessLog = false
enableDBLog = false

# MeiliSearch 生产环境配置
[search.default]
addr = "http://meilisearch-cluster:7700"
apiKey = "prod-api-key"
debug = false
enableMetric = true
enableAccessLog = false
slowThreshold = "500ms"
```

## 总结

Jupiter 配置系统提供了灵活、强大的配置管理能力，支持多种数据源、环境覆盖、动态更新和配置加密等功能。统一的命名规范和默认值策略使得配置更加清晰和一致，降低了学习和维护成本。

各组件的配置控制功能提供了灵活的方式来管理功能特性，通过合理配置功能开关，可以在不同环境中实现最佳的性能和功能平衡。

## 下一步

- 学习 [服务器组件](./servers.md)
- 探索 [客户端组件](./clients.md)
- 查看 [存储组件](./stores.md)
- 了解 [日志组件](./xlog.md)
