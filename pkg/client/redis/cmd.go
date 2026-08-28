package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func (rc *Client) MustGet(ctx context.Context, key string) string {
	return rc.Get(ctx, key).Val()
}

func (rc *Client) MustGetInt(ctx context.Context, key string) int {
	val, err := rc.Get(ctx, key).Int()
	if err != nil {
		return 0
	}
	return val
}

func (rc *Client) MustGetInt64(ctx context.Context, key string) int64 {
	val, err := rc.Get(ctx, key).Int64()
	if err != nil {
		return 0
	}
	return val
}

func (rc *Client) MustGetUint64(ctx context.Context, key string) uint64 {
	val, err := rc.Get(ctx, key).Uint64()
	if err != nil {
		return 0
	}
	return val
}

func (rc *Client) MustGetFloat32(ctx context.Context, key string) float32 {
	val, err := rc.Get(ctx, key).Float32()
	if err != nil {
		return 0
	}
	return val
}

func (rc *Client) MustGetFloat64(ctx context.Context, key string) float64 {
	val, err := rc.Get(ctx, key).Float64()
	if err != nil {
		return 0
	}
	return val
}

func (rc *Client) GetBool(ctx context.Context, key string) (bool, error) {
	return rc.Get(ctx, key).Bool()
}

func (rc *Client) HGetAllStruct(ctx context.Context, key string, to any) error {
	return rc.HGetAll(ctx, key).Scan(&to)
}

// MustSet 设置键值对，不返回错误
func (rc *Client) MustSet(ctx context.Context, key string, value any, expiration time.Duration) bool {
	err := rc.Set(ctx, key, value, expiration).Err()
	return err == nil
}

// MustSetNX 设置键值对（仅当键不存在时），不返回错误
func (rc *Client) MustSetNX(ctx context.Context, key string, value any, expiration time.Duration) bool {
	result, err := rc.SetNX(ctx, key, value, expiration).Result()
	return err == nil && result
}

// MustDel 删除键，不返回错误
func (rc *Client) MustDel(ctx context.Context, keys ...string) bool {
	err := rc.Del(ctx, keys...).Err()
	return err == nil
}

// MustExpire 设置键的过期时间，不返回错误
func (rc *Client) MustExpire(ctx context.Context, key string, expiration time.Duration) bool {
	err := rc.Expire(ctx, key, expiration).Err()
	return err == nil
}

// MustTTL 获取键的剩余生存时间，出错返回-1
func (rc *Client) MustTTL(ctx context.Context, key string) time.Duration {
	ttl, err := rc.TTL(ctx, key).Result()
	if err != nil {
		return -1
	}
	return ttl
}

// MustExists 检查键是否存在，出错返回false
func (rc *Client) MustExists(ctx context.Context, key string) bool {
	result, err := rc.Exists(ctx, key).Result()
	return err == nil && result > 0
}

