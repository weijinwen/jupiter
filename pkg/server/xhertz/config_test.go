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

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	assert.Equal(t, "release", c.Mode)
	assert.Equal(t, 9092, c.Port)
	assert.Equal(t, int64(500), c.SlowQueryThresholdInMilli)
	assert.False(t, c.DisableMetric)
	assert.NotNil(t, c.logger)
}

func TestConfig_Address(t *testing.T) {
	c := DefaultConfig()
	c.Host = "0.0.0.0"
	c.Port = 8080
	assert.Equal(t, "0.0.0.0:8080", c.Address())
}

func TestDefaultConfig_TimeoutAndTLS(t *testing.T) {
	c := DefaultConfig()
	assert.Zero(t, c.ReadTimeout)
	assert.Zero(t, c.WriteTimeout)
	assert.Zero(t, c.IdleTimeout)
	assert.Zero(t, c.ShutdownTimeout)
	assert.False(t, c.enableTLS())
}

func TestConfig_EnableTLS(t *testing.T) {
	c := DefaultConfig()
	c.CertFile = "cert.pem"
	assert.True(t, c.enableTLS())

	c = DefaultConfig()
	c.KeyFile = "key.pem"
	assert.True(t, c.enableTLS())
}

func TestConfig_Builders(t *testing.T) {
	c := DefaultConfig().WithHost("127.0.0.1").WithPort(9090)
	assert.Equal(t, "127.0.0.1", c.Host)
	assert.Equal(t, 9090, c.Port)
	assert.Equal(t, "127.0.0.1:9090", c.Address())
}
