package elasticsearch

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// esStub 启动一个最小的 ES 桩服务。根路径用于响应 Info 探活（客户端会做产品校验），
// 其余路由由 handler 决定。
func esStub(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 官方客户端依赖该响应头确认对端是 Elasticsearch，缺失会导致产品校验失败
		w.Header().Set("X-Elastic-Product", "Elasticsearch")

		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"stub","cluster_name":"test","version":{"number":"8.19.0"}}`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// testConfig 返回指向桩服务的配置，禁用重试以免测试放大请求次数。
func testConfig(name, addr string) *Config {
	config := DefaultConfig()
	config.name = name
	config.Addresses = []string{addr}
	config.DisableRetry = true
	config.OnDialError = "error"
	return config
}

// newClientForTest 跳过 Build 的探活直接构造客户端，
// 用于模拟"客户端已建好但运行期集群不可达"的场景。
func newClientForTest(config *Config) (*Client, error) {
	typedClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses:    config.Addresses,
		DisableRetry: true,
	})
	if err != nil {
		return nil, err
	}
	return &Client{TypedClient: typedClient, config: config}, nil
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, []string{"http://localhost:9200"}, config.Addresses)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "error", config.OnDialError)
	assert.True(t, config.EnableMetric)
	assert.Equal(t, time.Second, config.SlowThreshold)
}

func TestConfig_Build_EmptyAddresses(t *testing.T) {
	config := DefaultConfig()
	config.name = "empty-addr"
	config.Addresses = nil

	client, err := config.Build()

	assert.ErrorIs(t, err, ErrEmptyAddresses)
	assert.Nil(t, client)
}

func TestConfig_Build_ConnectFailed(t *testing.T) {
	config := testConfig("unreachable", "http://127.0.0.1:1")

	client, err := config.Build()

	assert.ErrorIs(t, err, ErrConnectFailed)
	assert.Nil(t, client)
}

func TestConfig_Build_Success(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client, err := testConfig("build-ok", srv.URL).Build()

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.TypedClient)
}

func TestConfig_Singleton_ReturnsSameInstance(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	config := testConfig("singleton-ok", srv.URL)

	first, err := config.Singleton()
	require.NoError(t, err)

	second, err := config.Singleton()
	require.NoError(t, err)

	assert.Same(t, first, second)
}

// buildLogger 返回 nil 时不得赋值给 Logger 接口字段，否则会构造出内含 nil 指针的非 nil 接口，
// 使客户端在每次请求时回调到空接收者并 panic。
func TestConfig_BuildLogger_NilWhenAllObservabilityDisabled(t *testing.T) {
	config := DefaultConfig()
	config.EnableMetric = false
	config.EnableAccessLog = false
	config.SlowThreshold = 0

	assert.Nil(t, config.buildLogger())
}

func TestConfig_BuildLogger_NotNilWhenAnySwitchEnabled(t *testing.T) {
	newConfig := func() *Config {
		config := DefaultConfig()
		config.EnableMetric = false
		config.EnableAccessLog = false
		config.SlowThreshold = 0
		return config
	}

	t.Run("EnableMetric", func(t *testing.T) {
		config := newConfig()
		config.EnableMetric = true
		assert.NotNil(t, config.buildLogger())
	})

	t.Run("EnableAccessLog", func(t *testing.T) {
		config := newConfig()
		config.EnableAccessLog = true
		assert.NotNil(t, config.buildLogger())
	})

	t.Run("SlowThreshold", func(t *testing.T) {
		config := newConfig()
		config.SlowThreshold = time.Millisecond
		assert.NotNil(t, config.buildLogger())
	})
}

// 观测开关全关时客户端仍须能正常发起请求（覆盖上面那个 nil 接口陷阱的实际路径）。
func TestConfig_Build_RequestSucceedsWithObservabilityDisabled(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"_index":"books","_id":"1","found":true,"_source":{"title":"Go"}}`))
	})

	config := testConfig("no-observability", srv.URL)
	config.EnableMetric = false
	config.EnableAccessLog = false
	config.SlowThreshold = 0

	client, err := config.Build()
	require.NoError(t, err)

	var book struct {
		Title string `json:"title"`
	}
	require.NoError(t, client.GetSource(t.Context(), "books", "1", &book))
	assert.Equal(t, "Go", book.Title)
}
