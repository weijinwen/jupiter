// Copyright 2023 Douyu
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

package elasticsearch

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(name string) *Config {
	config := DefaultConfig()
	config.Name = name
	config.EnableMetric = false
	config.EnableAccessLog = false
	config.SlowThreshold = 0
	config.RetryWaitTime = time.Millisecond
	return config
}

func TestRoundTripper_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	config := newTestConfig("test-success")
	transport := NewElasticsearchRoundTripper(config.Name, config)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/_search", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRoundTripper_RetryOnServerError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	config := newTestConfig("test-retry")
	config.Retry = 3
	transport := NewElasticsearchRoundTripper(config.Name, config)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/_search", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRoundTripper_RetryExhausted(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := newTestConfig("test-exhausted")
	config.Retry = 2
	transport := NewElasticsearchRoundTripper(config.Name, config)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/_search", nil)
	require.NoError(t, err)

	_, err = transport.RoundTrip(req)
	require.Error(t, err)

	// retry=2 表示首次尝试 + 2 次重试
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRoundTripper_ReplayBodyOnRetry(t *testing.T) {
	var (
		attempts int32
		bodies   []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))

		if atomic.AddInt32(&attempts, 1) < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	config := newTestConfig("test-replay")
	config.Retry = 2
	transport := NewElasticsearchRoundTripper(config.Name, config)

	payload := `{"query":{"match_all":{}}}`
	req, err := http.NewRequest(http.MethodPost, server.URL+"/_search", strings.NewReader(payload))
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, bodies, 2)
	// 两次尝试都必须携带完整的请求体
	assert.Equal(t, payload, bodies[0])
	assert.Equal(t, payload, bodies[1])
}

func TestGetStatement(t *testing.T) {
	tests := []struct {
		err      string
		expected string
	}{
		{"elasticsearch error: boom", "elasticsearch error"},
		{"request failed: timeout", "request failed"},
		{"server error: 500", "server error"},
		{"something else", "Unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, getStatement(tt.err))
	}
}
