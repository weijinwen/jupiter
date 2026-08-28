package meilisearch

import (
	"errors"
	"net/http"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/singleton"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

type (
	// Config options
	Config struct {
		// 目标服务地址
		Addr                  string        `json:"addr" toml:"addr"` // 目标地址
		ApiKey                string        `json:"apiKey" toml:"apiKey"`
		Timeout               time.Duration `json:"timeout" toml:"timeout" `
		KeepAlive             time.Duration `json:"keepAlive" toml:"keepAlive"`
		MaxIdleConns          int           `json:"maxIdleConns" toml:"maxIdleConns"`
		MaxIdleConnsPerHost   int           `json:"maxIdleConnsPerHost" toml:"maxIdleConnsPerHost"`
		IdleConnTimeout       time.Duration `json:"idleConnTimeout" toml:"idleConnTimeout"`
		DisableKeepAlives     bool          `json:"disableKeepAlives" toml:"disableKeepAlives"`
		MaxConnsPerHost       int           `json:"maxConnsPerHost" toml:"maxConnsPerHost"`
		ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout" toml:"responseHeaderTimeout"`
		// 配置名称
		Name string `json:"name"`
		// Debug 开关
		Debug bool `json:"debug" toml:"debug"`
		// 指标采集开关
		EnableMetric bool `json:"enableMetric" toml:"enableMetric"`
		// 访问日志开关
		EnableAccessLog bool `json:"enableAccessLog" toml:"enableAccessLog"`
		// 慢日志阈值
		SlowThreshold time.Duration `json:"slowThreshold" toml:"slowThreshold"`
		// 创建连接的错误级别，=panic时，如果创建失败，立即panic
		OnDialError string `json:"onDialError" toml:"onDialError"`
		// 重试次数
		Retry int `json:"retry" toml:"retry"`
		// 重试等待时间
		RetryWaitTime time.Duration `json:"retryWaitTime" toml:"retryWaitTime"`
	}
)

func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("search." + name))
}

// RawConfig 返回配置
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config.Name = key

	if err := conf.UnmarshalKey(key, &config, conf.TagName("toml")); err != nil {
		xlog.Jupiter().Panic("unmarshal config", xlog.FieldName(key), xlog.FieldExtMessage(config))
	}

	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint(key, config)
	}
	return config
}

func DefaultConfig() *Config {
	return &Config{
		Addr:                  "",
		Timeout:               cast.ToDuration("3000ms"),
		KeepAlive:             cast.ToDuration("30s"),
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       cast.ToDuration("30s"),
		DisableKeepAlives:     false,
		MaxConnsPerHost:       100,
		ResponseHeaderTimeout: cast.ToDuration("30s"),
		Name:                  "default",
		Debug:                 false,
		EnableMetric:          true,
		EnableAccessLog:       false,
		SlowThreshold:         cast.ToDuration("1s"),
		OnDialError:           "panic",
		Retry:                 2,
		RetryWaitTime:         cast.ToDuration("200ms"),
	}
}

func (config *Config) Build() (*Client, error) {
	if config.Addr == "" {
		return nil, errors.New("no addr found")
	}

	// 创建自定义传输层
	transport := NewMeiliSearchRoundTripper(config.Name, config)
	var client = http.Client{Transport: transport}

	meiliClient := meilisearch.New(config.Addr,
		meilisearch.WithAPIKey(config.ApiKey),
		meilisearch.WithCustomClient(&client),
	)

	// 返回包装后的客户端
	return &Client{ServiceManager: meiliClient}, nil
}

func (config *Config) Singleton() (*Client, error) {
	if client, ok := singleton.Load(constant.ModuleClientSearch, config.Name); ok && client != nil {
		return client.(*Client), nil
	}

	client, err := config.Build()
	if err != nil {
		xlog.Jupiter().Error("build meilisearch client failed", zap.Error(err))
		return nil, err
	}

	singleton.Store(constant.ModuleClientSearch, config.Name, client)
	return client, nil
}

func (config *Config) MustBuild() *Client {
	return lo.Must(config.Build())
}

func (config *Config) MustSingleton() *Client {
	return lo.Must(config.Singleton())
}
