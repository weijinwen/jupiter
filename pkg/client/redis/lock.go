package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// Lock 获取锁，阻塞直到获取成功或上下文取消
	Lock(ctx context.Context) error
	// TryLock 尝试获取锁，非阻塞
	TryLock(ctx context.Context) (bool, error)
	// Unlock 释放锁
	Unlock(ctx context.Context) error
	// Renew 续约锁，延长锁的过期时间
	Renew(ctx context.Context, expiration time.Duration) error
	// IsLocked 检查锁是否被当前实例持有
	IsLocked() bool
}

// LockOptions 锁选项配置
type LockOptions struct {
	// Expiration 锁的过期时间，默认为30秒
	Expiration time.Duration
	// RetryDelay 获取锁失败后的重试间隔，默认为100毫秒
	RetryDelay time.Duration
	// MaxRetries 最大重试次数，-1表示无限重试，默认为-1
	MaxRetries int
	// Timeout 获取锁的总超时时间，0表示不超时，默认为0
	Timeout time.Duration
}

// DefaultLockOptions 返回默认的锁选项
func (rc *Client) DefaultLockOptions() *LockOptions {
	// 尝试从配置中获取默认值
	var expiration time.Duration = 30 * time.Second
	var retryDelay time.Duration = 100 * time.Millisecond
	var maxRetries int = -1
	var timeout time.Duration = 0

	// 尝试获取配置
	if config, err := rc.getConfig(); err == nil {
		if config.LockDefaultExpiration > 0 {
			expiration = config.LockDefaultExpiration
		}
		if config.LockDefaultRetryDelay > 0 {
			retryDelay = config.LockDefaultRetryDelay
		}
		if config.LockDefaultMaxRetries >= -1 {
			maxRetries = config.LockDefaultMaxRetries
		}
		if config.LockDefaultTimeout >= 0 {
			timeout = config.LockDefaultTimeout
		}
	}

	return &LockOptions{
		Expiration: expiration,
		RetryDelay: retryDelay,
		MaxRetries: maxRetries,
		Timeout:    timeout,
	}
}

// DefaultLockOptions 返回默认的锁选项（全局函数）
func DefaultLockOptions() *LockOptions {
	return &LockOptions{
		Expiration: 30 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		MaxRetries: -1, // 无限重试
		Timeout:    0,  // 不超时
	}
}

// distributedLock 分布式锁实现
type distributedLock struct {
	client     *Client
	key        string
	value      string
	options    *LockOptions
	mu         sync.Mutex
	logger     *zap.Logger
	renewTimer *time.Timer
}

// NewDistributedLock 创建一个新的分布式锁实例
func (rc *Client) NewDistributedLock(key string, opts ...*LockOptions) DistributedLock {
	options := rc.DefaultLockOptions()
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].Expiration > 0 {
			options.Expiration = opts[0].Expiration
		}
		if opts[0].RetryDelay > 0 {
			options.RetryDelay = opts[0].RetryDelay
		}
		if opts[0].MaxRetries >= -1 {
			options.MaxRetries = opts[0].MaxRetries
		}
		if opts[0].Timeout >= 0 {
			options.Timeout = opts[0].Timeout
		}
	}

	// 生成唯一的锁值，用于标识锁的持有者
	value, err := generateUniqueValue()
	if err != nil {
		xlog.Jupiter().Error("生成锁唯一值失败", xlog.FieldErr(err))
		// 如果生成唯一值失败，使用时间戳作为后备方案
		value = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return &distributedLock{
		client:  rc,
		key:     key,
		value:   value,
		options: options,
		logger:  xlog.Jupiter().Named(ecode.ModClientRedis).With(zap.String("lock_key", key)),
	}
}

// Lock 获取锁，阻塞直到获取成功或上下文取消
func (dl *distributedLock) Lock(ctx context.Context) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 如果已经持有锁，直接返回成功
	if dl.isLockedInternal() {
		return nil
	}

	// 设置超时上下文
	var cancel context.CancelFunc
	if dl.options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, dl.options.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	retryCount := 0
	for {
		select {
		case <-ctx.Done():
			dl.logger.Error("获取锁超时或取消",
				zap.Int("retry_count", retryCount),
				zap.Error(ctx.Err()))
			return fmt.Errorf("获取锁超时或取消: %w", ctx.Err())
		default:
			// 尝试获取锁
			success, err := dl.tryLock(ctx)
			if err != nil {
				dl.logger.Error("尝试获取锁失败", zap.Error(err))
				return fmt.Errorf("尝试获取锁失败: %w", err)
			}

			if success {
				dl.logger.Info("成功获取分布式锁")
				return nil
			}

			// 检查重试次数
			if dl.options.MaxRetries >= 0 && retryCount >= dl.options.MaxRetries {
				dl.logger.Error("达到最大重试次数", zap.Int("max_retries", dl.options.MaxRetries))
				return errors.New("达到最大重试次数")
			}

			// 等待重试间隔
			select {
			case <-time.After(dl.options.RetryDelay):
				retryCount++
				continue
			case <-ctx.Done():
				dl.logger.Error("获取锁超时或取消",
					zap.Int("retry_count", retryCount),
					zap.Error(ctx.Err()))
				return fmt.Errorf("获取锁超时或取消: %w", ctx.Err())
			}
		}
	}
}

