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

package xgin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

type ginLogger struct {
	logger                    *xlog.Logger
	slowQueryThresholdInMilli int64
}

type ginLogWriter struct {
	logger *xlog.Logger
	level  zapcore.Level
}

func (w ginLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}
	if ce := w.logger.Check(w.level, msg); ce != nil {
		ce.Write()
	}
	return len(p), nil
}

func newGinLogger(logger *xlog.Logger, slowQueryThresholdInMilli int64) *ginLogger {
	return &ginLogger{
		logger:                    logger,
		slowQueryThresholdInMilli: slowQueryThresholdInMilli,
	}
}

func (l *ginLogger) begin() time.Time {
	return time.Now()
}

func (l *ginLogger) applyGinWriters() {
	gin.DefaultWriter = ginLogWriter{logger: l.logger, level: zap.InfoLevel}
	gin.DefaultErrorWriter = ginLogWriter{logger: l.logger, level: zap.ErrorLevel}
}

func (l *ginLogger) access(c *gin.Context, beg time.Time) {
	l.logger.Info("access", l.accessFields(c, beg)...)
}

func (l *ginLogger) panic(c *gin.Context, beg time.Time, rec any) {
	var brokenPipe bool
	if ne, ok := rec.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			msg := strings.ToLower(se.Error())
			if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer") {
				brokenPipe = true
			}
		}
	}

	err := rec.(error)
	fields := l.baseFields(beg)
	fields = append(fields, zap.ByteString("stack", stack(3)))
	fields = append(fields, zap.String("err", err.Error()))
	l.logger.Error("access", fields...)

	if brokenPipe {
		c.Error(err) // nolint: errcheck
		c.Abort()
		return
	}
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (l *ginLogger) accessFields(c *gin.Context, beg time.Time) []xlog.Field {
	fields := l.baseFields(beg)
	fields = append(fields,
		zap.String("method", c.Request.Method),
		zap.Int("code", c.Writer.Status()),
		zap.Int("size", c.Writer.Size()),
		zap.String("host", c.Request.Host),
		zap.String("path", c.Request.URL.Path),
		zap.String("ip", c.ClientIP()),
		zap.String("err", c.Errors.ByType(gin.ErrorTypePrivate).String()),
	)
	return fields
}

func (l *ginLogger) baseFields(beg time.Time) []xlog.Field {
	fields := make([]xlog.Field, 0, 8)
	fields = append(fields, zap.Float64("cost", time.Since(beg).Seconds()))
	if l.slowQueryThresholdInMilli > 0 {
		if cost := int64(time.Since(beg)) / 1e6; cost > l.slowQueryThresholdInMilli {
			fields = append(fields, zap.Int64("slow", cost))
		}
	}
	return fields
}

func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
