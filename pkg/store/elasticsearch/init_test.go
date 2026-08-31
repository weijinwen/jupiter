package elasticsearch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/weijinwen/jupiter/pkg/server/governor"
)

func TestStats_ReportsClusterHealth(t *testing.T) {
	srv := esStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"cluster_name":"test","status":"green","number_of_nodes":3,
			"active_shards":10,"active_shards_percent_as_number":100.0,
			"timed_out":false,"number_of_data_nodes":3,"active_primary_shards":5,
			"relocating_shards":0,"initializing_shards":0,"unassigned_shards":0,
			"delayed_unassigned_shards":0,"number_of_pending_tasks":0,
			"number_of_in_flight_fetch":0,"task_max_waiting_in_queue_millis":0
		}`))
	})

	name := "stats-health"
	_, err := testConfig(name, srv.URL).Build()
	require.NoError(t, err)
	t.Cleanup(func() { instances.Delete(name) })

	// Build 应已把实例注册进 instances
	_, ok := instances.Load(name)
	require.True(t, ok, "Build 未把客户端注册到 instances")

	result := stats(t.Context())

	entry, ok := result[name].(map[string]any)
	require.True(t, ok, "stats 未返回该实例，实际为 %#v", result[name])
	assert.Equal(t, "test", entry["cluster_name"])
	assert.Equal(t, "green", entry["status"])
	assert.Equal(t, 3, entry["number_of_nodes"])
}

func TestStats_ReportsErrorPerInstance(t *testing.T) {
	name := "stats-error"
	config := testConfig(name, "http://127.0.0.1:1")
	// 绕过 Build 的探活直接注册，模拟运行期集群不可达
	client, err := newClientForTest(config)
	require.NoError(t, err)

	instances.Store(name, client)
	t.Cleanup(func() { instances.Delete(name) })

	result := stats(t.Context())

	entry, ok := result[name].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, entry, "error")
}

func TestStats_SkipsUnexpectedValueTypes(t *testing.T) {
	instances.Store("bogus-entry", "not a client")
	t.Cleanup(func() { instances.Delete("bogus-entry") })

	result := stats(t.Context())

	assert.NotContains(t, result, "bogus-entry")
}

func TestGovernorStatsEndpointIsRegistered(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/elasticsearch/stats", nil)
	rec := httptest.NewRecorder()

	governor.DefaultServeMux.ServeHTTP(rec, req)

	// 端点在 init 中注册到 governor，注册失败会得到 404
	require.NotEqual(t, http.StatusNotFound, rec.Code)
	assert.True(t, json.Valid(rec.Body.Bytes()), "响应不是合法 JSON：%s", rec.Body.String())
}
