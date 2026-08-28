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
	"io"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// hertzLogger 将 jupiter 的 *xlog.Logger 适配为 hlog.FullLogger，
// 使 Hertz 引擎内部日志（路由注册、启停、异常等）经由 jupiter
// 日志管道输出，与业务日志统一切割与采集。
type hertzLogger struct {
	logger *xlog.Logger
}

// newHertzLogger 构造适配器并接管 hlog 默认与系统 logger。
// level 按 Mode 语义映射：debug -> Debug，其余 -> Info。
func newHertzLogger(logger *xlog.Logger, mode string) *hertzLogger {
	hl := &hertzLogger{logger: logger}
	if mode == "debug" {
		hlog.SetLevel(hlog.LevelDebug)
	} else {
		hlog.SetLevel(hlog.LevelInfo)
	}
	hlog.SetLogger(hl)
	return hl
}

// hlog level -> zap level；zap 无 Trace/Notice 级别，分别降为 Debug/Info
func toZapLevel(lv hlog.Level) zapcore.Level {
	switch lv {
	case hlog.LevelTrace, hlog.LevelDebug:
		return zap.DebugLevel
	case hlog.LevelInfo, hlog.LevelNotice:
		return zap.InfoLevel
	case hlog.LevelWarn:
		return zap.WarnLevel
	case hlog.LevelError:
		return zap.ErrorLevel
	case hlog.LevelFatal:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func (h *hertzLogger) log(lv hlog.Level, msg string) {
	if ce := h.logger.Check(toZapLevel(lv), msg); ce != nil {
		ce.Write()
	}
}

func (h *hertzLogger) logf(lv hlog.Level, format string, v ...interface{}) {
	h.log(lv, fmt.Sprintf(format, v...))
}

// Logger 非格式化接口
func (h *hertzLogger) Trace(v ...interface{})  { h.log(hlog.LevelTrace, fmt.Sprint(v...)) }
func (h *hertzLogger) Debug(v ...interface{})  { h.log(hlog.LevelDebug, fmt.Sprint(v...)) }
func (h *hertzLogger) Info(v ...interface{})   { h.log(hlog.LevelInfo, fmt.Sprint(v...)) }
func (h *hertzLogger) Notice(v ...interface{}) { h.log(hlog.LevelNotice, fmt.Sprint(v...)) }
func (h *hertzLogger) Warn(v ...interface{})   { h.log(hlog.LevelWarn, fmt.Sprint(v...)) }
func (h *hertzLogger) Error(v ...interface{})  { h.log(hlog.LevelError, fmt.Sprint(v...)) }
func (h *hertzLogger) Fatal(v ...interface{})  { h.log(hlog.LevelFatal, fmt.Sprint(v...)) }

// FormatLogger 格式化接口
func (h *hertzLogger) Tracef(format string, v ...interface{}) { h.logf(hlog.LevelTrace, format, v...) }
func (h *hertzLogger) Debugf(format string, v ...interface{}) { h.logf(hlog.LevelDebug, format, v...) }
func (h *hertzLogger) Infof(format string, v ...interface{})  { h.logf(hlog.LevelInfo, format, v...) }
func (h *hertzLogger) Noticef(format string, v ...interface{}) {
	h.logf(hlog.LevelNotice, format, v...)
}
func (h *hertzLogger) Warnf(format string, v ...interface{}) { h.logf(hlog.LevelWarn, format, v...) }
func (h *hertzLogger) Errorf(format string, v ...interface{}) {
	h.logf(hlog.LevelError, format, v...)
}
func (h *hertzLogger) Fatalf(format string, v ...interface{}) { h.logf(hlog.LevelFatal, format, v...) }

// CtxLogger 带上下文的格式化接口；jupiter 日志管道暂不消费 ctx，语义等同 Infof
func (h *hertzLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelTrace, format, v...)
}
func (h *hertzLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelDebug, format, v...)
}
func (h *hertzLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelInfo, format, v...)
}
func (h *hertzLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelNotice, format, v...)
}
func (h *hertzLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelWarn, format, v...)
}
func (h *hertzLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelError, format, v...)
}
func (h *hertzLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	h.logf(hlog.LevelFatal, format, v...)
}

// Control 日志级别与输出控制。
// 级别过滤由 jupiter 日志管道按自身 level 生效（hlog 包级 level 已在
// newHertzLogger 按 Mode 设置，用于框架自身的级别开关）；输出侧由
// jupiter 日志管道（rotate/async）统一管理，SetOutput 不适用。
func (h *hertzLogger) SetLevel(lv hlog.Level) {}

func (h *hertzLogger) SetOutput(w io.Writer) {}

// 编译期断言：适配器完整实现 hlog.FullLogger
var _ hlog.FullLogger = (*hertzLogger)(nil)
