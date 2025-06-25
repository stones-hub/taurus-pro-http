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
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/common"
	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	IPCapacity     int           // 每个IP的令牌桶容量
	GlobalCapacity int           // 全局令牌桶容量
	FillInterval   time.Duration // 填充令牌的时间间隔
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = RateLimitConfig{
	IPCapacity:     60,          // 每个IP每分钟60个请求
	GlobalCapacity: 1000,        // 全局每分钟1000个请求
	FillInterval:   time.Minute, // 每分钟填充一次令牌
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(config *RateLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	// 创建组合限流器
	limiter := common.NewCompositeRateLimiter(
		config.IPCapacity,
		config.GlobalCapacity,
		config.FillInterval,
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取客户端IP
			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = forwardedFor
			}

			// 检查是否允许请求
			allowed, message := limiter.Allow(ip)
			if !allowed {
				if message == "" {
					message = "Too many requests"
				}
				httpx.SendResponse(w, http.StatusTooManyRequests, message, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
