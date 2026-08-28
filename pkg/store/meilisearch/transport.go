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

package meilisearch

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	prome "github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/util/xretry"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

var knownErrorPrefixes = []string{"meilisearch", "request", "search", "server"}

func getStatement(err string) string {
	for _, prefix := range knownErrorPrefixes {
		if strings.HasPrefix(err, prefix) {
			if idx := strings.Index(err, ":"); idx > 0 {
				return err[:idx]
			}
			return prefix
		}
	}
	return "Unknown"
}

// MeiliSearchRoundTripper 自定义 HTTP 传输层，用于记录指标和日志
type MeiliSearchRoundTripper struct {
	http.Transport
	name   string
	config *Config
}

// NewMeiliSearchRoundTripper 创建新的 MeiliSearch 传输层
func NewMeiliSearchRoundTripper(name string, config *Config) *MeiliSearchRoundTripper {
	return &MeiliSearchRoundTripper{
		name:   name,
		config: config,
		Transport: http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   config.Timeout,
				KeepAlive: config.KeepAlive,
			}).DialContext,
			MaxIdleConns:          config.MaxIdleConns,
			MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
			IdleConnTimeout:       config.IdleConnTimeout,
			DisableKeepAlives:     config.DisableKeepAlives,
			MaxConnsPerHost:       config.MaxConnsPerHost,
			ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		},
	}
}

func (h *MeiliSearchRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		addr    = req.URL.String()
		hostURL = strings.TrimLeft(req.URL.Path, "/")
		method  = req.Method
		beg     = time.Now()
	)

	var resp *http.Response
	var err error

	err = xretry.Do(h.config.Retry, h.config.RetryWaitTime, func() error {
		reqCopy := req.Clone(req.Context())

		resp, err = h.Transport.RoundTrip(reqCopy)
		if err != nil {
			if h.config.EnableMetric {
				prome.ClientHandleCounter.WithLabelValues(prome.TypeMeiliSearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
			}
			return errors.New("request failed: " + err.Error())
		}

		if resp.StatusCode >= 500 {
			err = fmt.Errorf("server error with status code: %d", resp.StatusCode)
			if h.config.EnableMetric {
				prome.ClientHandleCounter.WithLabelValues(prome.TypeMeiliSearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
			}
			return errors.New("server error: " + err.Error())
		}

		return nil
	})

	// 如果重试后仍然失败，记录错误
	if err != nil {
		xlog.Jupiter().Error("meilisearch request failed after retries",
			xlog.FieldErr(err),
			xlog.FieldMethod(method),
			xlog.FieldAddr(addr),
			xlog.Int("retryCount", h.config.Retry),
		)
	}

	var cost = time.Since(beg)
	// 指标采集
	if h.config.EnableMetric {
		if err != nil {
			prome.ClientHandleCounter.WithLabelValues(prome.TypeMeiliSearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
		} else {
			prome.ClientHandleCounter.WithLabelValues(prome.TypeMeiliSearch, h.name, method, hostURL, resp.Status).Inc()
		}
		prome.ClientHandleHistogram.WithLabelValues(prome.TypeMeiliSearch, h.name, method, hostURL).Observe(cost.Seconds())
	}

	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}

	if xdebug.IsDevelopmentMode() {
		prefix := fmt.Sprintf("[%s:%s]", h.name, addr)
		fmt.Printf("%-50s => %s\n", color.GreenString(prefix), color.GreenString("Send: "+method+" | "+req.URL.Path))
	}

	// 访问日志
	if err != nil {
		xlog.Jupiter().Error("access", xlog.FieldErr(err), xlog.FieldMethod(method), xlog.FieldAddr(addr), xlog.FieldCode(int32(statusCode)), xlog.FieldCost(cost))
	} else {
		if h.config.EnableAccessLog {
			xlog.Jupiter().Info("access", xlog.FieldMethod(method), xlog.FieldAddr(addr), xlog.FieldCost(cost), xlog.FieldCode(int32(statusCode)))
		}
	}

	// 慢查询日志
	if h.config.SlowThreshold > time.Duration(0) {
		if cost > h.config.SlowThreshold {
			xlog.Jupiter().Error("slow query", xlog.FieldErr(ErrSlowQuery), xlog.FieldMethod(method), xlog.FieldCost(cost), xlog.FieldAddr(addr), xlog.FieldCode(int32(statusCode)))
		}
	}
	return resp, err
}
