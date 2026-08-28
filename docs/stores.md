# Jupiter 存储组件

本文档介绍 Jupiter 框架中的各种存储组件，包括 GORM、MongoDB、MeiliSearch 等，以及它们的使用方法和配置选项。

## 概述

Jupiter 框架提供了丰富的存储组件，用于与各种数据库进行交互。这些存储组件封装了连接管理、事务处理、查询构建等功能，简化了数据库操作。框架提供了统一的配置和接口，使开发者可以轻松地使用各种存储组件。

## GORM 存储

### 基本使用

```go
// 创建 GORM 数据库连接
db := gorm.StdConfig("mysql").Build()

// 定义模型
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100;not null"`
    Age  int    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 创建表
err := db.AutoMigrate(&User{})
if err != nil {
    panic(err)
}

// 插入数据
user := User{Name: "John", Age: 30}
result := db.Create(&user)
if result.Error != nil {
    panic(result.Error)
}

// 查询数据
var users []User
result = db.Find(&users)
if result.Error != nil {
    panic(result.Error)
}
for _, user := range users {
    fmt.Printf("用户: %s, 年龄: %d\n", user.Name, user.Age)
}
```

### 配置选项

```toml
# MySQL 配置
[store.mysql]
driver = "mysql"
dsn = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = "1h"
conn_max_idle_time = "10m"

# PostgreSQL 配置
[store.postgres]
driver = "postgres"
dsn = "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = "1h"
conn_max_idle_time = "10m"

# SQLite 配置
[store.sqlite]
driver = "sqlite"
dsn = "file::memory:?cache=shared"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = "1h"
conn_max_idle_time = "10m"
```

### 模型定义

```go
// 基本模型
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100;not null"`
    Age  int    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 关联模型
type Company struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:100;not null"`
    Users []User `gorm:"foreignKey:CompanyID"`
}

type User struct {
    ID        uint     `gorm:"primaryKey"`
    Name      string   `gorm:"size:100;not null"`
    Age       int      `gorm:"not null"`
    CompanyID uint     `gorm:"not null"`
    Company   Company  `gorm:"foreignKey:CompanyID"`
    Roles     []Role   `gorm:"many2many:user_roles;"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Role struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:50;not null;uniqueIndex"`
    Users     []User    `gorm:"many2many:user_roles;"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 查询操作

```go
// 基本查询
var user User
result := db.First(&user, 1) // 查询ID为1的用户
if result.Error != nil {
    panic(result.Error)
}

// 条件查询
var users []User
result = db.Where("age > ?", 18).Find(&users)
if result.Error != nil {
    panic(result.Error)
}

// 复杂查询
result = db.Where("name LIKE ? AND age BETWEEN ? AND ?", "%John%", 20, 30).Find(&users)
if result.Error != nil {
    panic(result.Error)
}

// 结构体查询
result = db.Where(&User{Name: "John", Age: 30}).First(&user)
if result.Error != nil {
    panic(result.Error)
}

// Map查询
result = db.Where(map[string]interface{}{"name": "John", "age": 30}).First(&user)
if result.Error != nil {
    panic(result.Error)
}

// 原生SQL查询
result = db.Raw("SELECT * FROM users WHERE name = ?", "John").Scan(&user)
if result.Error != nil {
    panic(result.Error)
}
```

### 更新操作

```go
// 更新单个字段
result := db.Model(&User{}).Where("id = ?", 1).Update("name", "John Doe")
if result.Error != nil {
    panic(result.Error)
}

// 更新多个字段
result = db.Model(&User{}).Where("id = ?", 1).Updates(User{Name: "John Doe", Age: 35})
if result.Error != nil {
    panic(result.Error)
}

// 使用Map更新
result = db.Model(&User{}).Where("id = ?", 1).Updates(map[string]interface{}{"name": "John Doe", "age": 35})
if result.Error != nil {
    panic(result.Error)
}

// 更新选中字段
result = db.Model(&User{}).Select("name").Where("id = ?", 1).Updates(User{Name: "John Doe"})
if result.Error != nil {
    panic(result.Error)
}

// 忽略字段
result = db.Model(&User{}).Omit("age").Where("id = ?", 1).Updates(User{Name: "John Doe", Age: 35})
if result.Error != nil {
    panic(result.Error)
}
```

### 删除操作

```go
// 删除记录
result := db.Delete(&User{}, 1) // 删除ID为1的用户
if result.Error != nil {
    panic(result.Error)
}

// 条件删除
result = db.Where("age < ?", 18).Delete(&User{})
if result.Error != nil {
    panic(result.Error)
}

// 软删除
result = db.Delete(&user) // 软删除，设置DeletedAt字段
if result.Error != nil {
    panic(result.Error)
}

// 永久删除
result = db.Unscoped().Delete(&user) // 永久删除
if result.Error != nil {
    panic(result.Error)
}
```

### 事务处理

```go
// 自动事务
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&User{Name: "John", Age: 30}).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&User{Name: "Jane", Age: 25}).Error; err != nil {
        return err
    }
    
    return nil
})
if err != nil {
    panic(err)
}

