# Redis 客户端

本文档介绍了 Jupiter 框架中 Redis 客户端的使用方法，包括基本配置、扩展方法和功能开关控制。

## 目录

- [配置说明](#配置说明)
- [功能开关](#功能开关)
- [基本键值操作](#基本键值操作)
- [计数器操作](#计数器操作)
- [哈希表操作](#哈希表操作)
- [列表操作](#列表操作)
- [集合操作](#集合操作)
- [有序集合操作](#有序集合操作)
- [JSON 操作](#json-操作)
- [分布式锁操作](#分布式锁操作)

## 配置说明

### 基本配置

Redis 客户端的基本配置项如下：

```toml
[redis.default]
# Redis 连接地址
addr = "127.0.0.1:6379"
# 连接密码
password = ""
# 数据库编号
db = 0
# 连接池最大空闲连接数
maxIdle = 10
# 连接池最大活跃连接数
maxActive = 100
# 连接空闲超时时间
idleTimeout = "80s"
# 拨号超时时间
dialTimeout = "200ms"
# 读取超时时间
readTimeout = "200ms"
# 写入超时时间
writeTimeout = "200ms"
# 连接失败时的处理方式："panic" 或 "error"
onDialError = "error"
# 慢查询阈值，0 表示不记录
slowLogThreshold = "0ms"
```

### 功能开关

Redis 客户端提供了多个功能开关，可以根据需要启用或禁用特定功能：

```toml
[redis.default]
# 是否启用指标收集
enableMetric = false
# 是否启用链路追踪
enableTrace = false
# 是否启用访问日志
enableAccessLog = false
# 是否启用哨兵模式
enableSentinel = false
```

### 完整配置示例

```toml
[redis.default]
addr = "127.0.0.1:6379"
password = ""
db = 0
maxIdle = 10
maxActive = 100
idleTimeout = "80s"
dialTimeout = "200ms"
readTimeout = "200ms"
writeTimeout = "200ms"
onDialError = "error"
slowLogThreshold = "100ms"
enableMetric = true
enableTrace = true
enableAccessLog = true
enableSentinel = false
```

### 代码中使用配置

```go
// 使用默认配置
client := redis.StdConfig("default").Build()

// 使用自定义配置
config := &redis.Config{
    Addr:               "127.0.0.1:6379",
    Password:           "",
    DB:                 0,
    MaxIdle:            10,
    MaxActive:          100,
    IdleTimeout:        80 * time.Second,
    DialTimeout:        200 * time.Millisecond,
    ReadTimeout:        200 * time.Millisecond,
    WriteTimeout:       200 * time.Millisecond,
    OnDialError:        "error",
    SlowLogThreshold:   100 * time.Millisecond,
    EnableMetric:       true,
    EnableTrace:        true,
    EnableAccessLog:    true,
    EnableSentinel:     false,
}
client := redis.Build(config)
```

## 基本键值操作

### MustSet

设置键值对，不返回错误。

```go
func (rc *Client) MustSet(ctx context.Context, key string, value any, expiration time.Duration) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `value`: 值
- `expiration`: 过期时间

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustSet(ctx, "user:1", "张三", time.Hour)
```

### MustSetNX

设置键值对（仅当键不存在时），不返回错误。

```go
func (rc *Client) MustSetNX(ctx context.Context, key string, value any, expiration time.Duration) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `value`: 值
- `expiration`: 过期时间

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
success := client.MustSetNX(ctx, "lock:resource1", "locked", 10*time.Second)
```

### MustDel

删除键，不返回错误。

```go
func (rc *Client) MustDel(ctx context.Context, keys ...string) bool
```

**参数:**
- `ctx`: 上下文
- `keys`: 一个或多个键名

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustDel(ctx, "user:1", "user:2")
```

### MustExpire

设置键的过期时间，不返回错误。

```go
func (rc *Client) MustExpire(ctx context.Context, key string, expiration time.Duration) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `expiration`: 过期时间

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustExpire(ctx, "session:abc123", 30*time.Minute)
```

### MustTTL

获取键的剩余生存时间，出错返回-1。

```go
func (rc *Client) MustTTL(ctx context.Context, key string) time.Duration
```

**参数:**
- `ctx`: 上下文
- `key`: 键名

**返回值:**
- `time.Duration`: 剩余生存时间，出错返回-1

**示例:**
```go
ttl := client.MustTTL(ctx, "session:abc123")
```

### MustExists

检查键是否存在，出错返回false。

```go
func (rc *Client) MustExists(ctx context.Context, key string) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名

**返回值:**
- `bool`: 键是否存在

**示例:**
```go
exists := client.MustExists(ctx, "user:1")
```

## 计数器操作

### MustIncr

递增键的值，出错返回0。

```go
func (rc *Client) MustIncr(ctx context.Context, key string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 键名

**返回值:**
- `int64`: 递增后的值，出错返回0

**示例:**
```go
count := client.MustIncr(ctx, "page:views")
```

### MustIncrBy

按指定值递增键的值，出错返回0。

```go
func (rc *Client) MustIncrBy(ctx context.Context, key string, value int64) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `value`: 递增值

**返回值:**
- `int64`: 递增后的值，出错返回0

**示例:**
```go
count := client.MustIncrBy(ctx, "page:views", 10)
```

### MustDecr

递减键的值，出错返回0。

```go
func (rc *Client) MustDecr(ctx context.Context, key string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 键名

**返回值:**
- `int64`: 递减后的值，出错返回0

**示例:**
```go
count := client.MustDecr(ctx, "inventory:item:123")
```

### MustDecrBy

按指定值递减键的值，出错返回0。

```go
func (rc *Client) MustDecrBy(ctx context.Context, key string, value int64) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `value`: 递减值

**返回值:**
- `int64`: 递减后的值，出错返回0

**示例:**
```go
count := client.MustDecrBy(ctx, "inventory:item:123", 5)
```

## 哈希表操作

### MustHSet

设置哈希表中的字段值，不返回错误。

```go
func (rc *Client) MustHSet(ctx context.Context, key string, values ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `values`: 字段和值的交替序列

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustHSet(ctx, "user:1", "name", "张三", "age", 30, "email", "zhangsan@example.com")
```

### MustHGet

获取哈希表中字段的值，出错返回空字符串。

```go
func (rc *Client) MustHGet(ctx context.Context, key, field string) string
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `field`: 字段名

**返回值:**
- `string`: 字段值，出错返回空字符串

**示例:**
```go
name := client.MustHGet(ctx, "user:1", "name")
```

### MustHGetInt

获取哈希表中字段的整数值，出错返回0。

```go
func (rc *Client) MustHGetInt(ctx context.Context, key, field string) int
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `field`: 字段名

**返回值:**
- `int`: 字段的整数值，出错返回0

**示例:**
```go
age := client.MustHGetInt(ctx, "user:1", "age")
```

### MustHGetInt64

获取哈希表中字段的int64值，出错返回0。

```go
func (rc *Client) MustHGetInt64(ctx context.Context, key, field string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `field`: 字段名

**返回值:**
- `int64`: 字段的int64值，出错返回0

**示例:**
```go
timestamp := client.MustHGetInt64(ctx, "user:1", "created_at")
```

### MustHDel

删除哈希表中的一个或多个字段，不返回错误。

```go
func (rc *Client) MustHDel(ctx context.Context, key string, fields ...string) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `fields`: 一个或多个字段名

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustHDel(ctx, "user:1", "email", "phone")
```

### MustHExists

检查哈希表中是否存在指定字段，出错返回false。

```go
func (rc *Client) MustHExists(ctx context.Context, key, field string) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `field`: 字段名

**返回值:**
- `bool`: 字段是否存在

**示例:**
```go
hasName := client.MustHExists(ctx, "user:1", "name")
```

### MustHGetAll

获取哈希表中的所有字段和值，出错返回nil。

```go
func (rc *Client) MustHGetAll(ctx context.Context, key string) map[string]string
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名

**返回值:**
- `map[string]string`: 所有字段和值，出错返回nil

**示例:**
```go
userData := client.MustHGetAll(ctx, "user:1")
```

### HGetAllStruct

获取哈希表中的所有字段和值，并扫描到结构体。

```go
func (rc *Client) HGetAllStruct(ctx context.Context, key string, to any) error
```

**参数:**
- `ctx`: 上下文
- `key`: 哈希表键名
- `to`: 目标结构体指针

**返回值:**
- `error`: 错误信息

**示例:**
```go
type User struct {
    Name  string `redis:"name"`
    Age   int    `redis:"age"`
    Email string `redis:"email"`
}
var user User
err := client.HGetAllStruct(ctx, "user:1", &user)
```

## 列表操作

### MustLPush

将一个或多个值插入到列表头部，不返回错误。

```go
func (rc *Client) MustLPush(ctx context.Context, key string, values ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名
- `values`: 一个或多个值

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustLPush(ctx, "messages", "msg1", "msg2", "msg3")
```

### MustRPush

将一个或多个值插入到列表尾部，不返回错误。

```go
func (rc *Client) MustRPush(ctx context.Context, key string, values ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名
- `values`: 一个或多个值

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustRPush(ctx, "messages", "msg4", "msg5")
```

### MustLPop

移出并获取列表的第一个元素，出错返回空字符串。

```go
func (rc *Client) MustLPop(ctx context.Context, key string) string
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名

**返回值:**
- `string`: 列表的第一个元素，出错返回空字符串

**示例:**
```go
firstMsg := client.MustLPop(ctx, "messages")
```

### MustRPop

移出并获取列表的最后一个元素，出错返回空字符串。

```go
func (rc *Client) MustRPop(ctx context.Context, key string) string
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名

**返回值:**
- `string`: 列表的最后一个元素，出错返回空字符串

**示例:**
```go
lastMsg := client.MustRPop(ctx, "messages")
```

### MustLLen

获取列表长度，出错返回0。

```go
func (rc *Client) MustLLen(ctx context.Context, key string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名

**返回值:**
- `int64`: 列表长度，出错返回0

**示例:**
```go
length := client.MustLLen(ctx, "messages")
```

### MustLIndex

通过索引获取列表中的元素，出错返回空字符串。

```go
func (rc *Client) MustLIndex(ctx context.Context, key string, index int64) string
```

**参数:**
- `ctx`: 上下文
- `key`: 列表键名
- `index`: 索引位置

**返回值:**
- `string`: 指定索引的元素，出错返回空字符串

**示例:**
```go
secondMsg := client.MustLIndex(ctx, "messages", 1)
```

## 集合操作

### MustSAdd

向集合添加一个或多个成员，不返回错误。

```go
func (rc *Client) MustSAdd(ctx context.Context, key string, members ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 集合键名
- `members`: 一个或多个成员

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustSAdd(ctx, "tags", "golang", "redis", "database")
```

### MustSRem

移除集合中一个或多个成员，不返回错误。

```go
func (rc *Client) MustSRem(ctx context.Context, key string, members ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 集合键名
- `members`: 一个或多个成员

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustSRem(ctx, "tags", "database")
```

### MustSCard

获取集合的成员数，出错返回0。

```go
func (rc *Client) MustSCard(ctx context.Context, key string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 集合键名

**返回值:**
- `int64`: 集合成员数，出错返回0

**示例:**
```go
count := client.MustSCard(ctx, "tags")
```

### MustSIsMember

判断成员元素是否是集合的成员，出错返回false。

```go
func (rc *Client) MustSIsMember(ctx context.Context, key string, member any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 集合键名
- `member`: 成员

**返回值:**
- `bool`: 成员是否存在于集合中

**示例:**
```go
isMember := client.MustSIsMember(ctx, "tags", "golang")
```

### MustSMembers

获取集合中的所有成员，出错返回nil。

```go
func (rc *Client) MustSMembers(ctx context.Context, key string) []string
```

**参数:**
- `ctx`: 上下文
- `key`: 集合键名

**返回值:**
- `[]string`: 集合中的所有成员，出错返回nil

**示例:**
```go
allTags := client.MustSMembers(ctx, "tags")
```

## 有序集合操作

### MustZAdd

向有序集合添加一个或多个成员，不返回错误。

```go
func (rc *Client) MustZAdd(ctx context.Context, key string, members ...redis.Z) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名
- `members`: 一个或多个 redis.Z 结构体

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustZAdd(ctx, "leaderboard", 
    redis.Z{Score: 100, Member: "user1"},
    redis.Z{Score: 200, Member: "user2"},
    redis.Z{Score: 150, Member: "user3"})
```

### MustZRem

移除有序集合中的一个或多个成员，不返回错误。

```go
func (rc *Client) MustZRem(ctx context.Context, key string, members ...any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名
- `members`: 一个或多个成员

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
client.MustZRem(ctx, "leaderboard", "user2")
```

### MustZCard

获取有序集合的成员数，出错返回0。

```go
func (rc *Client) MustZCard(ctx context.Context, key string) int64
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名

**返回值:**
- `int64`: 有序集合成员数，出错返回0

**示例:**
```go
count := client.MustZCard(ctx, "leaderboard")
```

### MustZScore

获取有序集合中指定成员的分数，出错返回0。

```go
func (rc *Client) MustZScore(ctx context.Context, key, member string) float64
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名
- `member`: 成员

**返回值:**
- `float64`: 成员的分数，出错返回0

**示例:**
```go
score := client.MustZScore(ctx, "leaderboard", "user1")
```

### MustZRange

获取有序集合中指定区间内的成员，出错返回nil。

```go
func (rc *Client) MustZRange(ctx context.Context, key string, start, stop int64) []string
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名
- `start`: 起始索引
- `stop`: 结束索引

**返回值:**
- `[]string`: 指定区间内的成员，出错返回nil

**示例:**
```go
topUsers := client.MustZRange(ctx, "leaderboard", 0, 9) // 获取前10名
```

### MustZRangeWithScores

获取有序集合中指定区间内的成员和分数，出错返回nil。

```go
func (rc *Client) MustZRangeWithScores(ctx context.Context, key string, start, stop int64) []redis.Z
```

**参数:**
- `ctx`: 上下文
- `key`: 有序集合键名
- `start`: 起始索引
- `stop`: 结束索引

**返回值:**
- `[]redis.Z`: 指定区间内的成员和分数，出错返回nil

**示例:**
```go
topUsersWithScores := client.MustZRangeWithScores(ctx, "leaderboard", 0, 9)
for _, user := range topUsersWithScores {
    fmt.Printf("用户: %v, 分数: %f\n", user.Member, user.Score)
}
```

## JSON 操作

### SetJSON

将对象以JSON格式存储到Redis。

```go
func (rc *Client) SetJSON(ctx context.Context, key string, obj any, expiration time.Duration) error
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `obj`: 要存储的对象
- `expiration`: 过期时间

**返回值:**
- `error`: 错误信息

**示例:**
```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
user := User{Name: "张三", Age: 30}
err := client.SetJSON(ctx, "user:1", user, time.Hour)
```

### MustSetJSON

将对象以JSON格式存储到Redis，不返回错误。

```go
func (rc *Client) MustSetJSON(ctx context.Context, key string, obj any, expiration time.Duration) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `obj`: 要存储的对象
- `expiration`: 过期时间

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
success := client.MustSetJSON(ctx, "user:1", user, time.Hour)
```

### GetJSON

从Redis获取JSON数据并解析到指定对象。

```go
func (rc *Client) GetJSON(ctx context.Context, key string, obj any) error
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `obj`: 目标对象指针

**返回值:**
- `error`: 错误信息

**示例:**
```go
var user User
err := client.GetJSON(ctx, "user:1", &user)
```

### MustGetJSON

从Redis获取JSON数据并解析到指定对象，不返回错误。

```go
func (rc *Client) MustGetJSON(ctx context.Context, key string, obj any) bool
```

**参数:**
- `ctx`: 上下文
- `key`: 键名
- `obj`: 目标对象指针

**返回值:**
- `bool`: 操作是否成功

**示例:**
```go
var user User
success := client.MustGetJSON(ctx, "user:1", &user)
```

## 注意事项

1. 所有 `Must` 前缀的方法在出错时不会返回错误，而是返回默认值或布尔值表示操作是否成功。
2. 对于需要精确错误处理的场景，建议使用原始的 Redis 客户端方法。
3. JSON 操作方法会自动处理对象的序列化和反序列化。
4. 所有方法都支持上下文传递，可以用于超时控制和取消操作。

## 示例代码

完整的示例代码可以参考 `example_usage.go` 文件，其中展示了如何使用这些扩展方法。

## 分布式锁操作

Jupiter Redis 客户端提供了功能完善的分布式锁实现，支持锁的获取、释放、续约等操作，并提供了多种安全机制防止死锁和误删。

### 配置分布式锁

分布式锁的配置可以通过配置文件设置默认值：

```toml
[redis.default]
# 基本Redis配置...
addr = "127.0.0.1:6379"
password = ""
db = 0

# 分布式锁默认配置
# 默认锁过期时间
lockDefaultExpiration = "30s"
# 默认重试间隔
lockDefaultRetryDelay = "100ms"
# 默认最大重试次数，-1表示无限重试
lockDefaultMaxRetries = -1
# 默认获取锁超时时间，0表示不超时
lockDefaultTimeout = "0s"
```

### 创建分布式锁

```go
// 使用默认配置创建锁
lock := client.Newlock("resource_lock")

// 使用自定义选项创建锁
customOptions := &redis.LockOptions{
    Expiration: 10 * time.Second,  // 锁过期时间
    RetryDelay: 200 * time.Millisecond, // 重试间隔
    MaxRetries: 5,                   // 最大重试次数
    Timeout:    5 * time.Second,     // 获取锁超时时间
}
lock := client.Newlock("resource_lock", customOptions)
```

### 获取和释放锁

```go
// 阻塞获取锁
err := lock.Lock(ctx)
if err != nil {
    // 处理错误
    log.Printf("获取锁失败: %v", err)
    return
}
defer lock.Unlock(ctx) // 确保锁被释放

// 执行需要互斥的操作
// ...
```

### 非阻塞尝试获取锁

```go
// 尝试获取锁，非阻塞
acquired, err := lock.TryLock(ctx)
if err != nil {
    // 处理错误
    log.Printf("尝试获取锁失败: %v", err)
    return
}

if acquired {
    defer lock.Unlock(ctx)
    // 执行需要互斥的操作
    // ...
} else {
    // 锁已被其他进程持有
    log.Println("锁已被其他进程持有")
}
```

### 锁续约

```go
// 获取锁
err := lock.Lock(ctx)
if err != nil {
    return
}
defer lock.Unlock(ctx)

// 在长时间操作中续约锁
go func() {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // 续约锁，延长过期时间
            err := lock.Renew(ctx, 30*time.Second)
            if err != nil {
                log.Printf("续约锁失败: %v", err)
                return
            }
        case <-ctx.Done():
            return
        }
    }
}()

// 执行长时间操作
// ...
```

### 使用 WithLock 简化操作

```go
// 使用 WithLock 自动处理锁的获取和释放
err := client.WithLock(ctx, "resource_lock", func() error {
    // 在锁保护下执行操作
    // 如果函数返回错误，锁仍然会被释放
    return doSomething()
})
if err != nil {
    log.Printf("执行操作失败: %v", err)
}
```

### 使用 TryWithLock 非阻塞执行

```go
// 使用 TryWithLock 尝试在锁保护下执行操作
executed, err := client.TryWithLock(ctx, "resource_lock", func() error {
    // 在锁保护下执行操作
    return doSomething()
})

if err != nil {
    log.Printf("执行操作失败: %v", err)
} else if executed {
    log.Println("操作成功执行")
} else {
    log.Println("未能获取锁，操作未执行")
}
```

### 并发安全示例

```go
// 多个goroutine安全地访问共享资源
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        err := client.WithLock(ctx, "shared_resource", func() error {
            // 安全地访问共享资源
            value := client.MustGet(ctx, "shared_counter")
            counter, _ := strconv.Atoi(value)
            counter++
            client.MustSet(ctx, "shared_counter", counter, 0)
            
            log.Printf("Goroutine %d 更新计数器为 %d", id, counter)
            return nil
        })
        
        if err != nil {
            log.Printf("Goroutine %d 操作失败: %v", id, err)
        }
    }(i)
}

wg.Wait()
```

### 分布式锁特性

1. **原子性保证**：使用 Redis 的 SET NX EX 命令确保锁获取的原子性
2. **防误删机制**：每个锁有唯一标识，只有锁的持有者才能释放锁
3. **自动续约**：支持自动续约机制，防止长时间操作导致锁过期
4. **超时控制**：支持获取锁的超时控制，避免无限等待
5. **可配置性**：支持全局配置和单次自定义配置
6. **容错处理**：完善的错误处理和日志记录

### 注意事项

1. 锁的过期时间应根据业务操作的实际耗时合理设置
2. 对于可能耗时较长的操作，建议使用续约机制或设置较长的过期时间
3. 确保在 finally 块或 defer 中释放锁，避免死锁
4. 分布式锁不能替代数据库事务，仅用于控制并发访问
5. 在高并发场景下，锁的竞争可能导致性能下降，应合理设计