// MustIncr 递增键的值，出错返回0
func (rc *Client) MustIncr(ctx context.Context, key string) int64 {
	result, err := rc.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustIncrBy 按指定值递增键的值，出错返回0
func (rc *Client) MustIncrBy(ctx context.Context, key string, value int64) int64 {
	result, err := rc.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustDecr 递减键的值，出错返回0
func (rc *Client) MustDecr(ctx context.Context, key string) int64 {
	result, err := rc.Decr(ctx, key).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustDecrBy 按指定值递减键的值，出错返回0
func (rc *Client) MustDecrBy(ctx context.Context, key string, value int64) int64 {
	result, err := rc.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustHSet 设置哈希表中的字段值，不返回错误
func (rc *Client) MustHSet(ctx context.Context, key string, values ...any) bool {
	err := rc.HSet(ctx, key, values...).Err()
	return err == nil
}

// MustHGet 获取哈希表中字段的值，出错返回空字符串
func (rc *Client) MustHGet(ctx context.Context, key, field string) string {
	result, err := rc.HGet(ctx, key, field).Result()
	if err != nil {
		return ""
	}
	return result
}

// MustHGetInt 获取哈希表中字段的整数值，出错返回0
func (rc *Client) MustHGetInt(ctx context.Context, key, field string) int {
	result, err := rc.HGet(ctx, key, field).Int()
	if err != nil {
		return 0
	}
	return result
}

// MustHGetInt64 获取哈希表中字段的int64值，出错返回0
func (rc *Client) MustHGetInt64(ctx context.Context, key, field string) int64 {
	result, err := rc.HGet(ctx, key, field).Int64()
	if err != nil {
		return 0
	}
	return result
}

// MustHDel 删除哈希表中的一个或多个字段，不返回错误
func (rc *Client) MustHDel(ctx context.Context, key string, fields ...string) bool {
	err := rc.HDel(ctx, key, fields...).Err()
	return err == nil
}

// MustHExists 检查哈希表中是否存在指定字段，出错返回false
func (rc *Client) MustHExists(ctx context.Context, key, field string) bool {
	result, err := rc.HExists(ctx, key, field).Result()
	return err == nil && result
}

// MustHGetAll 获取哈希表中的所有字段和值，出错返回nil
func (rc *Client) MustHGetAll(ctx context.Context, key string) map[string]string {
	result, err := rc.HGetAll(ctx, key).Result()
	if err != nil {
		return nil
	}
	return result
}

// MustLPush 将一个或多个值插入到列表头部，不返回错误
func (rc *Client) MustLPush(ctx context.Context, key string, values ...any) bool {
	err := rc.LPush(ctx, key, values...).Err()
	return err == nil
}

// MustRPush 将一个或多个值插入到列表尾部，不返回错误
func (rc *Client) MustRPush(ctx context.Context, key string, values ...any) bool {
	err := rc.RPush(ctx, key, values...).Err()
	return err == nil
}

// MustLPop 移出并获取列表的第一个元素，出错返回空字符串
func (rc *Client) MustLPop(ctx context.Context, key string) string {
	result, err := rc.LPop(ctx, key).Result()
	if err != nil {
		return ""
	}
	return result
}

// MustRPop 移出并获取列表的最后一个元素，出错返回空字符串
func (rc *Client) MustRPop(ctx context.Context, key string) string {
	result, err := rc.RPop(ctx, key).Result()
	if err != nil {
		return ""
	}
	return result
}

// MustLLen 获取列表长度，出错返回0
func (rc *Client) MustLLen(ctx context.Context, key string) int64 {
	result, err := rc.LLen(ctx, key).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustLIndex 通过索引获取列表中的元素，出错返回空字符串
func (rc *Client) MustLIndex(ctx context.Context, key string, index int64) string {
	result, err := rc.LIndex(ctx, key, index).Result()
	if err != nil {
		return ""
	}
	return result
}

// MustSAdd 向集合添加一个或多个成员，不返回错误
func (rc *Client) MustSAdd(ctx context.Context, key string, members ...any) bool {
	err := rc.SAdd(ctx, key, members...).Err()
	return err == nil
}

// MustSRem 移除集合中一个或多个成员，不返回错误
func (rc *Client) MustSRem(ctx context.Context, key string, members ...any) bool {
	err := rc.SRem(ctx, key, members...).Err()
	return err == nil
}

// MustSCard 获取集合的成员数，出错返回0
func (rc *Client) MustSCard(ctx context.Context, key string) int64 {
	result, err := rc.SCard(ctx, key).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustSIsMember 判断成员元素是否是集合的成员，出错返回false
func (rc *Client) MustSIsMember(ctx context.Context, key string, member any) bool {
	result, err := rc.SIsMember(ctx, key, member).Result()
	return err == nil && result
}

// MustSMembers 获取集合中的所有成员，出错返回nil
func (rc *Client) MustSMembers(ctx context.Context, key string) []string {
	result, err := rc.SMembers(ctx, key).Result()
	if err != nil {
		return nil
	}
	return result
}

// MustZAdd 向有序集合添加一个或多个成员，不返回错误
func (rc *Client) MustZAdd(ctx context.Context, key string, members ...redis.Z) bool {
	err := rc.ZAdd(ctx, key, members...).Err()
	return err == nil
}

// MustZRem 移除有序集合中的一个或多个成员，不返回错误
func (rc *Client) MustZRem(ctx context.Context, key string, members ...any) bool {
	err := rc.ZRem(ctx, key, members...).Err()
	return err == nil
}

// MustZCard 获取有序集合的成员数，出错返回0
func (rc *Client) MustZCard(ctx context.Context, key string) int64 {
	result, err := rc.ZCard(ctx, key).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustZScore 获取有序集合中指定成员的分数，出错返回0
func (rc *Client) MustZScore(ctx context.Context, key, member string) float64 {
	result, err := rc.ZScore(ctx, key, member).Result()
	if err != nil {
		return 0
	}
	return result
}

// MustZRange 获取有序集合中指定区间内的成员，出错返回nil
func (rc *Client) MustZRange(ctx context.Context, key string, start, stop int64) []string {
	result, err := rc.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil
	}
	return result
}

// MustZRangeWithScores 获取有序集合中指定区间内的成员和分数，出错返回nil
func (rc *Client) MustZRangeWithScores(ctx context.Context, key string, start, stop int64) []redis.Z {
	result, err := rc.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil
	}
	return result
}

// SetJSON 将对象以JSON格式存储到Redis
func (rc *Client) SetJSON(ctx context.Context, key string, obj any, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return rc.Set(ctx, key, data, expiration).Err()
}

// MustSetJSON 将对象以JSON格式存储到Redis，不返回错误
func (rc *Client) MustSetJSON(ctx context.Context, key string, obj any, expiration time.Duration) bool {
	err := rc.SetJSON(ctx, key, obj, expiration)
	return err == nil
}

// GetJSON 从Redis获取JSON数据并解析到指定对象
func (rc *Client) GetJSON(ctx context.Context, key string, obj any) error {
	data, err := rc.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), obj)
}

// MustGetJSON 从Redis获取JSON数据并解析到指定对象，不返回错误
func (rc *Client) MustGetJSON(ctx context.Context, key string, obj any) bool {
	err := rc.GetJSON(ctx, key, obj)
	return err == nil
}
