package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"go.uber.org/zap"

	"github.com/weijinwen/jupiter/pkg/core/singleton"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

type Client struct {
	redis.UniversalClient
	config *Config
}

// Singleton returns a singleton client conn.
func (config *Config) Singleton() (*Client, error) {
	if val, ok := singleton.Load(constant.ModuleClientRedis, config.name); ok && val != nil {
		return val.(*Client), nil
	}

	cc, err := config.Build()
	if err != nil {
		xlog.Jupiter().Error("build redis client failed", zap.Error(err))
		return nil, err
	}
	singleton.Store(constant.ModuleClientRedis, config.name, cc)
	return cc, nil
}

// MustSingleton panics when error found.
func (config *Config) MustSingleton() *Client {
	return lo.Must(config.Singleton())
}

// MustBuild panics when error found.
func (config *Config) MustBuild() *Client {
	return lo.Must(config.Build())
}

// Build ..
func (config *Config) Build() (*Client, error) {
	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint("redis's config: "+config.name, config)
	}
	if len(config.Addr) <= 0 {
		return nil, errors.New("cluster redis addr is empty")
	}
	cc, err := config.build()
	if err != nil {
		xlog.Jupiter().Error("build redis cluster client failed", zap.Error(err))
		return nil, err
	}
	return cc, nil
}

func (config *Config) build() (*Client, error) {
	stubClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        config.Addr,
		Username:     config.Username,
		Password:     config.Password,
		DB:           config.DB,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})
	for _, addr := range config.Addr {
		stubClient.AddHook(fixedInterceptor(config.name, addr, config, config.logger))
		if config.EnableMetric {
			stubClient.AddHook(metricInterceptor(config.name, addr, config, config.logger))
		}
		if config.Debug {
			stubClient.AddHook(debugInterceptor(config.name, addr, config, config.logger))
		}

		if config.EnableAccessLog {
			stubClient.AddHook(accessInterceptor(config.name, addr, config, config.logger))
		}

	}
	if err := stubClient.Ping(context.Background()).Err(); err != nil {
		if config.OnDialError == "panic" {
			config.logger.Panic("redis stub client start err: " + err.Error())
		}
		config.logger.Error("redis stub client start err", xlog.FieldErr(err))
		return nil, err
	}

	instances.Store(config.name, &storeRedis{
		Client: stubClient,
	})
	return &Client{
		UniversalClient: stubClient,
		config:          config,
	}, nil
}

// getConfig 获取客户端配置
func (c *Client) getConfig() (*Config, error) {
	if c.config == nil {
		return nil, errors.New("客户端配置未初始化")
	}
	return c.config, nil
}
