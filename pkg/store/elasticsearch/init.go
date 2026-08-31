package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/weijinwen/jupiter/pkg/server/governor"
)

var (
	// ErrEmptyAddresses 表示未配置任何节点地址
	ErrEmptyAddresses = errors.New("elasticsearch addresses is empty")

	// ErrConnectFailed 表示启动时连通性检查失败
	ErrConnectFailed = errors.New("elasticsearch connect failed")

	// ErrSlowQuery 表示请求耗时超过慢查询阈值
	ErrSlowQuery = errors.New("elasticsearch slow query")

	// ErrDocumentNotFound 表示文档不存在
	ErrDocumentNotFound = errors.New("elasticsearch document not found")
)

// instances 保存已构建的客户端，供 governor 调试端点查询集群状态
var instances = sync.Map{}

// statsTimeout 调试端点查询集群健康的超时
const statsTimeout = 3 * time.Second

func init() {
	governor.HandleFunc("/debug/elasticsearch/stats", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(stats(r.Context()))
	})
}

// stats 按需查询各实例的集群健康状态。
//
// 与 redis 不同，ES 官方客户端不暴露连接池统计，因此这里改为查询 cluster health；
// 也正因为它是一次真实的网络请求，故不像 redis 那样起后台协程轮询，只在访问端点时查询。
func stats(ctx context.Context) map[string]any {
	result := make(map[string]any)

	instances.Range(func(key, val any) bool {
		name := key.(string)
		client, ok := val.(*Client)
		if !ok {
			return true
		}

		callCtx, cancel := context.WithTimeout(ctx, statsTimeout)
		defer cancel()

		health, err := client.Cluster.Health().Do(callCtx)
		if err != nil {
			result[name] = map[string]any{"error": err.Error()}
			return true
		}

		result[name] = map[string]any{
			"cluster_name":          health.ClusterName,
			"status":                health.Status.String(),
			"number_of_nodes":       health.NumberOfNodes,
			"active_shards":         health.ActiveShards,
			"active_shards_percent": health.ActiveShardsPercentAsNumber,
		}
		return true
	})

	return result
}
