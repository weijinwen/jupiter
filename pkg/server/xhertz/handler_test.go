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
	"net/http"
	"testing"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"
)

func newTestContext() *hzapp.RequestContext {
	rc := hzapp.NewContext(0)
	rc.Request.SetMethod(http.MethodGet)
	rc.Request.SetRequestURI("/test")
	return rc
}

// 形态一：H 嵌套，handler 用 return 直接返回具体响应类型
func TestH_ReturnConcreteType(t *testing.T) {
	OfficialCheck := func(ctx context.Context, rc *hzapp.RequestContext) String {
		return NewString(200, "测试")
	}

	handler := H(OfficialCheck)
	rc := newTestContext()
	handler(context.Background(), rc)

	assert.Equal(t, 200, rc.Response.StatusCode())
	assert.Equal(t, "测试", string(rc.Response.Body()))
}

// 返回接口类型同样支持
func TestH_ReturnInterfaceType(t *testing.T) {
	handler := H(func(ctx context.Context, rc *hzapp.RequestContext) Renderable {
		if string(rc.QueryArgs().Peek("fail")) == "1" {
			return NewString(http.StatusBadRequest, "bad")
		}
		return NewJSON(200, map[string]any{"code": 0})
	})

	rc := newTestContext()
	handler(context.Background(), rc)
	assert.Equal(t, 200, rc.Response.StatusCode())
	assert.JSONEq(t, `{"code":0}`, string(rc.Response.Body()))

	rc = newTestContext()
	rc.QueryArgs().Set("fail", "1")
	handler(context.Background(), rc)
	assert.Equal(t, http.StatusBadRequest, rc.Response.StatusCode())
	assert.Equal(t, "bad", string(rc.Response.Body()))
}

// 自定义响应类型：实现 Render 即可，无需改 H
type menuResp struct {
	Items []string
}

func (m menuResp) Render(ctx context.Context, rc *hzapp.RequestContext) {
	rc.String(200, "%s", joinItems(m.Items))
}

func joinItems(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}

func TestH_CustomRenderable(t *testing.T) {
	handler := H(func(ctx context.Context, rc *hzapp.RequestContext) menuResp {
		return menuResp{Items: []string{"home", "user"}}
	})

	rc := newTestContext()
	handler(context.Background(), rc)

	assert.Equal(t, 200, rc.Response.StatusCode())
	assert.Equal(t, "home,user", string(rc.Response.Body()))
}

// 各响应类型经 H 写回
func TestH_AllKinds(t *testing.T) {
	str := H(func(ctx context.Context, rc *hzapp.RequestContext) String {
		return NewString(200, "hi %s", "a")
	})
	rc := newTestContext()
	str(context.Background(), rc)
	assert.Equal(t, "hi a", string(rc.Response.Body()))

	js := H(func(ctx context.Context, rc *hzapp.RequestContext) JSON {
		return NewJSON(200, map[string]any{"code": 0})
	})
	rc = newTestContext()
	js(context.Background(), rc)
	assert.JSONEq(t, `{"code":0}`, string(rc.Response.Body()))
}
