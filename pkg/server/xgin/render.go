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
	"github.com/gin-gonic/gin"
)

// Renderable 任何能把自身写入 *gin.Context 的响应类型，
// 「return 风格」handler 的返回值需实现该接口，由 H 统一写回。
// 新增响应类型时让其实现 Render(*gin.Context) 即可，无需改 H。
type Renderable interface {
	Render(c *gin.Context)
}

// String 字符串响应
type String struct {
	Code int
	Text string
	Args []any // 可选，Text 作为格式化模板
}

// NewString 构造字符串响应
func NewString(code int, text string, args ...any) String {
	return String{Code: code, Text: text, Args: args}
}

// Render implements Renderable.
func (s String) Render(c *gin.Context) {
	if len(s.Args) == 0 {
		c.String(s.Code, "%s", s.Text)
		return
	}
	c.String(s.Code, s.Text, s.Args...)
}

// JSON JSON 响应
type JSON struct {
	Code   int
	Object any
}

// NewJSON 构造 JSON 响应
func NewJSON(code int, obj any) JSON {
	return JSON{Code: code, Object: obj}
}

// Render implements Renderable.
func (j JSON) Render(c *gin.Context) {
	c.JSON(j.Code, j.Object)
}
