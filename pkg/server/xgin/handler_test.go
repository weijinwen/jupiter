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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 形态一：H 嵌套，handler 用 return 直接返回具体响应类型
func TestH_ReturnConcreteType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	OfficialCheck := func(c *gin.Context) String {
		return NewString(200, "测试")
	}

	router := gin.New()
	router.GET("/official", H(OfficialCheck))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/official", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "测试", rec.Body.String())
}

// 返回接口类型同样支持
func TestH_ReturnInterfaceType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) Renderable {
		if c.Query("fail") == "1" {
			return NewString(http.StatusBadRequest, "bad")
		}
		return NewJSON(200, gin.H{"code": 0})
	}

	router := gin.New()
	router.GET("/mixed", H(handler))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/mixed", nil))
	assert.JSONEq(t, `{"code":0}`, rec.Body.String())

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/mixed?fail=1", nil))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "bad", rec.Body.String())
}

// 自定义响应类型：实现 Render 即可，无需改 H
type menuResp struct {
	Items []string
}

func (m menuResp) Render(c *gin.Context) {
	c.String(200, "%s", joinItems(m.Items))
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
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/menu/my", H(func(c *gin.Context) menuResp {
		return menuResp{Items: []string{"home", "user"}}
	}))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/menu/my", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "home,user", rec.Body.String())
}

// 原生写法不经过 H，行为不变
func TestNativeRegistration_Unchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/official", func(c *gin.Context) {
		c.String(200, "测试")
		return
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/official", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "测试", rec.Body.String())
}

// 各响应类型经 H 写回
func TestH_AllKinds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/str", H(func(c *gin.Context) String {
		return NewString(200, "hi %s", "a")
	}))
	router.GET("/json", H(func(c *gin.Context) JSON {
		return NewJSON(200, gin.H{"code": 0})
	}))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/str", nil))
	assert.Equal(t, "hi a", rec.Body.String())

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/json", nil))
	assert.JSONEq(t, `{"code":0}`, rec.Body.String())
}
