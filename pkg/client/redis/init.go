package redis

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	prome "github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/server/governor"
)

var instances = sync.Map{}

type storeRedis struct {
	Client redis.UniversalClient
}

func init() {
	governor.HandleFunc("/debug/redis/stats", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(stats())
	})
	go monitor()
}
func monitor() {
	for {
		instances.Range(func(key, val any) bool {
			name := key.(string)
			obj := val.(*storeRedis)
			var poolStats *redis.PoolStats
			poolStats = obj.Client.PoolStats()
			if poolStats != nil {
				prome.ClientStatsGauge.Set(float64(poolStats.Hits), prome.TypeRedis, name, "hits")
				prome.ClientStatsGauge.Set(float64(poolStats.Misses), prome.TypeRedis, name, "misses")
				prome.ClientStatsGauge.Set(float64(poolStats.Timeouts), prome.TypeRedis, name, "timeouts")
				prome.ClientStatsGauge.Set(float64(poolStats.TotalConns), prome.TypeRedis, name, "total_conns")
				prome.ClientStatsGauge.Set(float64(poolStats.IdleConns), prome.TypeRedis, name, "idle_conns")
				prome.ClientStatsGauge.Set(float64(poolStats.StaleConns), prome.TypeRedis, name, "stale_conns")
			}
			return true
		})
		time.Sleep(time.Second * 10)
	}
}

// stats
func stats() (stats map[string]any) {
	stats = make(map[string]any)
	instances.Range(func(key, val any) bool {
		name := key.(string)
		obj := val.(*storeRedis)
		stats[name] = obj.Client.PoolStats()
		return true
	})
	return
}
