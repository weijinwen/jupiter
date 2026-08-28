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
	"net/http"
	"testing"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_DefaultAllowAll(t *testing.T) {
	called := false
	handler := corsMiddleware(DefaultCORSConfig())
	rc := hzapp.NewContext(1)
	rc.Request.SetMethod(http.MethodGet)
	rc.Request.SetHost("api.example.com")
	rc.Request.SetRequestURI("/ping")
	rc.Request.Header.Set("Origin", "https://example.com")
	rc.SetHandlers([]hzapp.HandlerFunc{
		handler,
		func(ctx context.Context, rc *hzapp.RequestContext) {
			called = true
			rc.String(http.StatusOK, "pong")
		},
	})
	rc.Next(context.Background())

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rc.Response.StatusCode())
	assert.Equal(t, "*", string(rc.Response.Header.Peek("Access-Control-Allow-Origin")))
	assert.Equal(t, "Content-Length", string(rc.Response.Header.Peek("Access-Control-Expose-Headers")))
}

func TestCORSMiddleware_DefaultPreflight(t *testing.T) {
	called := false
	handler := corsMiddleware(DefaultCORSConfig())
	rc := hzapp.NewContext(1)
	rc.Request.SetMethod(http.MethodOptions)
	rc.Request.SetHost("api.example.com")
	rc.Request.SetRequestURI("/ping")
	rc.Request.Header.Set("Origin", "https://example.com")
	rc.Request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rc.SetHandlers([]hzapp.HandlerFunc{
		handler,
		func(ctx context.Context, rc *hzapp.RequestContext) {
			called = true
			rc.Status(http.StatusOK)
		},
	})
	rc.Next(context.Background())

	assert.False(t, called)
	assert.Equal(t, http.StatusNoContent, rc.Response.StatusCode())
	assert.Equal(t, "*", string(rc.Response.Header.Peek("Access-Control-Allow-Origin")))
	assert.Contains(t, string(rc.Response.Header.Peek("Access-Control-Allow-Methods")), http.MethodPost)
}
