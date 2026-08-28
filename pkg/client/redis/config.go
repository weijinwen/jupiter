package redis

import (
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"

	cfg "github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// Config ...
type Config struct {
	Addr     []string `json:"addr" toml:"addr"`
	Username string   `json:"username" toml:"username"`
	Password string   `json:"password" toml:"password"`

	/****** for github.com/redis/go-redis/v9 ******/
	// DB default 0,not recommend
	DB int `json:"db" toml:"db"`
	// PoolSize applies per Stub node and not for the whole Stub.
	PoolSize int `json:"poolSize" toml:"poolSize"`
	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int `json:"maxRetries" toml:"maxRetries"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int `json:"minIdleConns" toml:"minIdleConns"`
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `json:"dialTimeout" toml:"dialTimeout"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value 0 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration `json:"readTimeout" toml:"readTimeout"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration `json:"writeTimeout" toml:"writeTimeout"`
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration `json:"idleTimeout" toml:"idleTimeout"`

	// nice option
	Debug bool `json:"debug" toml:"debug"`
	// a require will be recorded if cost bigger than this
	SlowLogThreshold time.Duration `json:"slowThreshold" toml:"slowThreshold"`
	// 指标采集开关
	EnableMetric bool `json:"enableMetric" toml:"enableMetric"`

	// 访问日志开关
	EnableAccessLog bool `json:"enableAccessLog" toml:"enableAccessLog"`

	// OnDialError panic|error
	OnDialError string `json:"level"`

	// 分布式锁配置
	// 默认锁过期时间
	LockDefaultExpiration time.Duration `json:"lockDefaultExpiration" toml:"lockDefaultExpiration"`
	// 默认重试间隔
	LockDefaultRetryDelay time.Duration `json:"lockDefaultRetryDelay" toml:"lockDefaultRetryDelay"`
	// 默认最大重试次数，-1表示无限重试
	LockDefaultMaxRetries int `json:"lockDefaultMaxRetries" toml:"lockDefaultMaxRetries"`
	// 默认获取锁超时时间，0表示不超时
	LockDefaultTimeout time.Duration `json:"lockDefaultTimeout" toml:"lockDefaultTimeout"`

	logger *zap.Logger
	name   string
}

// DefaultConfig default config ...
func DefaultConfig() *Config {
	return &Config{
		name:         "default",
		DB:           0,
		PoolSize:     200,
		MinIdleConns: 20,
		DialTimeout:  cast.ToDuration("3s"),
		ReadTimeout:  cast.ToDuration("1s"),
		WriteTimeout: cast.ToDuration("1s"),
		IdleTimeout:  cast.ToDuration("60s"),
		Debug:        false,
		EnableMetric: false,

		EnableAccessLog: false,

		SlowLogThreshold: 0,
		logger:           xlog.Jupiter().Named(ecode.ModClientRedis),
		OnDialError:      "error",
		// 分布式锁默认配置
		LockDefaultExpiration: cast.ToDuration("30s"),
		LockDefaultRetryDelay: cast.ToDuration("100ms"),
		LockDefaultMaxRetries: -1,                    // 无限重试
		LockDefaultTimeout:    cast.ToDuration("0s"), // 不超时
	}
}

// StdConfig 哨兵模式，支持主从结构redis集群
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("redis", name))
}

func RawConfig(key string) *Config {
	var config = DefaultConfig()

	if err := cfg.UnmarshalKey(key, &config, cfg.TagName("toml")); err != nil {
		config.logger.Panic("unmarshal config:"+key, xlog.FieldErr(err), xlog.FieldName(key), xlog.FieldExtMessage(config))
	}

	config.name = key
	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint(key, config)
	}

	return config
}

func NewRedis(name string) *Client {
	return StdConfig(name).MustSingleton()
}
