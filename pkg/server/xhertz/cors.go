// Copyright 2026 Douyu
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

package xhertz

import (
	"net/http"
	"time"

	hzapp "github.com/cloudwego/hertz/pkg/app"
	hertzcors "github.com/hertz-contrib/cors"
)

// CORSConfig configures Cross-Origin Resource Sharing headers.
type CORSConfig = hertzcors.Config

func DefaultCORSConfig() CORSConfig {
	config := hertzcors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"AID",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = false
	config.MaxAge = 12 * time.Hour
	return config
}

func corsMiddleware(config CORSConfig) hzapp.HandlerFunc {
	return hertzcors.New(config)
}
