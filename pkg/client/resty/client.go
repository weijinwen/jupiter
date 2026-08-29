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
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/singleton"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

// Client ...
type Client = resty.Client

// Build 组装 resty.Client：连接参数、重试策略、认证与默认请求头。
// 指标采集、访问日志、慢日志由 registerHooks（interceptor.go）统一挂载。
func (config *Config) Build() (*resty.Client, error) {
	if config.Addr == "" {
		return nil, errors.New("no addr found")
	}

	client := resty.New()
	client.SetBaseURL(config.Addr)
	client.SetTimeout(config.Timeout)
	client.SetDebug(config.Debug)
	client.SetRetryCount(config.RetryCount)
	client.SetCloseConnection(config.CloseConnection)
	client.SetRedirectPolicy(resty.NoRedirectPolicy())

	if config.RetryWaitTime != time.Duration(0) {
		client.SetRetryWaitTime(config.RetryWaitTime)
	}

	if config.RetryMaxWaitTime != time.Duration(0) {
		client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)
	}

	if config.RetryCondition != nil {
		client.AddRetryCondition(config.RetryCondition)
	}

	if len(config.Headers) > 0 {
		client.SetHeaders(config.Headers)
	}

	if config.AuthToken != "" {
		client.SetAuthToken(config.AuthToken)
	} else if config.Username != "" || config.Password != "" {
		client.SetBasicAuth(config.Username, config.Password)
	}

	config.registerHooks(client)

	return client, nil
}

// Singleton returns a singleton client conn.
func (config *Config) Singleton() (*Client, error) {
	if client, ok := singleton.Load(constant.ModuleClientResty, config.Name); ok && client != nil {
		return client.(*Client), nil
	}

	client, err := config.Build()
	if err != nil {
		xlog.Jupiter().Error("build resty client failed", zap.Error(err))
		return nil, err
	}

	singleton.Store(constant.ModuleClientResty, config.Name, client)

	return client, nil
}

// MustBuild panics when error found.
func (c *Config) MustBuild() *resty.Client {
	return lo.Must(c.Build())
}

// MustSingleton panics when error found.
func (config *Config) MustSingleton() *Client {
	return lo.Must(config.Singleton())
}
