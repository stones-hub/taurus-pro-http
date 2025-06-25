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
	"context"
	"net/http"

	"github.com/stones-hub/taurus-pro-http/pkg/common"
	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// TokenStore 定义了 token 存储接口
type TokenStore interface {
	// Store 存储 token，返回错误（如果有）
	Store(ctx context.Context, claims *common.Claims, token string, device string) error
	// Validate 验证 token 是否有效，返回是否有效和错误（如果有）
	Validate(ctx context.Context, claims *common.Claims, token string, device string) (bool, error)
}

type JWTContextKey string

// JWTConfig JWT 配置
type JWTConfig struct {
	TokenHeader   string        // token 在 header 中的键名，默认 "token"
	TokenStore    TokenStore    // token 存储实现
	JWTContextKey JWTContextKey // 上下文键名，默认 "jwt_claims"
}

// DefaultJWTConfig 默认的 JWT 配置
var DefaultJWTConfig = JWTConfig{
	TokenHeader:   "token",
	JWTContextKey: "jwt_claims",
}

// JWTMiddleware 创建 JWT 中间件
func JWTMiddleware(config *JWTConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &DefaultJWTConfig
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从请求中获取 token
			token := r.Header.Get(config.TokenHeader)
			if token == "" {
				httpx.SendResponse(w, http.StatusUnauthorized, "JWT Token is empty", nil)
				return
			}

			// 解析 token
			claims, err := common.ParseToken(token)
			if err != nil {
				httpx.SendResponse(w, http.StatusUnauthorized, "JWT Token parse error", nil)
				return
			}

			// 如果配置了 token store，验证 token 是否是最新的
			if config.TokenStore != nil {
				device := r.Header.Get("User-Agent")
				valid, err := config.TokenStore.Validate(r.Context(), claims, token, device)
				if err != nil {
					httpx.SendResponse(w, http.StatusUnauthorized, "Token validation error", nil)
					return
				}
				if !valid {
					httpx.SendResponse(w, http.StatusUnauthorized, "Token is not the latest", nil)
					return
				}
			}

			// 将 claims 信息存储到请求上下文中
			ctx := context.WithValue(r.Context(), config.JWTContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// StoreToken 存储 token（如果配置了 TokenStore）
func StoreToken(ctx context.Context, config *JWTConfig, claims *common.Claims, token string, device string) error {
	if config == nil || config.TokenStore == nil {
		return nil
	}
	return config.TokenStore.Store(ctx, claims, token, device)
}