// 手动事务
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
    }
}()

// 在事务中执行操作
if err := tx.Create(&User{Name: "John", Age: 30}).Error; err != nil {
    tx.Rollback()
    panic(err)
}

if err := tx.Create(&User{Name: "Jane", Age: 25}).Error; err != nil {
    tx.Rollback()
    panic(err)
}

// 提交事务
if err := tx.Commit().Error; err != nil {
    panic(err)
}
```

### 钩子函数

```go
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100;not null"`
    Age  int    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 验证数据
    if u.Name == "" {
        return errors.New("名称不能为空")
    }
    if u.Age < 0 || u.Age > 120 {
        return errors.New("年龄必须在0-120之间")
    }
    return nil
}

// 创建后钩子
func (u *User) AfterCreate(tx *gorm.DB) error {
    // 记录日志
    fmt.Printf("用户 %s 创建成功\n", u.Name)
    return nil
}

// 更新前钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 验证数据
    if u.Name == "" {
        return errors.New("名称不能为空")
    }
    if u.Age < 0 || u.Age > 120 {
        return errors.New("年龄必须在0-120之间")
    }
    return nil
}

// 更新后钩子
func (u *User) AfterUpdate(tx *gorm.DB) error {
    // 记录日志
    fmt.Printf("用户 %s 更新成功\n", u.Name)
    return nil
}
```

## MongoDB 存储

### 基本使用

```go
// 创建 MongoDB 连接
mongo := mongox.StdConfig("default").Build()

// 获取集合
collection := mongo.Database("testdb").Collection("users")

// 定义文档结构
type User struct {
    ID   primitive.ObjectID `bson:"_id,omitempty"`
    Name string             `bson:"name"`
    Age  int                `bson:"age"`
    CreatedAt time.Time      `bson:"created_at"`
    UpdatedAt time.Time      `bson:"updated_at"`
}

// 插入文档
user := User{
    Name: "John",
    Age:  30,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}
result, err := collection.InsertOne(context.Background(), user)
if err != nil {
    panic(err)
}
fmt.Printf("插入ID: %s\n", result.InsertedID)

// 查询文档
var resultUser User
err = collection.FindOne(context.Background(), bson.M{"name": "John"}).Decode(&resultUser)
if err != nil {
    panic(err)
}
fmt.Printf("用户: %s, 年龄: %d\n", resultUser.Name, resultUser.Age)
```

### 配置选项

```toml
# MongoDB 配置
[store.mongodb.default]
uri = "mongodb://localhost:27017"
database = "testdb"
max_pool_size = 100
min_pool_size = 10
max_conn_idle_time = "5m"
connect_timeout = "10s"
server_selection_timeout = "30s"

# 副本集配置
[store.mongodb.replica_set]
uri = "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0"
database = "testdb"
max_pool_size = 100
min_pool_size = 10
max_conn_idle_time = "5m"
connect_timeout = "10s"
server_selection_timeout = "30s"

