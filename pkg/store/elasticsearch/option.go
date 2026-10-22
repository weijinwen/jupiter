// Copyright 2023 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"errors"
	"os"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/singleton"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"

	elastic "github.com/elastic/go-elasticsearch/v8"
)

type (
	// Config Elasticsearch 客户端配置
	Config struct {
		// 配置名称
		Name string `json:"name"`
		// 目标服务地址，支持多节点
		Addresses []string `json:"addresses" toml:"addresses"`
		// 单个目标服务地址，Addresses 为空时生效，便于简单场景只配一个地址
		Addr string `json:"addr" toml:"addr"`
		// 用户名
		Username string `json:"username" toml:"username"`
		// 密码
		Password string `json:"password" toml:"password"`
		// APIKey 与 Username/Password 互斥，优先级更高
		APIKey string `json:"apiKey" toml:"apiKey"`
		// CACertFile 自定义 CA 证书的 PEM 文件路径
		CACertFile string `json:"caCertFile" toml:"caCertFile"`
		// 请求超时时间
		Timeout time.Duration `json:"timeout" toml:"timeout"`
		// 保持连接时间
		KeepAlive time.Duration `json:"keepAlive" toml:"keepAlive"`
		// 最大空闲连接数
		MaxIdleConns int `json:"maxIdleConns" toml:"maxIdleConns"`
		// 每个主机的最大空闲连接数
		MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost" toml:"maxIdleConnsPerHost"`
		// 空闲连接超时时间
		IdleConnTimeout time.Duration `json:"idleConnTimeout" toml:"idleConnTimeout"`
		// 是否禁用长连接
		DisableKeepAlives bool `json:"disableKeepAlives" toml:"disableKeepAlives"`
		// 每个主机的最大连接数
		MaxConnsPerHost int `json:"maxConnsPerHost" toml:"maxConnsPerHost"`
		// 响应头超时时间
		ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout" toml:"responseHeaderTimeout"`
		// Debug 开关
		Debug bool `json:"debug" toml:"debug"`
		// 指标采集开关
		EnableMetric bool `json:"enableMetric" toml:"enableMetric"`
		// 访问日志开关
		EnableAccessLog bool `json:"enableAccessLog" toml:"enableAccessLog"`
		// 慢查询阈值，0 表示不记录
		SlowThreshold time.Duration `json:"slowThreshold" toml:"slowThreshold"`
		// 创建连接的错误级别，=panic 时，如果创建失败，立即 panic
		OnDialError string `json:"onDialError" toml:"onDialError"`
		// 重试次数
		Retry int `json:"retry" toml:"retry"`
		// 重试等待时间
		RetryWaitTime time.Duration `json:"retryWaitTime" toml:"retryWaitTime"`
	}
)

// StdConfig 返回标准配置，键为 jupiter.es.{name}
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("es." + name))
}

// RawConfig 返回配置，key 为完整的配置键（不含前缀拼接）
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config.Name = key

	if err := conf.UnmarshalKey(key, config, conf.TagName("toml")); err != nil {
		xlog.Jupiter().Panic("unmarshal config", xlog.FieldName(key), xlog.FieldExtMessage(config))
	}

	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint(key, config)
	}

	return config
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Addresses:             nil,
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

// addrs 归一化目标地址：优先 Addresses，其次 Addr
func (config *Config) addrs() []string {
	if len(config.Addresses) > 0 {
		return config.Addresses
	}
	if config.Addr != "" {
		return []string{config.Addr}
	}
	return nil
}

// Build 创建 Elasticsearch 客户端
func (config *Config) Build() (*Client, error) {
	addrs := config.addrs()
	if len(addrs) == 0 {
		return nil, errors.New("no addr found")
	}

	esConfig := elastic.Config{
		Addresses: addrs,
		Username:  config.Username,
		Password:  config.Password,
		APIKey:    config.APIKey,
		// 重试统一由自定义 RoundTripper 处理，关闭官方传输层重试，避免双层重试放大流量
		DisableRetry: true,
		// 指标、慢日志、访问日志都在 RoundTripper 内完成
		Transport: NewElasticsearchRoundTripper(config.Name, config),
	}

	if config.CACertFile != "" {
		cert, err := os.ReadFile(config.CACertFile)
		if err != nil {
			return nil, err
		}
		esConfig.CACert = cert
	}

	es, err := elastic.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	return &Client{Client: es, config: config}, nil
}

// Singleton 返回单例客户端
func (config *Config) Singleton() (*Client, error) {
	if client, ok := singleton.Load(constant.ModuleStoreElasticsearch, config.Name); ok && client != nil {
		return client.(*Client), nil
	}

	client, err := config.Build()
	if err != nil {
		xlog.Jupiter().Error("build elasticsearch client failed", zap.Error(err))
		return nil, err
	}

	singleton.Store(constant.ModuleStoreElasticsearch, config.Name, client)
	return client, nil
}

// MustBuild 创建客户端，失败时 panic
func (config *Config) MustBuild() *Client {
	return lo.Must(config.Build())
}

// MustSingleton 返回单例客户端，失败时 panic
func (config *Config) MustSingleton() *Client {
	return lo.Must(config.Singleton())
}
