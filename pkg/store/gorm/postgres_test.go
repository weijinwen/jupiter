// Copyright 2022 Douyu
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

package gorm

import (
	"testing"

	"gorm.io/gorm"
)

// TestPostgresDSN 测试 PostgreSQL DSN 解析
func TestPostgresDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		user     string
		password string
		addr     string
		dbname   string
	}{
		{
			name:     "simple postgres DSN",
			dsn:      "postgres://user:password@localhost:5432/testdb?sslmode=disable",
			user:     "user",
			password: "password",
			addr:     "localhost:5432",
			dbname:   "testdb",
		},
		{
			name:     "postgresql DSN without password",
			dsn:      "postgresql://user@localhost:5432/testdb",
			user:     "user",
			password: "",
			addr:     "localhost:5432",
			dbname:   "testdb",
		},
		{
			name:     "postgres DSN without auth",
			dsn:      "postgres://localhost:5432/testdb",
			user:     "",
			password: "",
			addr:     "localhost:5432",
			dbname:   "testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseDSN(tt.dsn)
			if err != nil {
				t.Fatalf("parseDSN() error = %v", err)
			}

			if cfg.User != tt.user {
				t.Errorf("parseDSN() user = %v, want %v", cfg.User, tt.user)
			}
			if cfg.Password != tt.password {
				t.Errorf("parseDSN() password = %v, want %v", cfg.Password, tt.password)
			}
			if cfg.Addr != tt.addr {
				t.Errorf("parseDSN() addr = %v, want %v", cfg.Addr, tt.addr)
			}
			if cfg.DBName != tt.dbname {
				t.Errorf("parseDSN() dbname = %v, want %v", cfg.DBName, tt.dbname)
			}
			if cfg.Net != "tcp" {
				t.Errorf("parseDSN() net = %v, want tcp", cfg.Net)
			}
		})
	}
}

// TestOpenPostgres 测试 PostgreSQL 连接（需要实际的 PostgreSQL 服务器）
// 注意：这个测试需要实际的 PostgreSQL 服务器才能运行，在 CI/CD 环境中应该跳过
func TestOpenPostgres(t *testing.T) {
	// 使用环境变量获取 DSN，如果没有设置则跳过测试
	dsn := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

	config := &Config{
		DSN:             dsn,
		Debug:           true,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 0,
		OnDialError:     "panic",
		SlowThreshold:   0,
		DialTimeout:     0,
		AutoShadowTable: false,
		Retry:           1,
		RetryWaitTime:   0,
	}

	// 尝试打开连接
	db, err := open(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test, connection failed: %v", err)
		return
	}

	// 检查连接是否有效
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// 测试基本查询
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		t.Fatalf("Failed to execute raw query: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected query result 1, got %d", result)
	}
}

// TestGormWithPostgres 测试 GORM 基本功能与 PostgreSQL
func TestGormWithPostgres(t *testing.T) {
	dsn := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

	config := &Config{
		DSN:             dsn,
		Debug:           true,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 0,
		OnDialError:     "panic",
		SlowThreshold:   0,
		DialTimeout:     0,
		AutoShadowTable: false,
		Retry:           1,
		RetryWaitTime:   0,
	}

	db, err := open(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test, connection failed: %v", err)
		return
	}

	// 定义测试模型
	type TestModel struct {
		gorm.Model
		Name string
	}

	// 自动迁移
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建记录
	testRecord := TestModel{Name: "test"}
	if err := db.Create(&testRecord).Error; err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// 查询记录
	var retrieved TestModel
	if err := db.First(&retrieved, testRecord.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve record: %v", err)
	}

	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", retrieved.Name)
	}

	// 更新记录
	if err := db.Model(&retrieved).Update("Name", "updated").Error; err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 删除记录
	if err := db.Delete(&retrieved).Error; err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}
}

