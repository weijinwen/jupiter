// Copyright 2022 Douyu
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

package resty

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_RetryCondition_Applied(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Addr = server.URL
	config.RetryCount = 2
	config.RetryWaitTime = time.Millisecond
	config.RetryMaxWaitTime = time.Millisecond
	config.RetryCondition = func(r *resty.Response, err error) bool {
		return true
	}

	client := config.MustBuild()
	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	assert.Equal(t, int32(3), atomic.LoadInt32(&calls), "1 initial attempt + 2 retries forced by RetryCondition")
}

func TestBuild_HeadersAndAuthToken(t *testing.T) {
	var gotAuth, gotCustom string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCustom = r.Header.Get("X-App-Name")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Addr = server.URL
	config.Headers = map[string]string{"X-App-Name": "jupiter"}
	config.AuthToken = "tok123"

	client := config.MustBuild()
	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	assert.Equal(t, "jupiter", gotCustom)
	assert.Equal(t, "Bearer tok123", gotAuth)
}

func TestBuild_BasicAuth(t *testing.T) {
	var user, pass string
	var ok bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Addr = server.URL
	config.Username = "alice"
	config.Password = "secret"

	client := config.MustBuild()
	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	assert.True(t, ok)
	assert.Equal(t, "alice", user)
	assert.Equal(t, "secret", pass)
}