# 分片集群配置
[store.mongodb.sharded_cluster]
uri = "mongodb://mongos1:27017,mongos2:27017,mongos3:27017"
database = "testdb"
max_pool_size = 100
min_pool_size = 10
max_conn_idle_time = "5m"
connect_timeout = "10s"
server_selection_timeout = "30s"
```

### 查询操作

```go
// 基本查询
var user User
err := collection.FindOne(context.Background(), bson.M{"name": "John"}).Decode(&user)
if err != nil {
    panic(err)
}

// 条件查询
cursor, err := collection.Find(context.Background(), bson.M{"age": bson.M{"$gt": 18}})
if err != nil {
    panic(err)
}
defer cursor.Close(context.Background())

var users []User
for cursor.Next(context.Background()) {
    var user User
    if err := cursor.Decode(&user); err != nil {
        panic(err)
    }
    users = append(users, user)
}

// 复杂查询
filter := bson.M{
    "age": bson.M{"$gte": 18, "$lte": 30},
    "name": bson.M{"$regex": primitive.Regex{Pattern: "^J", Options: "i"}},
}
cursor, err = collection.Find(context.Background(), filter)
if err != nil {
    panic(err)
}
defer cursor.Close(context.Background())

// 排序
opts := options.Find().SetSort(bson.M{"age": 1, "name": 1})
cursor, err = collection.Find(context.Background(), bson.M{}, opts)
if err != nil {
    panic(err)
}
defer cursor.Close(context.Background())

// 分页
skip := (page - 1) * size
opts = options.Find().SetSort(bson.M{"_id": -1}).SetSkip(int64(skip)).SetLimit(int64(size))
cursor, err = collection.Find(context.Background(), bson.M{}, opts)
if err != nil {
    panic(err)
}
defer cursor.Close(context.Background())
```

### 更新操作

```go
// 更新单个文档
filter := bson.M{"name": "John"}
update := bson.M{"$set": bson.M{"age": 35}}
result, err := collection.UpdateOne(context.Background(), filter, update)
if err != nil {
    panic(err)
}
fmt.Printf("匹配文档数: %d, 更新文档数: %d\n", result.MatchedCount, result.ModifiedCount)

// 更新多个文档
filter = bson.M{"age": bson.M{"$lt": 18}}
update = bson.M{"$set": bson.M{"status": "minor"}}
result, err = collection.UpdateMany(context.Background(), filter, update)
if err != nil {
    panic(err)
}
fmt.Printf("匹配文档数: %d, 更新文档数: %d\n", result.MatchedCount, result.ModifiedCount)

// 替换文档
filter = bson.M{"name": "John"}
replacement := User{
    Name: "John Doe",
    Age:  35,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}
result, err = collection.ReplaceOne(context.Background(), filter, replacement)
if err != nil {
    panic(err)
}
fmt.Printf("匹配文档数: %d, 更新文档数: %d\n", result.MatchedCount, result.ModifiedCount)

// 更新或插入
filter = bson.M{"name": "Alice"}
update = bson.M{"$set": bson.M{"age": 25}}
opts := options.Update().SetUpsert(true)
result, err = collection.UpdateOne(context.Background(), filter, update, opts)
if err != nil {
    panic(err)
}
fmt.Printf("匹配文档数: %d, 更新文档数: %d, 插入ID: %v\n", 
    result.MatchedCount, result.ModifiedCount, result.UpsertedID)
```

### 删除操作

```go
// 删除单个文档
filter := bson.M{"name": "John"}
result, err := collection.DeleteOne(context.Background(), filter)
if err != nil {
    panic(err)
}
fmt.Printf("删除文档数: %d\n", result.DeletedCount)

// 删除多个文档
filter = bson.M{"age": bson.M{"$lt": 18}}
result, err = collection.DeleteMany(context.Background(), filter)
if err != nil {
    panic(err)
}
fmt.Printf("删除文档数: %d\n", result.DeletedCount)
```

### 事务处理

```go
// 开始会话
session, err := mongo.StartSession()
if err != nil {
    panic(err)
}
defer session.EndSession(context.Background())