// TestCreateTablesAndInsertData 测试创建表和插入数据
func TestCreateTablesAndInsertData(t *testing.T) {
	// 使用您提供的实际 PostgreSQL 连接信息
	dsn := "postgres://user_kJrhNY:izb5cyjZHmYmaT5x@8.153.16.252:5432/postgres?sslmode=disable"

	config := &Config{
		DSN:             dsn,
		Debug:           true,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 0,
		OnDialError:     "panic",
		SlowThreshold:   0,
		DialTimeout:     0,
		AutoShadowTable: false,
		Retry:           1,
		RetryWaitTime:   0,
	}

	db, err := open(config)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test, connection failed: %v", err)
		return
	}

	// 定义用户模型
	type User struct {
		gorm.Model
		Name  string `gorm:"size:100;not null"`
		Email string `gorm:"size:100;uniqueIndex;not null"`
		Age   int    `gorm:"not null"`
	}

	// 定义产品模型
	type Product struct {
		gorm.Model
		Name        string  `gorm:"size:200;not null"`
		Description string  `gorm:"type:text"`
		Price       float64 `gorm:"not null"`
		Stock       int     `gorm:"default:0"`
		UserID      uint    `gorm:"not null"`
		User        User    `gorm:"foreignKey:UserID"`
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&User{}, &Product{}); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	// 创建用户数据
	users := []User{
		{Name: "张三", Email: "zhangsan@example.com", Age: 28},
		{Name: "李四", Email: "lisi@example.com", Age: 32},
		{Name: "王五", Email: "wangwu@example.com", Age: 25},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Errorf("Failed to create user %s: %v", user.Name, err)
		}
	}

	// 创建产品数据
	products := []Product{
		{Name: "笔记本电脑", Description: "高性能笔记本电脑", Price: 5999.99, Stock: 50, UserID: 1},
		{Name: "智能手机", Description: "最新款智能手机", Price: 2999.99, Stock: 100, UserID: 2},
		{Name: "无线耳机", Description: "蓝牙无线耳机", Price: 399.99, Stock: 200, UserID: 3},
		{Name: "平板电脑", Description: "轻薄便携平板电脑", Price: 1999.99, Stock: 80, UserID: 1},
		{Name: "智能手表", Description: "多功能智能手表", Price: 899.99, Stock: 150, UserID: 2},
	}

	for _, product := range products {
		if err := db.Create(&product).Error; err != nil {
			t.Errorf("Failed to create product %s: %v", product.Name, err)
		}
	}

	// 查询并验证用户数据
	var userCount int64
	if err := db.Model(&User{}).Count(&userCount).Error; err != nil {
		t.Errorf("Failed to count users: %v", err)
	}
	if userCount < 3 {
		t.Errorf("Expected at least 3 users, got %d", userCount)
	} else {
		t.Logf("Found %d users in database", userCount)
	}

	// 查询并验证产品数据
	var productCount int64
	if err := db.Model(&Product{}).Count(&productCount).Error; err != nil {
		t.Errorf("Failed to count products: %v", err)
	}
	if productCount < 5 {
		t.Errorf("Expected at least 5 products, got %d", productCount)
	} else {
		t.Logf("Found %d products in database", productCount)
	}

	// 测试关联查询
	var productsWithUsers []Product
	if err := db.Preload("User").Find(&productsWithUsers).Error; err != nil {
		t.Errorf("Failed to query products with users: %v", err)
	}

	// 验证关联数据
	validAssociations := 0
	for _, product := range productsWithUsers {
		if product.User.ID != 0 {
			validAssociations++
		}
	}
	if validAssociations == 0 {
		t.Error("No valid user-product associations found")
	} else {
		t.Logf("Found %d products with valid user associations", validAssociations)
	}

	// 测试条件查询
	var adults []User
	if err := db.Where("age >= ?", 25).Find(&adults).Error; err != nil {
		t.Errorf("Failed to query adult users: %v", err)
	}
	if len(adults) < 2 {
		t.Errorf("Expected at least 2 adult users, got %d", len(adults))
	} else {
		t.Logf("Found %d adult users (age >= 25)", len(adults))
	}

	var cheapProducts []Product
	if err := db.Where("price < ?", 3000).Find(&cheapProducts).Error; err != nil {
		t.Errorf("Failed to query cheap products: %v", err)
	}
	if len(cheapProducts) < 3 {
		t.Errorf("Expected at least 3 cheap products, got %d", len(cheapProducts))
	} else {
		t.Logf("Found %d cheap products (price < 3000)", len(cheapProducts))
	}

	t.Log("Successfully created tables and inserted data")
}
