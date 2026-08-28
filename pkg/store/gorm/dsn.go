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
	"errors"
	"net/url"
	"strings"
)

// DSN ...
type DSN struct {
	User     string            // Username
	Password string            // Password (requires User)
	Net      string            // Network type
	Addr     string            // Network address (requires Net)
	DBName   string            // Database name
	Params   map[string]string // Connection parameters
}

var (
	errInvalidDSNUnescaped = errors.New("invalid DSN: did you forget to escape a param value")
	errInvalidDSNAddr      = errors.New("invalid DSN: network address not terminated (missing closing brace)")
	errInvalidDSNNoSlash   = errors.New("invalid DSN: missing the slash separating the database name")
)

// parseDSN parses the DSN string to a NodeConfig.
func parseDSN(dsn string) (cfg *DSN, err error) {
	// 检查是否是 PostgreSQL 格式
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return parsePostgresDSN(dsn)
	}
	
	// 默认使用 MySQL 格式解析
	return parseMySQLDSN(dsn)
}

// parseMySQLDSN parses the MySQL DSN string to a NodeConfig.
func parseMySQLDSN(dsn string) (cfg *DSN, err error) {
	// New config with some default values
	cfg = new(DSN)

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			foundSlash = true
			var j, k int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								cfg.Password = dsn[k+1 : j]
								break
							}
						}
						cfg.User = dsn[:k]

						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return nil, errInvalidDSNUnescaped
							}
							return nil, errInvalidDSNAddr
						}
						cfg.Addr = dsn[k+1 : i-1]
						break
					}
				}
				cfg.Net = dsn[j+1 : k]
			}

			// dbname[?param1=value1&...&paramN=valueN]
			// Find the first '?' in dsn[i+1:]
			for j = i + 1; j < len(dsn); j++ {
				if dsn[j] == '?' {
					if err = parseDSNParams(cfg, dsn[j+1:]); err != nil {
						return
					}
					break
				}
			}
			cfg.DBName = dsn[i+1 : j]

			break
		}
	}
	if !foundSlash && len(dsn) > 0 {
		return nil, errInvalidDSNNoSlash
	}
	return
}

func parseDSNParams(cfg *DSN, params string) (err error) {
	for _, v := range strings.Split(params, "&") {
		param := strings.SplitN(v, "=", 2)
		if len(param) != 2 {
			continue
		}
		// lazy init
		if cfg.Params == nil {
			cfg.Params = make(map[string]string)
		}
		value := param[1]
		if cfg.Params[param[0]], err = url.QueryUnescape(value); err != nil {
			return
		}
	}
	return
}

// parsePostgresDSN parses the PostgreSQL DSN string to a NodeConfig.
func parsePostgresDSN(dsn string) (cfg *DSN, err error) {
	// New config with some default values
	cfg = new(DSN)
	cfg.Params = make(map[string]string)
	
	// 移除 postgres:// 或 postgresql:// 前缀
	var prefix string
	if strings.HasPrefix(dsn, "postgres://") {
		prefix = "postgres://"
	} else {
		prefix = "postgresql://"
	}
	
	dsn = strings.TrimPrefix(dsn, prefix)
	
	// 分离认证信息和地址/数据库名
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) == 2 {
		// 有认证信息
		authPart := parts[0]
		addrDbPart := parts[1]
		
		// 解析用户名和密码
		authParts := strings.SplitN(authPart, ":", 2)
		cfg.User = authParts[0]
		if len(authParts) == 2 {
			cfg.Password = authParts[1]
		}
		
		// 解析地址和数据库名
		addrDbParts := strings.SplitN(addrDbPart, "/", 2)
		if len(addrDbParts) >= 1 {
			cfg.Addr = addrDbParts[0]
		}
		if len(addrDbParts) == 2 {
			// 分离数据库名和参数
			dbParams := strings.SplitN(addrDbParts[1], "?", 2)
			cfg.DBName = dbParams[0]
			if len(dbParams) == 2 {
				if err = parseDSNParams(cfg, dbParams[1]); err != nil {
					return
				}
			}
		}
	} else {
		// 没有认证信息，直接解析地址和数据库名
		addrDbParts := strings.SplitN(dsn, "/", 2)
		if len(addrDbParts) >= 1 {
			cfg.Addr = addrDbParts[0]
		}
		if len(addrDbParts) == 2 {
			// 分离数据库名和参数
			dbParams := strings.SplitN(addrDbParts[1], "?", 2)
			cfg.DBName = dbParams[0]
			if len(dbParams) == 2 {
				if err = parseDSNParams(cfg, dbParams[1]); err != nil {
					return
				}
			}
		}
	}
	
	// PostgreSQL 默认使用 tcp 网络
	cfg.Net = "tcp"
	
	return
}
