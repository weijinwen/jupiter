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

// xgin 使用示例。
//
// 运行方式：
//
//	# 使用默认配置（host 取 --host flag，port 9091，mode release）
//	go run .
//
//	# 或通过配置文件（见同目录 config.toml，键名须用 camelCase 才能绑定）
//	go run . --config=config.toml
//
// 验证：
//
//	curl http://127.0.0.1:9091/health
//	curl http://127.0.0.1:9091/users/42
//	curl -X POST http://127.0.0.1:9091/users -H 'Content-Type: application/json' -d '{"name":"jupiter"}'
//	curl http://127.0.0.1:9091/hello?name=jupiter
//	curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" -H "Sec-WebSocket-Version: 13" http://127.0.0.1:9091/ws
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weijinwen/jupiter"
	"github.com/weijinwen/jupiter/pkg/server/governor"
	"github.com/weijinwen/jupiter/pkg/server/xgin"
)

func main() {
	app, err := jupiter.New()
	if err != nil {
		panic(err)
	}

	// 方式一（推荐）：配置驱动，读取 [server.gin]（StdConfig 的参数即键名后缀）。
	ginServer := xgin.StdConfig("gin").MustBuild()

	// 方式二：纯代码构建，不读配置文件。
	// ginServer := xgin.DefaultConfig().WithHost("0.0.0.0").WithPort(8080).MustBuild()

	// MustBuild 已内置 recover+访问日志（含慢查询）与 Prometheus 指标中间件，
	// 可按需继续追加自定义中间件。
	ginServer.Use(func(c *gin.Context) {
		c.Header("X-Jupiter", "xgin")
		c.Next()
	})

	// JSON 响应
	ginServer.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 路径参数
	ginServer.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	// POST + 请求体绑定（body 非法 JSON 时返回 400）
	ginServer.POST("/users", func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"name": req.Name})
	})

	// ── return 风格 handler（H 包装器） ──
	// handler 用 return 直接返回响应，无需手动调用 c.JSON/c.String。
	ginServer.GET("/hello", xgin.H(helloHandler))

	// ── WebSocket ──
	ws := xgin.WebSocketOptions("/ws", echoWebSocket)
	ginServer.Upgrade(ws)

	// ── 多服务器：同时挂载 governor 管理端口 ──
	govServer := governor.StdConfig("governor").Build()

	// 注册到应用（可同时注册多个 server）
	app.Serve(ginServer, govServer)

	// Run 阻塞运行；收到 SIGINT/SIGTERM 时优雅停止
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// helloHandler 是 return 风格 handler，返回 xgin.String。
func helloHandler(c *gin.Context) xgin.String {
	name := c.DefaultQuery("name", "world")
	return xgin.NewString(200, fmt.Sprintf("hello, %s!", name))
}

// echoWebSocket 回显 WebSocket 消息。
func echoWebSocket(conn xgin.WebSocketConn, err error) {
	if err != nil {
		return
	}
	for {
		msgType, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			return
		}
		conn.WriteMessage(msgType, msg) // echo
	}
}
