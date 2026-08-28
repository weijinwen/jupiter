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
	"testing"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger() (*xlogAlias, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return (*xlogAlias)(zap.New(core)), logs
}

// xlog.Logger 即 zap.Logger 的别名，测试用别名避免直接依赖 xlog 包导入
type xlogAlias = zap.Logger

func TestHertzLogger_LevelsAndFormat(t *testing.T) {
	lg, logs := newObservedLogger()
	hl := &hertzLogger{logger: lg}

	hl.Trace("a", "b")
	hl.Debug("debug msg")
	hl.Info("info msg")
	hl.Notice("notice msg")
	hl.Warn("warn msg")
	hl.Error("error msg")
	hl.Infof("hello %s", "jupiter")
	hl.Warnf("count=%d", 42)
	hl.CtxErrorf(nil, "ctx %s", "err")

	entries := logs.All()
	assert.Len(t, entries, 9)

	// zap 无 Trace/Notice，分别降为 Debug/Info
	// fmt.Sprint 仅在相邻参数非全为字符串时加空格，("a","b") -> "ab"
	assert.Equal(t, zapcore.DebugLevel, entries[0].Level)
	assert.Equal(t, "ab", entries[0].Message)
	assert.Equal(t, zapcore.DebugLevel, entries[1].Level)
	assert.Equal(t, zapcore.InfoLevel, entries[2].Level)
	assert.Equal(t, zapcore.InfoLevel, entries[3].Level)
	assert.Equal(t, zapcore.WarnLevel, entries[4].Level)
	assert.Equal(t, zapcore.ErrorLevel, entries[5].Level)

	// 格式化接口真正展开
	assert.Equal(t, "hello jupiter", entries[6].Message)
	assert.Equal(t, "count=42", entries[7].Message)
	assert.Equal(t, "ctx err", entries[8].Message)
}

func TestHertzLogger_LevelFilter(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel) // jupiter 默认 info
	hl := &hertzLogger{logger: zap.New(core)}

	hl.Debug("should be filtered")
	hl.Info("should pass")

	entries := logs.All()
	assert.Len(t, entries, 1)
	assert.Equal(t, "should pass", entries[0].Message)
}

func TestToZapLevel(t *testing.T) {
	assert.Equal(t, zapcore.DebugLevel, toZapLevel(hlog.LevelTrace))
	assert.Equal(t, zapcore.DebugLevel, toZapLevel(hlog.LevelDebug))
	assert.Equal(t, zapcore.InfoLevel, toZapLevel(hlog.LevelInfo))
	assert.Equal(t, zapcore.InfoLevel, toZapLevel(hlog.LevelNotice))
	assert.Equal(t, zapcore.WarnLevel, toZapLevel(hlog.LevelWarn))
	assert.Equal(t, zapcore.ErrorLevel, toZapLevel(hlog.LevelError))
	assert.Equal(t, zapcore.FatalLevel, toZapLevel(hlog.LevelFatal))
}

func TestHertzLogger_ImplementsFullLogger(t *testing.T) {
	var _ hlog.FullLogger = (*hertzLogger)(nil)
}
