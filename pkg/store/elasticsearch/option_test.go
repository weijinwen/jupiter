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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "default", config.Name)
	assert.Equal(t, 3000*time.Millisecond, config.Timeout)
	assert.Equal(t, 30*time.Second, config.KeepAlive)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 100, config.MaxIdleConnsPerHost)
	assert.Equal(t, 30*time.Second, config.IdleConnTimeout)
	assert.False(t, config.DisableKeepAlives)
	assert.Equal(t, 100, config.MaxConnsPerHost)
	assert.Equal(t, 30*time.Second, config.ResponseHeaderTimeout)
	assert.False(t, config.Debug)
	assert.True(t, config.EnableMetric)
	assert.False(t, config.EnableAccessLog)
	assert.Equal(t, time.Second, config.SlowThreshold)
	assert.Equal(t, "panic", config.OnDialError)
	assert.Equal(t, 2, config.Retry)
	assert.Equal(t, 200*time.Millisecond, config.RetryWaitTime)
}

func TestConfigAddrs(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name:     "addresses 优先",
			config:   &Config{Addresses: []string{"http://a:9200", "http://b:9200"}, Addr: "http://c:9200"},
			expected: []string{"http://a:9200", "http://b:9200"},
		},
		{
			name:     "回落 addr",
			config:   &Config{Addr: "http://c:9200"},
			expected: []string{"http://c:9200"},
		},
		{
			name:     "都未配置",
			config:   &Config{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.addrs())
		})
	}
}

func TestConfigBuildWithoutAddr(t *testing.T) {
	config := DefaultConfig()
	config.Name = "test-no-addr"

	client, err := config.Build()

	assert.Nil(t, client)
	assert.EqualError(t, err, "no addr found")
}

func TestConfigBuildWithAddr(t *testing.T) {
	config := DefaultConfig()
	config.Name = "test-build"
	config.Addr = "http://127.0.0.1:9200"

	client, err := config.Build()

	if assert.NoError(t, err) && assert.NotNil(t, client) {
		assert.Equal(t, config, client.Config())
	}
}

func TestConfigSingleton(t *testing.T) {
	config := DefaultConfig()
	config.Name = "test-singleton"
	config.Addr = "http://127.0.0.1:9200"

	first, err := config.Singleton()
	if !assert.NoError(t, err) || !assert.NotNil(t, first) {
		return
	}

	second, err := config.Singleton()
	if !assert.NoError(t, err) || !assert.NotNil(t, second) {
		return
	}

	// 同名配置应返回同一实例
	assert.Same(t, first, second)
}
