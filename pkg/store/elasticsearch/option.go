package elasticsearch

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	cfg "github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/core/singleton"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// Config Elasticsearch 客户端配置。
//
// 注意：重试、调试日志、批量写入均由 go-elasticsearch 官方客户端内置提供，
// 本配置只暴露官方 Config 未覆盖的部分（连接池参数）以及框架侧的可观测性开关。
type Config struct {
	// Addresses ES 节点地址列表，如 ["http://localhost:9200"]
	Addresses []string `json:"addresses" toml:"addresses"`
	// Username Basic Auth 用户名
	Username string `json:"username" toml:"username"`
	// Password Basic Auth 密码
	Password string `json:"password" toml:"password"`
	// APIKey API Key 认证（设置后覆盖 Basic Auth）
	APIKey string `json:"apiKey" toml:"apiKey"`
	// CloudID Elastic Cloud 部署 ID
	CloudID string `json:"cloudID" toml:"cloudID"`
	// CertificateFingerprint ES 首次启动时输出的 SHA256 证书指纹
	CertificateFingerprint string `json:"certificateFingerprint" toml:"certificateFingerprint"`

	/****** 重试，直接透传给官方客户端，框架不再叠加一层 ******/
	// MaxRetries 最大重试次数，默认 3
	MaxRetries int `json:"maxRetries" toml:"maxRetries"`
	// DisableRetry 关闭重试
	DisableRetry bool `json:"disableRetry" toml:"disableRetry"`
	// RetryOnStatus 触发重试的状态码，默认 502、503、504
	RetryOnStatus []int `json:"retryOnStatus" toml:"retryOnStatus"`

	/****** 连接池，官方 Config 只暴露 Transport，故在此补齐 ******/
	// DialTimeout 建立连接超时，默认 3s
	DialTimeout time.Duration `json:"dialTimeout" toml:"dialTimeout"`
	// KeepAlive TCP keepalive 间隔，默认 30s
	KeepAlive time.Duration `json:"keepAlive" toml:"keepAlive"`
	// MaxIdleConns 最大空闲连接数，默认 100
	MaxIdleConns int `json:"maxIdleConns" toml:"maxIdleConns"`
	// MaxIdleConnsPerHost 每个 host 的最大空闲连接数，默认 100
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost" toml:"maxIdleConnsPerHost"`
	// MaxConnsPerHost 每个 host 的最大连接数，默认 100
	MaxConnsPerHost int `json:"maxConnsPerHost" toml:"maxConnsPerHost"`
	// IdleConnTimeout 空闲连接存活时间，默认 90s
	IdleConnTimeout time.Duration `json:"idleConnTimeout" toml:"idleConnTimeout"`
	// ResponseHeaderTimeout 等待响应头超时，默认 30s
	ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout" toml:"responseHeaderTimeout"`
	// DisableKeepAlives 关闭长连接
	DisableKeepAlives bool `json:"disableKeepAlives" toml:"disableKeepAlives"`

	/****** 可观测性 ******/
	// Debug 调试模式，启用官方 EnableDebugLogger 打印请求响应
	Debug bool `json:"debug" toml:"debug"`
	// EnableMetric 指标采集开关
	EnableMetric bool `json:"enableMetric" toml:"enableMetric"`
	// EnableAccessLog 访问日志开关
	EnableAccessLog bool `json:"enableAccessLog" toml:"enableAccessLog"`
	// SlowThreshold 慢查询阈值，0 表示关闭慢查询日志
	SlowThreshold time.Duration `json:"slowThreshold" toml:"slowThreshold"`

	// OnDialError 启动时连通性检查失败的处理方式："panic" | "error"
	OnDialError string `json:"onDialError" toml:"onDialError"`

	logger *zap.Logger
	name   string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		name:      "default",
		Addresses: []string{"http://localhost:9200"},

		MaxRetries: 3,

		DialTimeout:           cast.ToDuration("3s"),
		KeepAlive:             cast.ToDuration("30s"),
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       cast.ToDuration("90s"),
		ResponseHeaderTimeout: cast.ToDuration("30s"),

		Debug:           false,
		EnableMetric:    true,
		EnableAccessLog: false,
		SlowThreshold:   cast.ToDuration("1s"),

		OnDialError: "error",
		logger:      xlog.Jupiter().Named(ecode.ModStoreElasticsearch),
	}
}

