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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGinLogger_Access(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, logs := observer.New(zapcore.DebugLevel)
	logger := newGinLogger(zap.New(core), 0)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		beg := logger.begin()
		defer logger.access(c, beg)
		c.Next()
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusCreated, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	entries := logs.FilterMessage("access").All()
	assert.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "GET", ctx["method"])
	assert.Equal(t, int64(http.StatusCreated), ctx["code"])
	assert.Equal(t, "example.com", ctx["host"])
	assert.Equal(t, "/ping", ctx["path"])
}

func TestGinLogger_DefaultWriters(t *testing.T) {
	oldWriter := gin.DefaultWriter
	oldErrorWriter := gin.DefaultErrorWriter
	defer func() {
		gin.DefaultWriter = oldWriter
		gin.DefaultErrorWriter = oldErrorWriter
	}()

	core, logs := observer.New(zapcore.DebugLevel)
	logger := newGinLogger(zap.New(core), 0)
	logger.applyGinWriters()

	_, err := gin.DefaultWriter.Write([]byte("gin debug log\n"))
	assert.NoError(t, err)
	_, err = gin.DefaultErrorWriter.Write([]byte("gin error log\n"))
	assert.NoError(t, err)

	entries := logs.All()
	assert.Len(t, entries, 2)
	assert.Equal(t, zapcore.InfoLevel, entries[0].Level)
	assert.Equal(t, "gin debug log", entries[0].Message)
	assert.Equal(t, zapcore.ErrorLevel, entries[1].Level)
	assert.Equal(t, "gin error log", entries[1].Message)
}
