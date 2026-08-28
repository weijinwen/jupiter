package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLock_Basic 测试分布式锁的基本功能
func TestLock_Basic(t *testing.T) {
	// 创建测试客户端
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_lock")
	}()

	// 测试获取和释放锁
	lock := client.NewDistributedLock("test_lock")
	ctx := context.Background()

	// 获取锁
	err := lock.Lock(ctx)
	assert.NoError(t, err)
	assert.True(t, lock.IsLocked())

	// 尝试再次获取锁，应该失败
	lock2 := client.NewDistributedLock("test_lock")
	acquired, err := lock2.TryLock(ctx)
	assert.NoError(t, err)
	assert.False(t, acquired)

	// 释放锁
	err = lock.Unlock(ctx)
	assert.NoError(t, err)
	assert.False(t, lock.IsLocked())

	// 现在第二个锁应该能够获取
	acquired, err = lock2.TryLock(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.True(t, lock2.IsLocked())

	// 清理
	err = lock2.Unlock(ctx)
	assert.NoError(t, err)
}

// TestLock_WithLock 测试 WithLock 方法
func TestLock_WithLock(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_with_lock")
	}()

	ctx := context.Background()
	var executed bool
	var mu sync.Mutex

	// 使用 WithLock 执行函数
	err := client.WithLock(ctx, "test_with_lock", func() error {
		mu.Lock()
		executed = true
		mu.Unlock()
		return nil
	})
	assert.NoError(t, err)

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()
}

// TestLock_TryWithLock 测试 TryWithLock 方法
func TestLock_TryWithLock(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_try_with_lock")
	}()

	ctx := context.Background()

	// 先获取一个锁
	lock1 := client.NewDistributedLock("test_try_with_lock")
	err := lock1.Lock(ctx)
	require.NoError(t, err)
	defer lock1.Unlock(ctx)

	// 尝试使用 TryWithLock，应该失败
	acquired, err := client.TryWithLock(ctx, "test_try_with_lock", func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.False(t, acquired)

	// 释放第一个锁
	err = lock1.Unlock(ctx)
	require.NoError(t, err)

	// 现在应该能够成功
	var executed bool
	acquired, err = client.TryWithLock(ctx, "test_try_with_lock", func() error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.True(t, executed)
}

// TestLock_Renew 测试锁续约功能
func TestLock_Renew(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_renew")
	}()

	ctx := context.Background()
	lock := client.NewDistributedLock("test_renew", &LockOptions{
		Expiration: 2 * time.Second,
	})

	// 获取锁
	err := lock.Lock(ctx)
	require.NoError(t, err)
	defer lock.Unlock(ctx)

	// 检查初始TTL
	ttl := client.MustTTL(ctx, "test_renew")
	assert.True(t, ttl > 0 && ttl <= 2*time.Second)

	// 等待一段时间
	time.Sleep(1 * time.Second)

	// 续约锁
	err = lock.Renew(ctx, 5*time.Second)
	assert.NoError(t, err)

	// 检查续约后的TTL
	ttl = client.MustTTL(ctx, "test_renew")
	assert.True(t, ttl > 3*time.Second && ttl <= 5*time.Second)
}

// TestLock_Concurrent 测试并发获取锁
func TestLock_Concurrent(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_concurrent")
	}()

	ctx := context.Background()
	var wg sync.WaitGroup
	var counter int
	var mu sync.Mutex

	// 启动多个goroutine竞争同一个锁
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.WithLock(ctx, "test_concurrent", func() error {
				mu.Lock()
				current := counter
				counter++
				mu.Unlock()

				// 模拟一些工作
				time.Sleep(100 * time.Millisecond)

				// 验证没有并发修改
				mu.Lock()
				assert.Equal(t, current, counter-1)
				mu.Unlock()

				return nil
			})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, 10, counter)
}

// TestLock_Timeout 测试锁超时
func TestLock_Timeout(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_timeout")
	}()

	ctx := context.Background()

	// 先获取一个锁
	lock1 := client.NewDistributedLock("test_timeout")
	err := lock1.Lock(ctx)
	require.NoError(t, err)
	defer lock1.Unlock(ctx)

	// 尝试获取锁，设置短超时
	lock2 := client.NewDistributedLock("test_timeout", &LockOptions{
		Timeout: 100 * time.Millisecond,
	})

	start := time.Now()
	err = lock2.Lock(ctx)
	elapsed := time.Since(start)

	// 应该超时
	assert.Error(t, err)
	assert.True(t, elapsed >= 100*time.Millisecond)
}

// TestLock_Expiration 测试锁过期
func TestLock_Expiration(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_expiration")
	}()

	ctx := context.Background()

	// 获取一个短过期时间的锁
	lock := client.NewDistributedLock("test_expiration", &LockOptions{
		Expiration: 500 * time.Millisecond,
	})

	err := lock.Lock(ctx)
	require.NoError(t, err)

	// 等待锁过期
	time.Sleep(600 * time.Millisecond)

	// 现在应该能够获取锁
	lock2 := client.NewDistributedLock("test_expiration")
	acquired, err := lock2.TryLock(ctx)
	assert.NoError(t, err)
	assert.True(t, acquired)

	// 清理
	err = lock2.Unlock(ctx)
	assert.NoError(t, err)
}

// TestLock_CustomOptions 测试自定义锁选项
func TestLock_CustomOptions(t *testing.T) {
	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: &Config{},
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_custom_options")
	}()

	ctx := context.Background()

	// 使用自定义选项创建锁
	customOptions := &LockOptions{
		Expiration: 5 * time.Second,
		RetryDelay: 200 * time.Millisecond,
		MaxRetries: 3,
		Timeout:    1 * time.Second,
	}

	lock := client.NewDistributedLock("test_custom_options", customOptions)
	err := lock.Lock(ctx)
	assert.NoError(t, err)
	assert.True(t, lock.IsLocked())

	// 检查TTL是否与自定义选项匹配
	ttl := client.MustTTL(ctx, "test_custom_options")
	assert.True(t, ttl > 4*time.Second && ttl <= 5*time.Second)

	// 清理
	err = lock.Unlock(ctx)
	assert.NoError(t, err)
}

// TestLock_ConfigDefaultOptions 测试配置默认选项
func TestLock_ConfigDefaultOptions(t *testing.T) {
	// 创建带有自定义默认配置的客户端
	customConfig := &Config{
		LockDefaultExpiration: 10 * time.Second,
		LockDefaultRetryDelay: 200 * time.Millisecond,
		LockDefaultMaxRetries: 5,
		LockDefaultTimeout:    2 * time.Second,
	}

	client := &Client{
		UniversalClient: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: customConfig,
	}

	// 清理测试环境
	defer func() {
		client.Del(context.Background(), "test_config_default")
	}()

	ctx := context.Background()

	// 不指定选项，应该使用配置的默认值
	lock := client.NewDistributedLock("test_config_default")
	err := lock.Lock(ctx)
	assert.NoError(t, err)
	assert.True(t, lock.IsLocked())

	// 检查TTL是否与配置默认值匹配
	ttl := client.MustTTL(ctx, "test_config_default")
	assert.True(t, ttl > 9*time.Second && ttl <= 10*time.Second)

	// 清理
	err = lock.Unlock(ctx)
	assert.NoError(t, err)
}
