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
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

var errSlowCommand = errors.New("http resty slow command")

// registerHooks 挂载指标采集、访问日志与慢日志相关的回调。
func (config *Config) registerHooks(client *resty.Client) {
	client.OnError(config.onError)

	client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		return nil
	})

	client.SetPreRequestHook(func(c *resty.Client, r *http.Request) error {
		return nil
	})

	client.OnAfterResponse(config.onAfterResponse)
}

// onError 处理请求执行失败（连接失败、超时等，未收到 HTTP 响应）的情况：指标 + 错误日志。
func (config *Config) onError(r *resty.Request, err error) {
	host := ""
	if r.RawRequest != nil {
		host = r.RawRequest.Host
	}

	if config.EnableMetric {
		metric.ClientHandleCounter.WithLabelValues(metric.TypeHTTP, config.Name, r.Method, host, "error").Inc()
	}

	config.logger.Error("access",
		xlog.FieldErr(err),
		xlog.FieldName(config.Name),
		xlog.FieldMethod(r.Method),
		xlog.FieldAddr(host),
	)
}

// onAfterResponse 处理已收到 HTTP 响应（含非 2xx）的情况：指标 + 访问日志 + 慢日志。
func (config *Config) onAfterResponse(c *resty.Client, r *resty.Response) error {
	cost := r.Time()

	if config.EnableMetric {
		metric.ClientHandleCounter.WithLabelValues(metric.TypeHTTP, config.Name, r.Request.Method, c.HostURL, r.Status()).Inc()
		metric.ClientHandleHistogram.WithLabelValues(metric.TypeHTTP, config.Name, r.Request.Method, c.HostURL).Observe(cost.Seconds())
	}

	if config.EnableAccessLog {
		config.logger.Info("access",
			xlog.FieldName(config.Name),
			xlog.FieldMethod(r.Request.Method),
			xlog.FieldAddr(r.Request.URL),
			xlog.FieldCode(int32(r.StatusCode())),
			xlog.FieldCost(cost),
		)
	}

	if config.SlowThreshold > time.Duration(0) && cost > config.SlowThreshold {
		config.logger.Error("slow",
			xlog.FieldErr(errSlowCommand),
			xlog.FieldName(config.Name),
			xlog.FieldMethod(r.Request.Method),
			xlog.FieldCost(cost),
			xlog.FieldAddr(r.Request.URL),
			xlog.FieldCode(int32(r.StatusCode())),
		)
	}

	return nil
}
