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
	"testing"
	"time"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

func Test_Server(t *testing.T) {
	c := DefaultConfig()
	c.Port = 0
	s := c.MustBuild()
	s.GET("/health", func(ctx context.Context, rc *hzapp.RequestContext) {
		rc.SetStatusCode(consts.StatusOK)
	})

	assert.True(t, s.Healthz())
	assert.NotNil(t, s.Info())

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve() }()

	time.Sleep(time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	assert.NoError(t, s.GracefulStop(ctx))

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Serve() did not return after GracefulStop")
	}
}

func Test_Server_OptionsApplied(t *testing.T) {
	c := DefaultConfig()
	c.Port = 0
	c.ReadTimeout = time.Second
	c.WriteTimeout = 2 * time.Second
	c.IdleTimeout = 30 * time.Second
	c.ShutdownTimeout = 10 * time.Second

	s := c.MustBuild()
	opts := s.Hertz.GetOptions()
	assert.Equal(t, time.Second, opts.ReadTimeout)
	assert.Equal(t, 2*time.Second, opts.WriteTimeout)
	assert.Equal(t, 30*time.Second, opts.IdleTimeout)
	assert.Equal(t, 10*time.Second, opts.ExitWaitTimeout)
	assert.Equal(t, "http", s.Info().Scheme)
}
