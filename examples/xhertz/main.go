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

// xhertz 使用示例。
//
// 运行方式：
//
//	# 使用默认配置（host 取 --host flag，port 9092，mode release）
//	go run .
//
//	# 或通过配置文件（见同目录 config.toml，键名须用 camelCase 才能绑定）
//	go run . --config=config.toml
//
// 验证：
//
//	curl http://127.0.0.1:9092/health
//	curl http://127.0.0.1:9092/users/42
//	curl -X POST http://127.0.0.1:9092/users -H 'Content-Type: application/json' -d '{"name":"jupiter"}'
package main

import (
	"context"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/weijinwen/jupiter"
	"github.com/weijinwen/jupiter/pkg/server/xhertz"
)

func main() {
	app, err := jupiter.New()
	if err != nil {
		panic(err)
	}

	// 方式一（推荐）：配置驱动，读取 [server.hertz]（StdConfig 的参数即键名后缀）。
	hertzServer := xhertz.StdConfig("hertz").MustBuild()

	// 方式二：纯代码构建，不读配置文件。
	// hertzServer := xhertz.DefaultConfig().WithHost("0.0.0.0").WithPort(9092).MustBuild()

	// MustBuild 已内置 recover+访问日志（含慢查询）与 Prometheus 指标中间件，
	// 可按需继续追加自定义中间件。
	hertzServer.Use(func(c context.Context, rc *hzapp.RequestContext) {
		rc.Response.Header.Set("X-Jupiter", "xhertz")
		rc.Next(c)
	})

	// JSON 响应
	hertzServer.GET("/health", func(c context.Context, rc *hzapp.RequestContext) {
		rc.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})

	// 路径参数（Hertz 使用 ":param" 风格）
	hertzServer.GET("/users/:id", func(c context.Context, rc *hzapp.RequestContext) {
		rc.JSON(consts.StatusOK, map[string]string{"id": rc.Param("id")})
	})

	// POST + 请求体绑定（body 非法 JSON 时返回 400）
	hertzServer.POST("/users", func(c context.Context, rc *hzapp.RequestContext) {
		var req struct {
			Name string `json:"name"`
		}
		if err := rc.BindAndValidate(&req); err != nil {
			rc.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		rc.JSON(consts.StatusOK, map[string]string{"name": req.Name})
	})

	// 注册到应用（可同时注册多个 server，如 governor）
	app.Serve(hertzServer)

	// Run 阻塞运行；收到 SIGINT/SIGTERM 时优雅停止
	if err := app.Run(); err != nil {
		panic(err)
	}
}
