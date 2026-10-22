// Copyright 2023 Douyu
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

package elasticsearch

import (
	"errors"
)

var (
	// ErrIndexNotFound 表示索引不存在
	ErrIndexNotFound = errors.New("elasticsearch index not found")

	// ErrDocumentNotFound 表示文档不存在
	ErrDocumentNotFound = errors.New("elasticsearch document not found")

	// ErrInvalidQuery 表示无效的查询条件
	ErrInvalidQuery = errors.New("invalid query condition")

	// ErrSlowQuery 表示查询过慢
	ErrSlowQuery = errors.New("elasticsearch slow query")

	// ErrMaxRetriesExceeded 表示超过最大重试次数
	ErrMaxRetriesExceeded = errors.New("elasticsearch max retries exceeded")
)