// StdConfig 从标准配置路径读取，key 为 "jupiter.elasticsearch.<name>"
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("elasticsearch", name))
}

// RawConfig 从指定配置 key 读取
func RawConfig(key string) *Config {
	config := DefaultConfig()
	if err := cfg.UnmarshalKey(key, &config, cfg.TagName("toml")); err != nil {
		config.logger.Panic("unmarshal config:"+key, xlog.FieldErr(err), xlog.FieldName(key), xlog.FieldExtMessage(config))
	}
	config.name = key
	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint(key, config)
	}
	return config
}

// NewElasticsearch 通过名称创建单例 Elasticsearch 客户端
func NewElasticsearch(name string) *Client {
	return StdConfig(name).MustSingleton()
}

// Singleton 返回单例客户端，同名配置只创建一次
func (config *Config) Singleton() (*Client, error) {
	if val, ok := singleton.Load(constant.ModuleStoreElasticsearch, config.name); ok && val != nil {
		return val.(*Client), nil
	}
	cc, err := config.Build()
	if err != nil {
		xlog.Jupiter().Error("build elasticsearch client failed", zap.Error(err))
		return nil, err
	}
	singleton.Store(constant.ModuleStoreElasticsearch, config.name, cc)
	return cc, nil
}

// MustSingleton 返回单例客户端，出错时 panic
func (config *Config) MustSingleton() *Client {
	return lo.Must(config.Singleton())
}

// MustBuild 每次创建新客户端，出错时 panic
func (config *Config) MustBuild() *Client {
	return lo.Must(config.Build())
}

// Build 创建新客户端并验证连通性
func (config *Config) Build() (*Client, error) {
	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint("elasticsearch's config: "+config.name, config)
	}
	if len(config.Addresses) == 0 && config.CloudID == "" {
		return nil, ErrEmptyAddresses
	}

	esCfg := elasticsearch.Config{
		Addresses:              config.Addresses,
		Username:               config.Username,
		Password:               config.Password,
		APIKey:                 config.APIKey,
		CloudID:                config.CloudID,
		CertificateFingerprint: config.CertificateFingerprint,

		// 重试完全交给官方客户端，框架不再叠加，避免重试次数相乘
		MaxRetries:    config.MaxRetries,
		DisableRetry:  config.DisableRetry,
		RetryOnStatus: config.RetryOnStatus,

		Transport: config.buildTransport(),

		// EnableDebugLogger 与 Logger 是两个独立的装配点，可同时生效
		EnableDebugLogger: config.Debug,
	}

	// 必须按指针判空后再赋值：直接把可能为 nil 的 *esLogger 赋给 Logger 接口字段，
	// 会得到一个非 nil 的接口（内含 nil 指针），使客户端误以为已配置日志器并回调到空接收者。
	if l := config.buildLogger(); l != nil {
		esCfg.Logger = l
	}

	typedClient, err := elasticsearch.NewTypedClient(esCfg)
	if err != nil {
		config.logger.Error("new elasticsearch typed client failed", xlog.FieldErr(err))
		return nil, err
	}

	client := &Client{TypedClient: typedClient, config: config}

	if _, err = typedClient.Info().Do(context.Background()); err != nil {
		config.logger.Error("elasticsearch ping failed", xlog.FieldErr(err), xlog.FieldName(config.name))
		if config.OnDialError == "panic" {
			config.logger.Panic("elasticsearch client start err: " + err.Error())
		}
		return nil, ErrConnectFailed
	}

	instances.Store(config.name, client)
	return client, nil
}

// buildTransport 构建带连接池参数的 http.Transport。
// 官方 Config 只暴露 Transport 字段而不暴露池化参数，故在此补齐。
func (config *Config) buildTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
	}
}

// buildLogger 构建观测用的 Logger。三项开关全关时返回 nil，让客户端完全跳过日志回调。
func (config *Config) buildLogger() *esLogger {
	if !config.EnableMetric && !config.EnableAccessLog && config.SlowThreshold <= 0 {
		return nil
	}
	return &esLogger{
		name:   config.name,
		config: config,
		logger: config.logger,
	}
}
