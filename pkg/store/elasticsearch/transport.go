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
	"bytes"
	"errors"
	"fmt"
	"io"
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

// knownErrorPrefixes 用于把错误信息收敛到有限的指标标签，避免 Prometheus 基数爆炸
var knownErrorPrefixes = []string{"elasticsearch", "request", "search", "server"}

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

// ElasticsearchRoundTripper 自定义 HTTP 传输层，用于记录指标和日志
type ElasticsearchRoundTripper struct {
	http.Transport
	name   string
	config *Config
}

// NewElasticsearchRoundTripper 创建新的 Elasticsearch 传输层
func NewElasticsearchRoundTripper(name string, config *Config) *ElasticsearchRoundTripper {
	return &ElasticsearchRoundTripper{
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

// prepareBody 在重试场景下把请求体缓存下来，使每次重试都能重放
// 官方传输层在 DisableRetry=true 且未开启请求体日志时不会设置 GetBody，
// 而重试由本层负责，因此这里自行补齐。
func (h *ElasticsearchRoundTripper) prepareBody(req *http.Request) error {
	if h.config.Retry <= 0 || req.Body == nil || req.Body == http.NoBody {
		return nil
	}
	if req.GetBody != nil {
		return nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	_ = req.Body.Close()

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	req.Body, err = req.GetBody()
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(body))
	return nil
}

// RoundTrip 实现 http.RoundTripper，内置重试、指标、慢查询与访问日志
func (h *ElasticsearchRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		addr    = req.URL.String()
		hostURL = strings.TrimLeft(req.URL.Path, "/")
		method  = req.Method
		beg     = time.Now()
	)

	if err := h.prepareBody(req); err != nil {
		return nil, fmt.Errorf("request body: %w", err)
	}

	var resp *http.Response
	var err error

	err = xretry.Do(h.config.Retry, h.config.RetryWaitTime, func() error {
		// 每次尝试都用独立副本，避免上一次尝试消耗掉请求体
		reqCopy := req.Clone(req.Context())
		if req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return errors.New("request failed: " + bodyErr.Error())
			}
			reqCopy.Body = body
		}

		resp, err = h.Transport.RoundTrip(reqCopy)
		if err != nil {
			if h.config.EnableMetric {
				prome.ClientHandleCounter.WithLabelValues(prome.TypeElasticsearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
			}
			return errors.New("request failed: " + err.Error())
		}

		if resp.StatusCode >= 500 {
			// 重试前必须释放连接，否则会泄漏
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			err = fmt.Errorf("server error with status code: %d", resp.StatusCode)
			if h.config.EnableMetric {
				prome.ClientHandleCounter.WithLabelValues(prome.TypeElasticsearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
			}
			return errors.New("server error: " + err.Error())
		}

		return nil
	})

	// 如果重试后仍然失败，记录错误
	if err != nil {
		xlog.Jupiter().Error("elasticsearch request failed after retries",
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
			prome.ClientHandleCounter.WithLabelValues(prome.TypeElasticsearch, h.name, method, hostURL, getStatement(err.Error())).Inc()
		} else {
			prome.ClientHandleCounter.WithLabelValues(prome.TypeElasticsearch, h.name, method, hostURL, resp.Status).Inc()
		}
		prome.ClientHandleHistogram.WithLabelValues(prome.TypeElasticsearch, h.name, method, hostURL).Observe(cost.Seconds())
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
