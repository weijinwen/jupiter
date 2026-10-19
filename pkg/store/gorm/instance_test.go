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
	"fmt"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	cfg "github.com/weijinwen/jupiter/pkg/conf"
)

type User struct {
	Birthday time.Time
	Age      int
	Name     string `gorm:"size:255"` // string默认长度为255, 使用这种tag重设。
}

func (u User) TableName() string {
	return "users"
}

func TestStdNew(t *testing.T) {
	type args struct {
		name string
		opts []any
	}
	tests := []struct {
		name    string
		args    args
		config  string
		wantErr bool
	}{
		// 注意：StdConfig 读取的键是 {prefix}.database.{name}，即 jupiter.database.demo。
		{
			name: "std new",
			args: args{
				name: "demo",
				opts: []any{},
			},
			wantErr: false,
			config: fmt.Sprintf(`
			[jupiter.database.demo]
				dsn=%q
				debug=true
				maxIdleConns=50
				connMaxLifeTime="20m"
				level="panic"
				slowThreshold="400ms"
				dialTimeout="1s"
			`, testDSN(t)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, cfg.LoadFromReader(bytes.NewReader([]byte(tt.config)), toml.Unmarshal))
			defer cfg.Reset()
			if tt.wantErr {
				assert.Panics(t, func() { StdConfig(tt.args.name).MustBuild() })
			} else {
				assert.NotPanics(t, func() { closeDB(t, StdConfig(tt.args.name).MustBuild()) })
			}
		})

		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, cfg.LoadFromReader(bytes.NewReader([]byte(tt.config)), toml.Unmarshal))
			defer cfg.Reset()
			assert.NotPanics(t, func() { closeDB(t, StdConfig(tt.args.name).MustBuild()) })
		})
	}
}

// closeDB 关闭底层连接，避免 SQLite 临时文件被占用导致 t.TempDir() 清理失败。
func closeDB(t *testing.T, db *DB) {
	t.Helper()

	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
}
