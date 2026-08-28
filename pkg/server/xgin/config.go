// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xgin

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/flag"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// ModName ..
const ModName = "server.gin"

// Config HTTP config
type Config struct {
	Host          string
	Port          int
	Deployment    string
	Mode          string
	DisableMetric bool
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string

	SlowQueryThresholdInMilli int64 `json:"slowQueryThresholdInMilli"`

	// ReadTimeout 读请求（含 body）超时，0 表示不限制
	ReadTimeout time.Duration
	// WriteTimeout 写响应超时，0 表示不限制
	WriteTimeout time.Duration
	// IdleTimeout keep-alive 空闲超时，0 表示不限制
	IdleTimeout time.Duration
	// CertFile TLS 证书路径，与 KeyFile 成对配置后启用 HTTPS
	CertFile string
	// KeyFile TLS 私钥路径，与 CertFile 成对配置后启用 HTTPS
	KeyFile string
	// ShutdownTimeout 优雅停止排空超时，0 表示沿用调用方传入 ctx 的期限
	ShutdownTimeout time.Duration
	// DisableCORS 是否关闭默认跨域中间件
	DisableCORS bool
	// Cors 跨域配置，默认全部授权
	Cors CORSConfig

	logger *xlog.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      flag.String("host"),
		Port:                      9091,
		Mode:                      gin.ReleaseMode,
		SlowQueryThresholdInMilli: 500, // 500ms
		Cors:                      DefaultCORSConfig(),
		logger:                    xlog.Jupiter().With(xlog.FieldMod(ModName)),
	}
}

// StdConfig Jupiter Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.FieldKey(key), xlog.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) MustBuild() *Server {
	server := newServer(config)
	server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))
	if !config.DisableCORS {
		server.Use(corsMiddleware(config.Cors))
	}

	if !config.DisableMetric {
		server.Use(metricServerInterceptor())
	}

	return server
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// enableTLS 报告是否配置了 TLS（两者需成对出现）
func (config *Config) enableTLS() bool {
	return config.CertFile != "" || config.KeyFile != ""
}

// loadTLS 在构建期加载并校验证书，失败即 panic（与 listen 失败语义一致）
func (config *Config) loadTLS() *tls.Config {
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		config.logger.Panic("load xgin tls cert err", xlog.FieldErrKind(ecode.ErrKindAny), xlog.FieldErr(err),
			xlog.FieldKey("certFile"), xlog.String("certFile", config.CertFile))
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}
}
