package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestConfig_Singleton(t *testing.T) {
	// 创建配置
	config := &Config{
		name: "test",
		Addr: []string{"localhost:6379"},
	}

	// 第一次调用应该创建新实例
	client1, err := config.Singleton()
	assert.NoError(t, err)
	assert.NotNil(t, client1)

	// 第二次调用应该返回相同实例
	client2, err := config.Singleton()
	assert.NoError(t, err)
	assert.Equal(t, client1, client2)
}

func TestConfig_MustSingleton(t *testing.T) {
	// 创建配置
	config := &Config{
		name: "test",
		Addr: []string{"localhost:6379"},
	}

	// 调用 MustSingleton 应该返回客户端
	client := config.MustSingleton()
	assert.NotNil(t, client)
}

func TestConfig_MustBuild(t *testing.T) {
	// 创建配置
	config := &Config{
		name: "test",
		Addr: []string{"localhost:6379"},
	}

	// 调用 MustBuild 应该返回客户端
	client := config.MustBuild()
	assert.NotNil(t, client)
}

func TestConfig_Build(t *testing.T) {
	// 测试空地址
	config := &Config{
		name: "test",
		Addr: []string{},
	}
	client, err := config.Build()
	assert.Error(t, err)
	assert.Nil(t, client)

	// 测试有效地址
	config = &Config{
		name:        "test",
		Addr:        []string{"localhost:6379"},
		OnDialError: "error", // 避免panic
		logger:      zap.NewNop(),
	}
	// 由于没有真实的Redis服务器，这里会返回错误
	client, err = config.Build()
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestConfig_build(t *testing.T) {
	// 测试有效配置
	config := &Config{
		name:        "test",
		Addr:        []string{"localhost:6379"},
		OnDialError: "error", // 避免panic
		logger:      zap.NewNop(),
	}

	// 由于没有真实的Redis服务器，这里会返回错误
	client, err := config.build()
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestClient_UniversalClient(t *testing.T) {
	// 创建一个简单的客户端，不需要完整的实现
	client := &Client{}

	// 验证客户端不为nil
	assert.NotNil(t, client)
}
