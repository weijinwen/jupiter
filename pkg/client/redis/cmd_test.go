package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMustGet(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGet(context.Background(), "test")
	})
}

func TestMustGetInt(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGetInt(context.Background(), "test")
	})
}

func TestMustGetInt64(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGetInt64(context.Background(), "test")
	})
}

func TestMustGetUint64(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGetUint64(context.Background(), "test")
	})
}

func TestMustGetFloat32(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGetFloat32(context.Background(), "test")
	})
}

func TestMustGetFloat64(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustGetFloat64(context.Background(), "test")
	})
}

func TestGetBool(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_, _ = rc.GetBool(context.Background(), "test")
	})
}

func TestHGetAllStruct(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.HGetAllStruct(context.Background(), "test", nil)
	})
}

// 测试新增的方法
func TestMustSet(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSet(context.Background(), "test", "value", time.Second)
	})
}

func TestMustSetNX(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSetNX(context.Background(), "test", "value", time.Second)
	})
}

func TestMustDel(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustDel(context.Background(), "test")
	})
}

func TestMustExpire(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustExpire(context.Background(), "test", time.Second)
	})
}

func TestMustTTL(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustTTL(context.Background(), "test")
	})
}

func TestMustExists(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustExists(context.Background(), "test")
	})
}

func TestMustIncr(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustIncr(context.Background(), "test")
	})
}

func TestMustIncrBy(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustIncrBy(context.Background(), "test", 1)
	})
}

func TestMustDecr(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustDecr(context.Background(), "test")
	})
}

func TestMustDecrBy(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustDecrBy(context.Background(), "test", 1)
	})
}

func TestMustHSet(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHSet(context.Background(), "test", "field", "value")
	})
}

func TestMustHGet(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHGet(context.Background(), "test", "field")
	})
}

func TestMustHGetInt(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHGetInt(context.Background(), "test", "field")
	})
}

func TestMustHGetInt64(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHGetInt64(context.Background(), "test", "field")
	})
}

func TestMustHDel(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHDel(context.Background(), "test", "field")
	})
}

func TestMustHExists(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHExists(context.Background(), "test", "field")
	})
}

func TestMustHGetAll(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustHGetAll(context.Background(), "test")
	})
}

func TestMustLPush(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustLPush(context.Background(), "test", "value")
	})
}

func TestMustRPush(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustRPush(context.Background(), "test", "value")
	})
}

func TestMustLPop(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustLPop(context.Background(), "test")
	})
}

func TestMustRPop(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustRPop(context.Background(), "test")
	})
}

func TestMustLLen(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustLLen(context.Background(), "test")
	})
}

func TestMustLIndex(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustLIndex(context.Background(), "test", 0)
	})
}

func TestMustSAdd(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSAdd(context.Background(), "test", "member")
	})
}

func TestMustSRem(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSRem(context.Background(), "test", "member")
	})
}

func TestMustSCard(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSCard(context.Background(), "test")
	})
}

func TestMustSIsMember(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSIsMember(context.Background(), "test", "member")
	})
}

func TestMustSMembers(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSMembers(context.Background(), "test")
	})
}

func TestMustZAdd(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZAdd(context.Background(), "test", redis.Z{Score: 1, Member: "member"})
	})
}

func TestMustZRem(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZRem(context.Background(), "test", "member")
	})
}

func TestMustZCard(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZCard(context.Background(), "test")
	})
}

func TestMustZScore(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZScore(context.Background(), "test", "member")
	})
}

func TestMustZRange(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZRange(context.Background(), "test", 0, -1)
	})
}

func TestMustZRangeWithScores(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustZRangeWithScores(context.Background(), "test", 0, -1)
	})
}

func TestSetJSON(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.SetJSON(context.Background(), "test", map[string]string{"key": "value"}, time.Second)
	})
}

func TestMustSetJSON(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		_ = rc.MustSetJSON(context.Background(), "test", map[string]string{"key": "value"}, time.Second)
	})
}

func TestGetJSON(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		var result map[string]string
		_ = rc.GetJSON(context.Background(), "test", &result)
	})
}

func TestMustGetJSON(t *testing.T) {
	var rc *Client
	assert.Panics(t, func() {
		var result map[string]string
		_ = rc.MustGetJSON(context.Background(), "test", &result)
	})
}
