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
	"net/url"
	"testing"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func Test_Server_WebSocket(t *testing.T) {
	c := DefaultConfig()
	c.Host = "127.0.0.1"
	c.Port = 19392
	s := c.MustBuild()

	upgradeErr := make(chan error, 1)
	s.Upgrade(WebSocketOptions("/ws", func(conn WebSocketConn, err error) {
		if err != nil {
			upgradeErr <- err
			return
		}
		msgType, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			upgradeErr <- readErr
			return
		}
		upgradeErr <- conn.WriteMessage(msgType, msg)
	}))

	assert.Equal(t, "http", s.Info().Scheme)

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve() }()

	// 端口由使用方固定指定，拨号重试等待 Serve goroutine 完成监听
	wsURL := url.URL{Scheme: "ws", Host: c.Address(), Path: "/ws"}

	var conn *gorillaws.Conn
	var dialErr error
	dialer := gorillaws.Dialer{HandshakeTimeout: 3 * time.Second}
	for i := 0; i < 20; i++ {
		conn, _, dialErr = dialer.Dial(wsURL.String(), nil)
		if dialErr == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.NoError(t, dialErr)
	if conn == nil {
		t.Fatal("dial websocket failed")
	}
	defer conn.Close()

	assert.NoError(t, conn.WriteMessage(gorillaws.TextMessage, []byte("hello jupiter")))

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	msgType, msg, readErr := conn.ReadMessage()
	assert.NoError(t, readErr)
	assert.Equal(t, gorillaws.TextMessage, msgType)
	assert.Equal(t, "hello jupiter", string(msg))
	assert.NoError(t, <-upgradeErr)

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

func Test_WebSocketOptions_Defaults(t *testing.T) {
	ws := WebSocketOptions("/ws", func(conn WebSocketConn, err error) {})
	assert.Equal(t, "/ws", ws.Pattern)
	assert.NotNil(t, ws.Handler)
	assert.NotNil(t, ws.HertzUpgrader)
	assert.False(t, ws.HertzUpgrader.EnableCompression)
}
