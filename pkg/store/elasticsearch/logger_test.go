package elasticsearch

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// newObservedLogger 返回一个可断言输出的 esLogger 及其日志记录。
func newObservedLogger(config *Config) (*esLogger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return &esLogger{
		name:   "test",
		config: config,
		logger: zap.New(core),
	}, logs
}

func loggerConfig() *Config {
	config := DefaultConfig()
	config.EnableMetric = false
	config.EnableAccessLog = false
	config.SlowThreshold = 0
	return config
}

func newRequest(t *testing.T, method, rawURL string) *http.Request {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	return &http.Request{Method: method, URL: parsed}
}

// 接口约定要求实现必须容忍 req 与 resp 为 nil。
func TestESLogger_LogRoundTrip_ToleratesNilRequestAndResponse(t *testing.T) {
	config := loggerConfig()
	config.EnableAccessLog = true
	logger, logs := newObservedLogger(config)

	require.NotPanics(t, func() {
		err := logger.LogRoundTrip(nil, nil, nil, time.Now(), time.Millisecond)
		assert.NoError(t, err)
	})

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, zapcore.InfoLevel, entry.Level)
	assert.Equal(t, int32(-1), entry.ContextMap()["code"])
}

func TestESLogger_LogRoundTrip_AccessLogDisabledStaysQuietOnSuccess(t *testing.T) {
	logger, logs := newObservedLogger(loggerConfig())

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_doc/1")
	resp := &http.Response{StatusCode: http.StatusOK}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), time.Millisecond))

	assert.Zero(t, logs.Len(), "访问日志关闭且请求成功时不应产生日志")
}

func TestESLogger_LogRoundTrip_AccessLogRecordsRequestFields(t *testing.T) {
	config := loggerConfig()
	config.EnableAccessLog = true
	logger, logs := newObservedLogger(config)

	req := newRequest(t, http.MethodPost, "http://es.local:9200/books/_search")
	resp := &http.Response{StatusCode: http.StatusOK}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), 20*time.Millisecond))

	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	// FieldMethod 的键是 "meth"，为框架统一约定
	assert.Equal(t, http.MethodPost, fields["meth"])
	assert.Equal(t, "es.local:9200", fields["addr"])
	assert.Equal(t, "/books/_search", fields["path"])
	assert.Equal(t, int32(200), fields["code"])
}

// 传输错误必须记为 error，且不受访问日志开关影响。
func TestESLogger_LogRoundTrip_LogsTransportError(t *testing.T) {
	logger, logs := newObservedLogger(loggerConfig())

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_doc/1")

	require.NoError(t, logger.LogRoundTrip(req, nil, errors.New("connection refused"), time.Now(), time.Millisecond))

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, zapcore.ErrorLevel, entry.Level)
	assert.Equal(t, "connection refused", entry.ContextMap()["error"])
}

// 5xx 必须记为 error，即便访问日志关闭。
func TestESLogger_LogRoundTrip_LogsServerError(t *testing.T) {
	logger, logs := newObservedLogger(loggerConfig())

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_search")
	resp := &http.Response{StatusCode: http.StatusServiceUnavailable}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), time.Millisecond))

	require.Equal(t, 1, logs.Len())
	assert.Equal(t, zapcore.ErrorLevel, logs.All()[0].Level)
}

// 4xx 属于业务语义（如文档不存在），访问日志关闭时不应记录。
func TestESLogger_LogRoundTrip_ClientErrorNotLoggedWhenAccessLogDisabled(t *testing.T) {
	logger, logs := newObservedLogger(loggerConfig())

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_doc/404")
	resp := &http.Response{StatusCode: http.StatusNotFound}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), time.Millisecond))

	assert.Zero(t, logs.Len())
}

func TestESLogger_LogRoundTrip_SlowQuery(t *testing.T) {
	config := loggerConfig()
	config.SlowThreshold = 10 * time.Millisecond
	logger, logs := newObservedLogger(config)

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_search")
	resp := &http.Response{StatusCode: http.StatusOK}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), 50*time.Millisecond))

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "slow", entry.Message)
	assert.Equal(t, zapcore.WarnLevel, entry.Level)
	assert.Equal(t, ErrSlowQuery.Error(), entry.ContextMap()["error"])
}

func TestESLogger_LogRoundTrip_NotSlowUnderThreshold(t *testing.T) {
	config := loggerConfig()
	config.SlowThreshold = time.Second
	logger, logs := newObservedLogger(config)

	req := newRequest(t, http.MethodGet, "http://es.local:9200/books/_search")
	resp := &http.Response{StatusCode: http.StatusOK}

	require.NoError(t, logger.LogRoundTrip(req, resp, nil, time.Now(), 5*time.Millisecond))

	assert.Zero(t, logs.Len())
}

// 请求体复制有内存开销，仅调试模式下才应开启。
func TestESLogger_BodyEnabledFollowsDebug(t *testing.T) {
	config := loggerConfig()
	logger, _ := newObservedLogger(config)

	config.Debug = false
	assert.False(t, logger.RequestBodyEnabled())
	assert.False(t, logger.ResponseBodyEnabled())

	config.Debug = true
	assert.True(t, logger.RequestBodyEnabled())
	assert.True(t, logger.ResponseBodyEnabled())
}

func TestStatusText(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       string
	}{
		{"传输错误", 0, errors.New("boom"), "error"},
		{"错误优先于状态码", http.StatusOK, errors.New("boom"), "error"},
		{"无响应", -1, nil, "unknown"},
		{"正常状态码", http.StatusOK, nil, "200"},
		{"服务端错误状态码", http.StatusBadGateway, nil, "502"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, statusText(tt.statusCode, tt.err))
		})
	}
}

// 端到端确认客户端确实会回调到我们的 Logger（即 Logger 已正确装配进官方客户端）。
func TestClient_LoggerIsWiredIntoOfficialClient(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","found":true,"_source":{"title":"Go"}}`))
	})

	config := testConfig("logger-wired", srv.URL)
	config.EnableAccessLog = true

	core, logs := observer.New(zapcore.DebugLevel)
	config.logger = zap.New(core)

	client, err := config.Build()
	require.NoError(t, err)

	var book struct {
		Title string `json:"title"`
	}
	require.NoError(t, client.GetSource(t.Context(), "books", "1", &book))

	entries := logs.FilterMessage("access").All()
	require.NotEmpty(t, entries, "官方客户端未回调 esLogger，Logger 装配有误")

	var sawDocRequest bool
	for _, entry := range entries {
		if entry.ContextMap()["path"] == "/books/_doc/1" {
			sawDocRequest = true
			assert.Equal(t, int32(200), entry.ContextMap()["code"])
		}
	}
	assert.True(t, sawDocRequest, "未记录到取文档请求，实际日志：%v", entries)
}
