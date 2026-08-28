package redis

import (
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/weijinwen/jupiter/pkg/core/constant"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "default", config.name)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 200, config.PoolSize)
	assert.Equal(t, 20, config.MinIdleConns)
	assert.Equal(t, cast.ToDuration("3s"), config.DialTimeout)
	assert.Equal(t, cast.ToDuration("1s"), config.ReadTimeout)
	assert.Equal(t, cast.ToDuration("1s"), config.WriteTimeout)
	assert.Equal(t, cast.ToDuration("60s"), config.IdleTimeout)
	assert.False(t, config.Debug)
	assert.False(t, config.EnableMetric)
	assert.False(t, config.EnableAccessLog)
	assert.Equal(t, time.Duration(0), config.SlowLogThreshold)
	assert.Equal(t, "error", config.OnDialError)
}

func TestStdConfig(t *testing.T) {
	// 测试默认配置
	config := StdConfig("default")

	assert.NotNil(t, config)
	assert.Equal(t, constant.ConfigKey("redis", "default"), config.name)
}

func TestRawConfig(t *testing.T) {
	// 测试空键
	key := "redis.test"
	config := RawConfig(key)

	assert.NotNil(t, config)
	assert.Equal(t, key, config.name)
}

func TestNewRedis(t *testing.T) {
	// 测试创建Redis客户端
	client := NewRedis("default")

	assert.NotNil(t, client)
}

func TestConfig_Fields(t *testing.T) {
	config := &Config{
		Addr:             []string{"localhost:6379", "localhost:6380"},
		Username:         "testuser",
		Password:         "testpass",
		DB:               1,
		PoolSize:         100,
		MaxRetries:       5,
		MinIdleConns:     10,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
		IdleTimeout:      30 * time.Second,
		Debug:            true,
		SlowLogThreshold: 500 * time.Millisecond,
		EnableMetric:     false,
		EnableAccessLog:  true,
		OnDialError:      "error",
		logger:           zap.NewNop(),
		name:             "test",
	}

	assert.Equal(t, []string{"localhost:6379", "localhost:6380"}, config.Addr)
	assert.Equal(t, "testuser", config.Username)
	assert.Equal(t, "testpass", config.Password)
	assert.Equal(t, 1, config.DB)
	assert.Equal(t, 100, config.PoolSize)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10, config.MinIdleConns)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)
	assert.Equal(t, 30*time.Second, config.IdleTimeout)
	assert.True(t, config.Debug)
	assert.Equal(t, 500*time.Millisecond, config.SlowLogThreshold)
	assert.False(t, config.EnableMetric)
	assert.True(t, config.EnableAccessLog)
	assert.Equal(t, "error", config.OnDialError)
	assert.Equal(t, "test", config.name)
}