// TryLock 尝试获取锁，非阻塞
func (dl *distributedLock) TryLock(ctx context.Context) (bool, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 如果已经持有锁，直接返回成功
	if dl.isLockedInternal() {
		return true, nil
	}

	return dl.tryLock(ctx)
}

// tryLock 内部实现，实际执行获取锁操作
func (dl *distributedLock) tryLock(ctx context.Context) (bool, error) {
	// 使用SET命令的NX和EX选项实现原子性获取锁
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.options.Expiration).Result()
	if err != nil {
		return false, fmt.Errorf("执行SET NX命令失败: %w", err)
	}

	if result {
		// 成功获取锁，启动自动续约
		dl.startAutoRenew(ctx)
	}

	return result, nil
}

// Unlock 释放锁
func (dl *distributedLock) Unlock(ctx context.Context) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 停止自动续约
	dl.stopAutoRenew()

	// 使用Lua脚本确保只有锁的持有者才能释放锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		dl.logger.Error("释放锁失败", zap.Error(err))
		return fmt.Errorf("释放锁失败: %w", err)
	}

	// 检查结果
	if result.(int64) == 0 {
		dl.logger.Warn("尝试释放不持有的锁或锁已过期")
		return errors.New("尝试释放不持有的锁或锁已过期")
	}

	dl.logger.Info("成功释放分布式锁")
	return nil
}

// Renew 续约锁，延长锁的过期时间
func (dl *distributedLock) Renew(ctx context.Context, expiration time.Duration) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 使用Lua脚本确保只有锁的持有者才能续约
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	// 将秒数转换为字符串
	expireSeconds := int(expiration.Seconds())
	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value, expireSeconds).Result()
	if err != nil {
		dl.logger.Error("续约锁失败", zap.Error(err))
		return fmt.Errorf("续约锁失败: %w", err)
	}

	// 检查结果
	if result.(int64) == 0 {
		dl.logger.Warn("尝试续约不持有的锁或锁已过期")
		return errors.New("尝试续约不持有的锁或锁已过期")
	}

	dl.logger.Info("成功续约分布式锁", zap.Duration("expiration", expiration))
	return nil
}

// IsLocked 检查锁是否被当前实例持有
func (dl *distributedLock) IsLocked() bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.isLockedInternal()
}

// isLockedInternal 内部实现，不加锁检查
func (dl *distributedLock) isLockedInternal() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 检查锁是否存在且值匹配
	value, err := dl.client.Get(ctx, dl.key).Result()
	if err != nil {
		// 锁不存在或获取失败
		return false
	}

	return value == dl.value
}

// startAutoRenew 启动自动续约
func (dl *distributedLock) startAutoRenew(ctx context.Context) {
	dl.stopAutoRenew() // 确保没有重复启动

	// 续约间隔为过期时间的1/3，确保有足够的安全边际
	renewInterval := dl.options.Expiration / 3
	if renewInterval < 1*time.Second {
		renewInterval = 1 * time.Second
	}

	dl.renewTimer = time.AfterFunc(renewInterval, func() {
		select {
		case <-ctx.Done():
			return
		default:
			// 尝试续约
			if err := dl.Renew(ctx, dl.options.Expiration); err != nil {
				dl.logger.Error("自动续约失败", zap.Error(err))
				// 续约失败，停止自动续约
				return
			}

			// 续约成功，继续下一次续约
			dl.startAutoRenew(ctx)
		}
	})
}

// stopAutoRenew 停止自动续约
func (dl *distributedLock) stopAutoRenew() {
	if dl.renewTimer != nil {
		dl.renewTimer.Stop()
		dl.renewTimer = nil
	}
}

// generateUniqueValue 生成唯一的锁值
func generateUniqueValue() (string, error) {
	// 生成16字节的随机数据
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 使用base64编码
	return base64.StdEncoding.EncodeToString(b), nil
}

// WithLock 使用分布式锁执行函数，自动处理锁的获取和释放
func (rc *Client) WithLock(ctx context.Context, key string, fn func() error, opts ...*LockOptions) error {
	lock := rc.NewDistributedLock(key, opts...)

	// 获取锁
	if err := lock.Lock(ctx); err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}

	// 确保锁被释放
	defer func() {
		if err := lock.Unlock(context.Background()); err != nil {
			// 记录错误但不中断程序执行
			xlog.Jupiter().Error("释放锁失败",
				zap.String("key", key),
				zap.Error(err))
		}
	}()

	// 执行函数
	return fn()
}

// TryWithLock 尝试使用分布式锁执行函数，非阻塞
func (rc *Client) TryWithLock(ctx context.Context, key string, fn func() error, opts ...*LockOptions) (bool, error) {
	lock := rc.NewDistributedLock(key, opts...)

	// 尝试获取锁
	acquired, err := lock.TryLock(ctx)
	if err != nil {
		return false, fmt.Errorf("尝试获取锁失败: %w", err)
	}

	if !acquired {
		return false, nil
	}

	// 确保锁被释放
	defer func() {
		if err := lock.Unlock(context.Background()); err != nil {
			// 记录错误但不中断程序执行
			xlog.Jupiter().Error("释放锁失败",
				zap.String("key", key),
				zap.Error(err))
		}
	}()

	// 执行函数
	err = fn()
	return true, err
}
