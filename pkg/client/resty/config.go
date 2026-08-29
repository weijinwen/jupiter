// Copyright 2022 Douyu
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

package resty

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

// Config ...
type (
	// Config options
	Config struct {
		// Debug 开关
		Debug bool `json:"debug" toml:"debug"` // debug开关
		// 指标采集开关
		EnableMetric bool `json:"enableMetric" toml:"enableMetric"` // 指标采集开关

		// 失败重试次数
		RetryCount int `json:"retryCount" toml:"retryCount"` // 重试次数
		// 失败重试的间隔时间
		RetryWaitTime time.Duration `json:"retryWaitTime" toml:"retryWaitTime"` // 重试间隔时间
		// 失败重试的最贱等待时间
		RetryMaxWaitTime time.Duration `json:"retryMaxWaitTime" toml:"retryMaxWaitTime"` // 重试最大间隔时间
		// 目标服务地址
		Addr string `json:"addr" toml:"addr"` // 目标地址
		// 请求超时时间
		Timeout time.Duration `json:"timeout" toml:"timeout" `
		// 收到响应以后是否立即关闭连接
		CloseConnection bool `json:"closeConnection" toml:"closeConnection" `
		// 慢日志阈值
		SlowThreshold time.Duration `json:"slowThreshold" toml:"slowThreshold"` // slowlog 时间阈
		// 访问日志开关
		EnableAccessLog bool `json:"enableAccessLog" toml:"enableAccessLog"`

		// 默认请求头，每次请求都会带上
		Headers map[string]string `json:"headers" toml:"headers"`
		// Bearer Token 认证；与 Username/Password 二选一，优先生效
		AuthToken string `json:"authToken" toml:"authToken"`
		// HTTP Basic 认证用户名
		Username string `json:"username" toml:"username"`
		// HTTP Basic 认证密码
		Password string `json:"password" toml:"password"`

		// 重试条件，返回 true 时触发重试，直到达到 RetryCount 次数
		RetryCondition resty.RetryConditionFunc `json:"-" toml:"-"`
		// 日志
		logger *zap.Logger
		// 配置名称
		Name string `json:"name"`
	}
)

// StdConfig 返回标准配置
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("resty." + name))
}

// RawConfig 返回配置
func RawConfig(key string) *Config {
	config := DefaultConfig()
	config.Name = key

	if err := conf.UnmarshalKey(key, &config, conf.TagName("toml")); err != nil {
		xlog.Jupiter().Panic("unmarshal config", xlog.FieldName(key), xlog.FieldExtMessage(config))
	}

	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint(key, config)
	}
	return config
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Debug:        false,
		EnableMetric: true,

		RetryCount:       0,
		RetryWaitTime:    cast.ToDuration("100ms"),
		RetryMaxWaitTime: cast.ToDuration("100ms"),
		Addr:             "",
		SlowThreshold:    cast.ToDuration("500ms"),
		Timeout:          cast.ToDuration("3000ms"),
		EnableAccessLog:  false,

		logger: xlog.Jupiter().Named(ecode.ModeClientResty),
		Name:   "default",
	}
}
