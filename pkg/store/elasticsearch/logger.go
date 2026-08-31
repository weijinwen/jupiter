package elasticsearch

import (
	"net/http"
	"strconv"
	"time"

	prome "github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// esLogger 实现 elastictransport.Logger，是 go-elasticsearch 官方提供的观测扩展点。
// 客户端在每次 round trip（含内部重试的每一次）结束后回调 LogRoundTrip，
// 并直接给出请求、响应、错误、起始时间与耗时，因此框架无需再包一层 RoundTripper。
type esLogger struct {
	name   string
	config *Config
	logger *xlog.Logger
}

// RequestBodyEnabled 仅在调试模式下要求客户端复制请求体，避免常态下的内存开销。
func (l *esLogger) RequestBodyEnabled() bool { return l.config.Debug }

// ResponseBodyEnabled 仅在调试模式下要求客户端复制响应体，避免常态下的内存开销。
func (l *esLogger) ResponseBodyEnabled() bool { return l.config.Debug }

// LogRoundTrip 记录指标、访问日志与慢查询日志。
// 按接口约定，实现必须容忍 req 与 resp 为 nil，且不得修改它们。
func (l *esLogger) LogRoundTrip(req *http.Request, resp *http.Response, err error, start time.Time, cost time.Duration) error {
	var method, path, addr string
	if req != nil {
		method = req.Method
		if req.URL != nil {
			path = req.URL.Path
			addr = req.URL.Host
		}
	}

	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}

	if l.config.EnableMetric {
		prome.ClientHandleHistogram.
			WithLabelValues(prome.TypeElasticsearch, l.name, method, addr).
			Observe(cost.Seconds())
		prome.ClientHandleCounter.
			WithLabelValues(prome.TypeElasticsearch, l.name, method, addr, statusText(statusCode, err)).
			Inc()
	}

	fields := []xlog.Field{
		xlog.FieldKey(l.name),
		xlog.FieldMethod(method),
		xlog.FieldAddr(addr),
		xlog.String("path", path),
		xlog.FieldCode(int32(statusCode)),
		xlog.FieldCost(cost),
	}

	switch {
	case err != nil:
		l.logger.Error("access", append(fields, xlog.FieldErr(err))...)
	case statusCode >= http.StatusInternalServerError:
		l.logger.Error("access", fields...)
	case l.config.EnableAccessLog:
		l.logger.Info("access", fields...)
	}

	if l.config.SlowThreshold > 0 && cost > l.config.SlowThreshold {
		l.logger.Warn("slow", append(fields, xlog.FieldErr(ErrSlowQuery))...)
	}

	return nil
}

// statusText 生成指标的 code 标签：请求失败时为 "error"，无响应时为 "unknown"，否则为状态码。
func statusText(statusCode int, err error) string {
	if err != nil {
		return "error"
	}
	if statusCode < 0 {
		return "unknown"
	}
	return strconv.Itoa(statusCode)
}
