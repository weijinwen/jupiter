# gorm

## 特性

- 全功能 ORM
- 关联 (Has One，Has Many，Belongs To，Many To Many，多态，单表继承)
- Create，Save，Update，Delete，Find 中钩子方法
- 支持 Preload、Joins 的预加载
- 事务，嵌套事务，Save Point，Rollback To Saved Point
- Context、预编译模式、DryRun 模式
- 批量插入，FindInBatches，Find/Create with Map，使用 SQL 表达式、Context Valuer 进行 CRUD
- SQL 构建器，Upsert，数据库锁，Optimizer/Index/Comment Hint，命名参数，子查询
- 复合主键，索引，约束
- Auto Migration
- 自定义 Logger
- 灵活的可扩展插件 API：Database Resolver（多数据库，读写分离）、Prometheus…
- 每个特性都经过了测试的重重考验
- 开发者友好
- 支持 MySQL 和 PostgreSQL 数据库

## 配置说明

### 基本配置

GORM 组件的基本配置项如下：

```toml
[database.mysql_example]
# 数据库连接字符串
dsn = "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
# 是否启用调试模式
debug = true
# 连接池最大空闲连接数
maxIdleConns = 10
# 连接池最大活跃连接数
maxOpenConns = 100
# 连接最大生命周期
connMaxLifetime = "300s"
# 慢查询阈值，0 表示不记录
slowThreshold = "200ms"
# 是否禁用外键约束
disableForeignKeyConstraintWhenMigrating = false
```

### 功能开关

GORM 组件提供了多个功能开关，可以根据需要启用或禁用特定功能：

```toml
[database.mysql_example]
# 是否启用指标收集
enableMetric = false
# 是否启用链路追踪
enableTrace = false
# 是否启用访问日志
enableAccessLog = false
# 是否启用数据库日志
enableDBLog = false
```

### 完整配置示例

```toml
[database.mysql_example]
dsn = "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
debug = true
maxIdleConns = 10
maxOpenConns = 100
connMaxLifetime = "300s"
slowThreshold = "200ms"
disableForeignKeyConstraintWhenMigrating = false
enableMetric = true
enableTrace = true
enableAccessLog = true
enableDBLog = true
```

### 代码中使用配置

```go
// 使用默认配置
db := gorm.StdConfig("mysql_example").Build()

// 使用自定义配置
config := &gorm.Config{
    DSN:                                  "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    Debug:                                true,
    MaxIdleConns:                         10,
    MaxOpenConns:                         100,
    ConnMaxLifetime:                      300 * time.Second,
    SlowThreshold:                        200 * time.Millisecond,
    DisableForeignKeyConstraintWhenMigrating: false,
    EnableMetric:                         true,
    EnableTrace:                          true,
    EnableAccessLog:                      true,
    EnableDBLog:                          true,
}
db := gorm.Build(config)
```

## 快速入门

### MySQL 示例

```golang
package main

import (
  "github.com/weijinwen/jupiter/pkg/store/gorm"
)

type Product struct {
  gorm.Model
  Code  string
  Price uint
}

func main() {
  // MySQL 配置
  db := gorm.StdNew("mysql_example")

  // 迁移 schema
  db.AutoMigrate(&Product{})

  // Create
  db.Create(&Product{Code: "D42", Price: 100})

  // Read
  var product Product
  db.First(&product, 1) // 根据整形主键查找
  db.First(&product, "code = ?", "D42") // 查找 code 字段值为 D42 的记录

  // Update - 将 product 的 price 更新为 200
  db.Model(&product).Update("Price", 200)
  // Update - 更新多个字段
  db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
  db.Model(&product).Updates(map[string]any{"Price": 200, "Code": "F42"})

  // Delete - 删除 product
  db.Delete(&product, 1)
}
```

### PostgreSQL 示例

```golang
package main

import (
  "github.com/weijinwen/jupiter/pkg/store/gorm"
)

type User struct {
  gorm.Model
  Name  string
  Email string
}

func main() {
  // PostgreSQL 配置
  db := gorm.StdNew("postgres_example")

  // 迁移 schema
  db.AutoMigrate(&User{})

  // Create
  db.Create(&User{Name: "John Doe", Email: "john@example.com"})

  // Read
  var user User
  db.First(&user, 1) // 根据整形主键查找
  db.First(&user, "email = ?", "john@example.com") // 查找 email 字段值为 john@example.com 的记录

  // Update - 将 user 的 name 更新为 Jane Doe
  db.Model(&user).Update("Name", "Jane Doe")
  // Update - 更新多个字段
  db.Model(&user).Updates(User{Name: "Jane Smith", Email: "jane@example.com"}) // 仅更新非零值字段
  db.Model(&user).Updates(map[string]any{"Name": "Jane Smith", "Email": "jane@example.com"})

  // Delete - 删除 user
  db.Delete(&user, 1)
}
```

## 配置说明

### MySQL 配置示例

```toml
[database.mysql_example]
dsn = "mysql://user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
debug = true
maxIdleConns = 10
maxOpenConns = 100
connMaxLifetime = "300s"
```

### PostgreSQL 配置示例

```toml
[database.postgres_example]
dsn = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
debug = true
maxIdleConns = 10
maxOpenConns = 100
connMaxLifetime = "300s"
```
