// Copyright 2026 Douyu
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

package xhertz

import (
	"context"
	"net"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	hzserver "github.com/cloudwego/hertz/pkg/app/server"
	hzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/weijinwen/jupiter/pkg/core/constant"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/server"
	"github.com/weijinwen/jupiter/pkg/util/xnet"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

// Server ...
type Server struct {
	*hzserver.Hertz
	config   *Config
	listener net.Listener
}

func newServer(config *Config) *Server {
	listener, err := net.Listen("tcp", config.Address())
	if err != nil {
		config.logger.Panic("new xhertz server err",
			xlog.FieldErrKind(ecode.ErrKindListenErr), xlog.FieldErr(err))
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port

	// 自定义 listener 必须搭配 standard transporter，netpoll 不支持外部 listener。
	options := []hzconfig.Option{
		hzserver.WithListener(listener),
		hzserver.WithTransport(standard.NewTransporter),
		hzserver.WithReadTimeout(config.ReadTimeout),
		hzserver.WithWriteTimeout(config.WriteTimeout),
		hzserver.WithIdleTimeout(config.IdleTimeout),
	}
	if config.ShutdownTimeout > 0 {
		options = append(options, hzserver.WithExitWaitTime(config.ShutdownTimeout))
	}
	if config.enableTLS() {
		// WithTLS 会自动切换到 standard 传输层（netpoll 不支持 TLS）
		options = append(options, hzserver.WithTLS(config.loadTLS()))
	}
	h := hzserver.New(options...)
	newHertzLogger(config.logger, config.Mode)
	return &Server{
		Hertz:    h,
		config:   config,
		listener: listener,
	}
}

// Upgrade protocol to WebSocket
func (s *Server) Upgrade(ws *WebSocket) {
	s.GET(ws.Pattern, func(ctx context.Context, rc *hzapp.RequestContext) {
		ws.Upgrade(rc)
	})
}

// Serve implements server.Server interface.
func (s *Server) Serve() error {
	for _, route := range s.Hertz.Routes() {
		s.config.logger.Info("add route", xlog.FieldMethod(route.Method), xlog.String("path", route.Path))
	}
	return s.Hertz.Run()
}

// Stop implements server.Server interface
// it will terminate hertz server immediately
func (s *Server) Stop() error {
	return s.Hertz.Close()
}

// GracefulStop implements server.Server interface
// it will stop hertz server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Hertz.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	scheme := "http"
	if s.config.enableTLS() {
		scheme = "https"
	}
	info := server.ApplyOptions(
		server.WithScheme(scheme),
		server.WithAddress(xnet.Address(s.listener)),
		server.WithKind(constant.ServiceProvider),
		server.WithMetaData("scheme", scheme),
	)
	return &info
}

// Healthz ...
func (s *Server) Healthz() bool {
	return true
}
