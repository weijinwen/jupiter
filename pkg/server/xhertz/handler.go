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

	hzapp "github.com/cloudwego/hertz/pkg/app"
)

// H 将「return 风格」的 handler 适配为 hzapp.HandlerFunc。
//
// 用法：
//
//	h.GET("/official", H(OfficialCheck))    // func(ctx, rc) xhertz.String
//	h.GET("/menu/my", H(adminHandler.MenuMy))
//
// handler 返回类型需实现 Renderable（String/JSON 均已实现）；
// 自定义响应类型实现 Render(ctx, *hzapp.RequestContext) 即可，无需改 H。
// 原生写法不受影响，直接注册即可。
func H[T Renderable](fn func(context.Context, *hzapp.RequestContext) T) hzapp.HandlerFunc {
	return func(ctx context.Context, rc *hzapp.RequestContext) { fn(ctx, rc).Render(ctx, rc) }
}
