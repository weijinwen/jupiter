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
	hzapp "github.com/cloudwego/hertz/pkg/app"
	hzws "github.com/hertz-contrib/websocket"
)

// WebSocketConn websocket conn，API 与 gorilla/websocket 的 *Conn 兼容
type WebSocketConn = *hzws.Conn

// WebSocketFunc ..
type WebSocketFunc func(WebSocketConn, error)

// WebSocket ..
type WebSocket struct {
	Pattern string
	Handler WebSocketFunc
	*hzws.HertzUpgrader
}

// Upgrade 将 HTTP 连接升级为 WebSocket；握手失败时以 nil conn 回调 Handler
func (ws *WebSocket) Upgrade(rc *hzapp.RequestContext) {
	err := ws.HertzUpgrader.Upgrade(rc, func(conn WebSocketConn) {
		defer conn.Close()
		ws.Handler(conn, nil)
	})
	if err != nil {
		ws.Handler(nil, err)
	}
}

// WebSocketOption ..
type WebSocketOption func(*WebSocket)

// WebSocketOptions ..
func WebSocketOptions(pattern string, handler WebSocketFunc, opts ...WebSocketOption) *WebSocket {
	ws := &WebSocket{
		Pattern:         pattern,
		Handler:         handler,
		HertzUpgrader:   &hzws.HertzUpgrader{},
	}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}
