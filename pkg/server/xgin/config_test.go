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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestServer_ServeTimeouts(t *testing.T) {
	c := DefaultConfig()
	c.Port = 0
	c.ReadTimeout = time.Second
	c.WriteTimeout = 2 * time.Second
	c.IdleTimeout = 30 * time.Second
	s := c.MustBuild()

	assert.Equal(t, time.Second, s.config.ReadTimeout)
	assert.Equal(t, 2*time.Second, s.config.WriteTimeout)
	assert.Equal(t, 30*time.Second, s.config.IdleTimeout)
}
