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

package xretry

import (
	"math/rand"
	"time"
)

// 改为在需要使用随机数的地方创建一个新的随机数生成器实例
func Do(attempts int, sleep time.Duration, f func() error) error {
	// 创建随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if err := f(); err != nil {
		if attempts--; attempts >= 0 {
			// 使用局部随机数生成器
			jitter := time.Duration(r.Int63n(int64(sleep)))
			sleep = sleep + jitter/2
			time.Sleep(sleep)
			return Do(attempts, 2*sleep, f)
		}
		return err
	}
	return nil
}
