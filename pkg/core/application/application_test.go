// Copyright 2021 rex lv
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

package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/hooks"
	"github.com/weijinwen/jupiter/pkg/server"
)

// MockServer 是 server.Server 的模拟实现
type MockServer struct {
	mock.Mock
	name     string
	address  string
	scheme   string
	stopped  bool
	stopOnce sync.Once
}

func (m *MockServer) Serve() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockServer) Stop() error {
	m.stopOnce.Do(func() {
		m.stopped = true
	})
	args := m.Called()
	return args.Error(0)
}

func (m *MockServer) GracefulStop(ctx context.Context) error {
	m.stopOnce.Do(func() {
		m.stopped = true
	})
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServer) Info() *server.ServiceInfo {
	return &server.ServiceInfo{
		Name:    m.name,
		Address: m.address,
		Scheme:  m.scheme,
	}
}

func (m *MockServer) Healthz() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestNew 测试创建新的应用程序实例
func TestNew(t *testing.T) {
	// 测试成功创建应用
	app, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.cycle)
	assert.NotNil(t, app.logger)
	assert.NotNil(t, app.stopped)
	assert.NotNil(t, app.servers)
	assert.NotNil(t, app.disableMap)
}

// TestDefaultApp 测试创建默认应用程序实例
func TestDefaultApp(t *testing.T) {
	app := DefaultApp()
	assert.NotNil(t, app)
	assert.NotNil(t, app.cycle)
	assert.NotNil(t, app.logger)
	assert.NotNil(t, app.stopped)
	assert.NotNil(t, app.servers)
	assert.NotNil(t, app.disableMap)
}

// TestApplication_RegisterHooks 测试注册钩子函数
func TestApplication_RegisterHooks(t *testing.T) {
	app := DefaultApp()

	// 注册钩子函数
	executed := false
	app.RegisterHooks(hooks.Stage_BeforeRun, func() {
		executed = true
	})

	// 执行钩子
	app.runHooks(hooks.Stage_BeforeRun)

	assert.True(t, executed)
}

// TestApplication_Serve 测试注册服务器
func TestApplication_Serve(t *testing.T) {
	app := DefaultApp()

	// 创建模拟服务器
	server1 := &MockServer{name: "server1", address: "localhost:8080", scheme: "http"}
	server1.On("Serve").Return(nil)

	// 注册服务器
	err := app.Serve(server1)
	assert.NoError(t, err)
	assert.Len(t, app.servers, 1)
}

// TestApplication_Startup 测试启动应用程序
func TestApplication_Startup(t *testing.T) {
	app := &Application{}

	// 测试成功启动
	err := app.Startup()
	assert.NoError(t, err)
	assert.NotNil(t, app.cycle)
	assert.NotNil(t, app.logger)

	// 测试启动失败
	err = app.Startup(func() error {
		return assert.AnError
	})
	assert.Error(t, err)
}

// TestApplication_isDisable 测试禁用功能
func TestApplication_isDisable(t *testing.T) {
	app := DefaultApp()

	// 默认不禁用任何功能
	assert.False(t, app.isDisable(DisableParserFlag))
	assert.False(t, app.isDisable(DisableLoadConfig))

	// 禁用特定功能
	app.disableMap[DisableParserFlag] = true
	assert.True(t, app.isDisable(DisableParserFlag))
	assert.False(t, app.isDisable(DisableLoadConfig))
}

// TestApplication_parseFlags 测试解析命令行参数
func TestApplication_parseFlags(t *testing.T) {
	app := DefaultApp()

	// 测试正常解析
	err := app.parseFlags()
	assert.NoError(t, err)

	// 测试禁用解析
	app.disableMap[DisableParserFlag] = true
	err = app.parseFlags()
	assert.NoError(t, err)
}

// TestApplication_printBanner 测试打印横幅
func TestApplication_printBanner(t *testing.T) {
	app := DefaultApp()

	// 测试正常打印
	err := app.printBanner()
	assert.NoError(t, err)

	// 测试隐藏横幅
	app.HideBanner = true
	err = app.printBanner()
	assert.NoError(t, err)
}

// TestApplication_Stop 测试停止应用程序
func TestApplication_Stop(t *testing.T) {
	app := DefaultApp()

	// 创建模拟服务器
	server1 := &MockServer{name: "server1", address: "localhost:8080", scheme: "http"}
	server1.On("Stop").Return(nil).Once()

	// 添加服务器
	app.servers = append(app.servers, server1)

	// 启动一个goroutine来处理cycle.Done()
	go func() {
		// 模拟cycle.Done()的行为
		<-app.cycle.Done()
	}()

	// 停止应用
	err := app.Stop()
	assert.NoError(t, err)
	assert.True(t, server1.stopped)
}

// TestApplication_GracefulStop 测试优雅停止应用程序
func TestApplication_GracefulStop(t *testing.T) {
	app := DefaultApp()

	// 创建模拟服务器
	server1 := &MockServer{name: "server1", address: "localhost:8080", scheme: "http"}
	server1.On("GracefulStop", mock.Anything).Return(nil).Once()

	// 添加服务器
	app.servers = append(app.servers, server1)

	// 优雅停止应用
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.GracefulStop(ctx)
	assert.NoError(t, err)
	assert.True(t, server1.stopped)
}

// TestApplication_startServers 测试启动服务器
func TestApplication_startServers(t *testing.T) {
	app := DefaultApp()

	// 创建模拟服务器
	server1 := &MockServer{name: "server1", address: "localhost:8080", scheme: "http"}
	server1.On("Serve").Return(nil)

	// 添加服务器
	app.servers = append(app.servers, server1)

	// 启动服务器
	err := app.startServers()
	assert.NoError(t, err)
}

// TestApplication_WithOptions 测试应用选项
func TestApplication_WithOptions(t *testing.T) {
	app := DefaultApp()

	// 使用自定义配置解析器选项
	customParser := func(data []byte, v any) error {
		return nil
	}

	app.WithOptions(WithConfigParser(customParser))
	assert.Equal(t, conf.Unmarshaller(customParser), app.configParser)

	// 使用禁用选项
	app.WithOptions(WithDisable(DisableParserFlag))
	assert.True(t, app.isDisable(DisableParserFlag))
}

// TestApplication_clean 测试清理资源
func TestApplication_clean(t *testing.T) {
	app := DefaultApp()

	// 清理资源不会返回错误
	app.clean()
}
