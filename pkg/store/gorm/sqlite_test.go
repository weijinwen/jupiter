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
	"bytes"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cfg "github.com/weijinwen/jupiter/pkg/conf"
	"gorm.io/gorm"
)

func TestDetectDSNType(t *testing.T) {
	tests := []struct {
		dsn  string
		want DSNType
	}{
		{"root:123456@tcp(127.0.0.1:3306)/mysql?timeout=20s", DSNTypeMySQL},
		{"tcp(127.0.0.1)/config", DSNTypeMySQL},
		{"postgres://user:pass@localhost:5432/db", DSNTypePostgres},
		{"postgresql://user@localhost:5432/db", DSNTypePostgres},
		{"file:data.db?cache=shared", DSNTypeSQLite},
		{"data.db", DSNTypeSQLite},
		{"data.SQLite", DSNTypeSQLite},
		{"data.sqlite3", DSNTypeSQLite},
		{"data.db3", DSNTypeSQLite},
		{":memory:", DSNTypeSQLite},
		{":memory:?cache=shared", DSNTypeSQLite},
		{"sqlite://data/app.db", DSNTypeSQLite},
		{"sqlite3://data/app.db", DSNTypeSQLite},
	}
	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			assert.Equal(t, tt.want, detectDSNType(tt.dsn))
		})
	}
}

func TestParseSQLiteDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		wantDB   string
		wantAddr string
		wantNet  string
		wantPar  map[string]string
	}{
		{
			name:     "file with params",
			dsn:      "file:data/app.db?cache=shared&mode=rwc",
			wantDB:   "data/app.db",
			wantAddr: "data/app.db",
			wantNet:  "file",
			wantPar:  map[string]string{"cache": "shared", "mode": "rwc"},
		},
		{
			name:     "plain path",
			dsn:      "data/app.db",
			wantDB:   "data/app.db",
			wantAddr: "data/app.db",
			wantNet:  "file",
			wantPar:  nil,
		},
		{
			name:     "memory",
			dsn:      ":memory:?cache=shared",
			wantDB:   ":memory:",
			wantAddr: ":memory:",
			wantNet:  "file",
			wantPar:  map[string]string{"cache": "shared"},
		},
		{
			name:     "sqlite scheme",
			dsn:      "sqlite://data/app.db?_pragma=busy_timeout(5000)",
			wantDB:   "data/app.db",
			wantAddr: "data/app.db",
			wantNet:  "file",
			wantPar:  map[string]string{"_pragma": "busy_timeout(5000)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDSN(tt.dsn)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDB, got.DBName)
			assert.Equal(t, tt.wantAddr, got.Addr)
			assert.Equal(t, tt.wantNet, got.Net)
			assert.Equal(t, tt.wantPar, got.Params)
		})
	}
}

func TestDriverDSN(t *testing.T) {
	// 非 SQLite 原样返回
	assert.Equal(t, "root:1@tcp(127.0.0.1:3306)/mysql", driverDSN("root:1@tcp(127.0.0.1:3306)/mysql"))
	// SQLite 去掉自定义 scheme
	assert.Equal(t, "data/app.db?cache=shared", driverDSN("sqlite://data/app.db?cache=shared"))
	assert.Equal(t, "data/app.db", driverDSN("data/app.db"))
	assert.Equal(t, "file:data.db?cache=shared", driverDSN("file:data.db?cache=shared"))
}

func TestOpenSQLiteMemory(t *testing.T) {
	db, err := open(&Config{
		Name:            "sqlite-memory",
		DSN:             ":memory:",
		MaxIdleConns:    1,
		MaxOpenConns:    1,
		ConnMaxLifetime: 0,
		SlowThreshold:   0,
		Retry:           0,
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	type TestModel struct {
		gorm.Model
		Name string
	}

	require.NoError(t, db.AutoMigrate(&TestModel{}))
	require.NoError(t, db.Create(&TestModel{Name: "jupiter"}).Error)

	var got TestModel
	require.NoError(t, db.First(&got).Error)
	assert.Equal(t, "jupiter", got.Name)

	var version string
	require.NoError(t, db.Raw("select sqlite_version()").Scan(&version).Error)
	assert.NotEmpty(t, version)
}

func TestOpenSQLiteFile(t *testing.T) {
	dsn := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "jupiter.db")) + "?_pragma=busy_timeout(5000)"

	db, err := open(&Config{
		Name:            "sqlite-file",
		DSN:             dsn,
		MaxIdleConns:    2,
		MaxOpenConns:    2,
		ConnMaxLifetime: 0,
		SlowThreshold:   0,
		Retry:           1,
	})
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	type kv struct {
		ID    uint `gorm:"primaryKey"`
		Value string
	}
	require.NoError(t, db.Table("kv").AutoMigrate(&kv{}))
	require.NoError(t, db.Table("kv").Create(&kv{Value: "v1"}).Error)

	var count int64
	require.NoError(t, db.Table("kv").Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

// TestStdConfigSQLite 通过标准配置加载 SQLite 并构建连接
func TestStdConfigSQLite(t *testing.T) {
	config := `
	[jupiter.database.sqlite]
		dsn="file::memory:?cache=shared"
		debug=false
		maxIdleConns=1
		maxOpenConns=1
		level="panic"
		retry=0
	`
	assert.Nil(t, cfg.LoadFromReader(bytes.NewReader([]byte(config)), toml.Unmarshal))
	defer cfg.Reset()

	assert.NotPanics(t, func() {
		db := StdConfig("sqlite").MustBuild()
		assert.NotNil(t, db)
	})
}
