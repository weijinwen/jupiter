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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

func extractAID(ctx *gin.Context) string {
	return ctx.Request.Header.Get("AID")
}

func recoverMiddleware(logger *xlog.Logger, slowQueryThresholdInMilli int64) gin.HandlerFunc {
	ginLogger := newGinLogger(logger, slowQueryThresholdInMilli)
	return func(c *gin.Context) {
		beg := ginLogger.begin()
		defer func() {
			if rec := recover(); rec != nil {
				ginLogger.panic(c, beg, rec)
				return
			}
			ginLogger.access(c, beg)
		}()
		c.Next()
	}
}

func metricServerInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		c.Next()
		metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeHTTP, c.Request.Method+"."+c.Request.URL.Path, extractAID(c))
		metric.ServerHandleCounter.Inc(metric.TypeHTTP, c.Request.Method+"."+c.Request.URL.Path, extractAID(c), http.StatusText(c.Writer.Status()))
	}
}