// 在事务中执行操作
err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
    // 开始事务
    if err := sc.StartTransaction(); err != nil {
        return err
    }
    
    // 在事务中执行操作
    if _, err := collection.InsertOne(sc, User{Name: "John", Age: 30}); err != nil {
        // 中止事务
        _ = sc.AbortTransaction(sc)
        return err
    }
    
    if _, err := collection.InsertOne(sc, User{Name: "Jane", Age: 25}); err != nil {
        // 中止事务
        _ = sc.AbortTransaction(sc)
        return err
    }
    
    // 提交事务
    if err := sc.CommitTransaction(sc); err != nil {
        return err
    }
    
    return nil
})
if err != nil {
    panic(err)
}
```

## 存储最佳实践

### 1. 连接池配置

合理配置连接池大小：

```go
// GORM 连接池配置
db := gorm.New(
    gorm.WithConfig("mysql"),
    gorm.WithMaxOpenConns(100),       // 最大打开连接数
    gorm.WithMaxIdleConns(10),        // 最大空闲连接数
    gorm.WithConnMaxLifetime(time.Hour), // 连接最大生存时间
    gorm.WithConnMaxIdleTime(10*time.Minute), // 连接最大空闲时间
)

// MongoDB 连接池配置
mongo := mongox.New(
    mongox.WithConfig("default"),
    mongox.WithMaxPoolSize(100),      // 最大连接池大小
    mongox.WithMinPoolSize(10),       // 最小连接池大小
    mongox.WithMaxConnIdleTime(5*time.Minute), // 连接最大空闲时间
)
```

### 2. 事务处理

合理使用事务：

```go
// GORM 事务
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&User{Name: "John", Age: 30}).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&User{Name: "Jane", Age: 25}).Error; err != nil {
        return err
    }
    
    return nil
})
if err != nil {
    panic(err)
}

// MongoDB 事务
session, err := mongo.StartSession()
if err != nil {
    panic(err)
}
defer session.EndSession(context.Background())

err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
    if err := sc.StartTransaction(); err != nil {
        return err
    }
    
    if _, err := collection.InsertOne(sc, User{Name: "John", Age: 30}); err != nil {
        _ = sc.AbortTransaction(sc)
        return err
    }
    
    if _, err := collection.InsertOne(sc, User{Name: "Jane", Age: 25}); err != nil {
        _ = sc.AbortTransaction(sc)
        return err
    }
    
    return sc.CommitTransaction(sc)
})
if err != nil {
    panic(err)
}
```

### 3. 索引优化

合理创建索引：

```go
// GORM 索引
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100;not null;index"`
    Age  int    `gorm:"not null;index"`
    Email string `gorm:"size:100;not null;uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// MongoDB 索引
indexModel := mongo.IndexModel{
    Keys: bson.M{"name": 1, "age": 1},
    Options: options.Index().SetName("name_age_index"),
}
_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
if err != nil {
    panic(err)
}
```

### 4. 查询优化

优化查询性能：

```go
// GORM 查询优化
var users []User
// 使用 Select 指定字段
result := db.Select("id", "name", "age").Where("age > ?", 18).Find(&users)
if result.Error != nil {
    panic(result.Error)
}

// 使用 Preload 预加载关联
result = db.Preload("Company").Preload("Roles").Find(&users)
if result.Error != nil {
    panic(result.Error)
}

// MongoDB 查询优化
projection := bson.M{"name": 1, "age": 1}
cursor, err := collection.Find(context.Background(), bson.M{"age": bson.M{"$gt": 18}}, options.Find().SetProjection(projection))
if err != nil {
    panic(err)
}
defer cursor.Close(context.Background())
```

## 总结

Jupiter 框架提供了丰富的存储组件，支持多种数据库和搜索引擎。通过统一的配置和接口，简化了数据存储和查询操作。合理使用这些存储组件，可以构建高性能、高可用的数据存储系统。

## 下一步

- 了解 [MeiliSearch 存储](./meilisearch.md)
- 探索 [缓存组件](./caches.md)
- 查看 [工具函数](./utils.md)
- 学习 [日志组件](./xlog.md)