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
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

func recoverMiddleware(logger *xlog.Logger, slowQueryThresholdInMilli int64) hzapp.HandlerFunc {
	return func(c context.Context, ctx *hzapp.RequestContext) {
		beg := time.Now()
		fields := make([]xlog.Field, 0, 8)
		var brokenPipe bool
		defer func() {
			fields = append(fields, zap.Float64("cost", time.Since(beg).Seconds()))
			if slowQueryThresholdInMilli > 0 {
				if cost := int64(time.Since(beg)) / 1e6; cost > slowQueryThresholdInMilli {
					fields = append(fields, zap.Int64("slow", cost))
				}
			}
			if rec := recover(); rec != nil {
				if ne, ok := rec.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						msg := strings.ToLower(se.Error())
						if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				fields = append(fields, zap.ByteString("stack", debug.Stack()))
				fields = append(fields, zap.String("err", fmt.Sprint(rec)))
				logger.Error("access", fields...)
				if brokenPipe {
					_ = ctx.Error(fmt.Errorf("%v", rec))
					ctx.Abort()
					return
				}
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			fields = append(fields,
				zap.String("method", string(ctx.Method())),
				zap.Int("code", ctx.Response.StatusCode()),
				zap.Int("size", len(ctx.Response.Body())),
				zap.String("host", string(ctx.Host())),
				zap.String("path", string(ctx.Path())),
				zap.String("ip", ctx.ClientIP()),
			)
			logger.Info("access", fields...)
		}()
		ctx.Next(c)
	}
}

func metricServerInterceptor() hzapp.HandlerFunc {
	return func(c context.Context, ctx *hzapp.RequestContext) {
		beg := time.Now()
		ctx.Next(c)
		path := string(ctx.Path())
		method := string(ctx.Method())
		aid := string(ctx.GetHeader("AID"))
		metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeHTTP, method+"."+path, aid)
		metric.ServerHandleCounter.Inc(metric.TypeHTTP, method+"."+path, aid, http.StatusText(ctx.Response.StatusCode()))
	}
}
