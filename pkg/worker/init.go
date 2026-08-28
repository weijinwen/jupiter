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

package worker

import (
	"context"
	"os"
	"time"

	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/hooks"
	"github.com/weijinwen/jupiter/pkg/flag"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

func init() {
	// 注册命令行参数
	flag.Register(&flag.StringFlag{
		Name:    "cmd",
		Usage:   "--cmd=cmd_name",
		Default: "",
	})

	flag.Register(&flag.StringFlag{
		Name:    "cmd-timeout",
		Usage:   "--cmd-timeout=30m",
		Default: "30m",
	})

	// 配置加载完成后，注册到 Stage_BeforeRun 阶段执行命令
	// 这样可以确保所有 worker 都已在初始化阶段注册完毕
	conf.OnLoaded(func(c *conf.Configuration) {
		// 先从 flag 获取命令参数，如果没有提供命令则跳过
		cmd := flag.String("cmd")
		if cmd == "" {
			xlog.Jupiter().Debug("no command provided, skipping execution")
			return
		}

		// 获取超时配置并解析
		timeoutStr := flag.String("cmd-timeout")
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			xlog.Jupiter().Error("invalid timeout format, using default 30m",
				zap.String("timeout", timeoutStr),
				zap.Error(err))
			timeout = 30 * time.Minute
		}

		// 注册到 Stage_BeforeRun 阶段执行命令
		// 这样可以确保所有 worker 在应用启动前都已注册
		hooks.Register(hooks.Stage_BeforeRun, func() {
			// 创建带超时的上下文
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// 执行命令
			if err := Execute(ctx, cmd); err != nil {
				xlog.Jupiter().Error("execute command failed",
					zap.String("cmd", cmd),
					zap.Error(err))
				os.Exit(1)
			}

			// 执行成功，退出程序
			xlog.Jupiter().Info("command executed, exiting",
				zap.String("cmd", cmd))
			os.Exit(0)
		})
	})
}
