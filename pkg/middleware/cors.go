// Copyright (c) 2025 Taurus Team. All rights reserved.
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

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package middleware

import (
	"net/http"
)

// CorsConfig CORS 配置
type CorsConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials bool
	MaxAge           string
}

// DefaultCorsConfig 默认的 CORS 配置
var DefaultCorsConfig = CorsConfig{
	AllowOrigins:     "*",
	AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	AllowHeaders:     "Content-Type, Authorization",
	AllowCredentials: false,
	MaxAge:           "86400",
}

// CorsMiddleware 添加 CORS 头到响应中
func CorsMiddleware(config *CorsConfig) func(http.Handler) http.Handler {
	// 如果没有提供配置，使用默认配置
	if config == nil {
		config = &DefaultCorsConfig
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置 CORS 头
			w.Header().Set("Access-Control-Allow-Origin", config.AllowOrigins)
			w.Header().Set("Access-Control-Allow-Methods", config.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", config.AllowHeaders)
			w.Header().Set("Access-Control-Max-Age", config.MaxAge)

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 处理预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
