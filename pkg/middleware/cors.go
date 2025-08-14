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
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// CorsConfig CORS 配置
type CorsConfig struct {
	// AllowOrigins 支持多个域名，用逗号分隔，如："http://localhost:8080,https://example.com"
	// 如果设置为"*"且AllowCredentials为false时允许所有域名
	// 如果设置为具体域名，则只允许列表中的域名访问
	AllowOrigins string
	AllowMethods string
	AllowHeaders string
	// 是否允许携带凭证（cookies, HTTP认证及客户端SSL证书等）
	// 当设置为true时，AllowOrigins不能为"*"，必须指定具体域名
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

// validateOrigin 验证 Origin 是否合法
func validateOrigin(origin string) bool {
	if origin == "*" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// Origin 必须是 http 或 https
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	// Host 必须存在
	if u.Host == "" {
		return false
	}
	// 根据 RFC 6454，Origin 头只能包含 scheme://host[:port]
	if u.RawQuery != "" || u.Fragment != "" || (u.Path != "/" && u.Path != "") {
		return false
	}
	return true
}

// validateConfig 验证 CORS 配置是否合法
// 有效的 HTTP 方法列表
var validMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
	http.MethodPatch:   true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

func validateConfig(config *CorsConfig) error {
	// 验证 AllowOrigins
	if config.AllowCredentials && config.AllowOrigins == "*" {
		return fmt.Errorf("不能同时设置 AllowCredentials=true 和 AllowOrigins='*'")
	}

	// 验证配置的 AllowOrigins 是否合法
	if config.AllowOrigins != "*" {
		origins := strings.Split(config.AllowOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin == "" {
				return fmt.Errorf("origin 不能为空")
			}
			if !validateOrigin(origin) {
				return fmt.Errorf("无效的 Origin: %s", origin)
			}
		}
	}

	// 验证配置的 AllowMethods 是否合法
	if config.AllowMethods != "" && strings.TrimSpace(config.AllowMethods) != "*" {
		methods := strings.Split(config.AllowMethods, ",")
		for _, method := range methods {
			method = strings.TrimSpace(strings.ToUpper(method))
			if !validMethods[method] {
				return fmt.Errorf("无效的 HTTP 方法: %s", method)
			}
		}
	}

	// 验证配置的 MaxAge 是否合法
	if config.MaxAge != "" {
		maxAge, err := strconv.Atoi(config.MaxAge)
		if err != nil || maxAge < 0 {
			return fmt.Errorf("MaxAge 必须是非负整数")
		}
	}

	// 验证配置的 AllowHeaders 是否合法
	if config.AllowHeaders != "" && strings.TrimSpace(config.AllowHeaders) != "*" {
		headers := strings.Split(config.AllowHeaders, ",")
		for _, header := range headers {
			header = strings.TrimSpace(header)
			if header == "" {
				return fmt.Errorf("header 不能为空")
			}
			// 检查 header 是否是有效的 HTTP 头名称格式
			// HTTP 头名称只能包含字母、数字和连字符(-)
			for _, char := range header {
				isLetter := unicode.IsLetter(char) // 是否是字母
				isDigit := unicode.IsDigit(char)   // 是否是数字
				isHyphen := char == '-'            // 是否是连字符

				if !isLetter && !isDigit && !isHyphen {
					return fmt.Errorf("header 名称 '%s' 包含非法字符，只允许字母、数字和连字符(-)", header)
				}
			}
		}
	}

	return nil
}

// CorsMiddleware 添加 CORS 头到响应中
func CorsMiddleware(config *CorsConfig) func(http.Handler) http.Handler {
	// 如果没有提供配置，使用默认配置
	if config == nil {
		config = &DefaultCorsConfig
	}

	// 验证配置 config 是否合法
	if err := validateConfig(config); err != nil {
		panic(fmt.Sprintf("CORS 配置无效: %v", err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取请求的 Origin
			origin := r.Header.Get("Origin")
			if origin == "" {
				// 不是 CORS 请求，直接处理
				next.ServeHTTP(w, r)
				return
			}

			// 验证 origin 格式, 但是排除 origin 为空的情况
			if !validateOrigin(origin) {
				httpx.SendResponse(w, http.StatusForbidden, "Invalid Origin", nil)
				return
			}
			// 已验证完 origin ， 获取的 origin 是合法的且不为空， 接下来检查是否在允许的域名列表中

			// 1. 检查 Origin 是否允许
			if config.AllowOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// 检查是否在允许的域名列表中
				allowedOrigins := strings.Split(config.AllowOrigins, ",")
				for _, allowedOrigin := range allowedOrigins {
					if strings.TrimSpace(allowedOrigin) == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
						break
					}
				}
			}
			// 如果允许跨域，那已经设置 Access-Control-Allow-Origin 头， 接下来检查是否允许该 Origin
			// 如果没有设置 Access-Control-Allow-Origin 头，说明不允许该 Origin
			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				httpx.SendResponse(w, http.StatusForbidden, "Forbidden", nil)
				return
			}

			// ----------------------------------------------验证请求方法和请求头(开始)--------------------------------------------------
			// 2. 统一验证请求方法和请求头（不管是否是预检请求）
			// 2.1 验证请求方法, 强制如果是跨域，预检请求必须携带 Access-Control-Request-Method 头
			var requestMethod string
			if r.Method == http.MethodOptions {
				// 预检请求从头部获取方法
				requestMethod = r.Header.Get("Access-Control-Request-Method")

				// 如果预检请求没有携带 Access-Control-Request-Method 头，说明不是 CORS 请求，直接返回
				if requestMethod == "" {
					log.Println("[CORS] 预检请求没有携带 Access-Control-Request-Method 头，说明不是 CORS 请求，直接返回.")
					httpx.SendResponse(w, http.StatusForbidden, "Method not allowed", nil)
					return
				}

			} else {
				// 非预检请求直接使用请求方法
				requestMethod = r.Method
			}
			methodAllowed := false
			if strings.TrimSpace(config.AllowMethods) == "*" {
				methodAllowed = true
			} else {
				for _, method := range strings.Split(config.AllowMethods, ",") {
					if strings.TrimSpace(method) == requestMethod {
						methodAllowed = true
						break
					}
				}
			}
			if !methodAllowed {
				log.Printf("[CORS] 请求方法不允许, 请求方法: %s, 允许的方法: %s", requestMethod, config.AllowMethods)
				httpx.SendResponse(w, http.StatusForbidden, "Method not allowed", nil)
				return
			}

			// 2.2 验证请求头
			var requestHeaders string
			if r.Method == http.MethodOptions {
				// 预检请求从头部获取
				optionsHeaders := r.Header.Get("Access-Control-Request-Headers")
				optionsCustomHeaders := []string{}
				// 预检请求可能没有自定义头，这是正常的, 但是过滤掉标准头
				for _, headerName := range strings.Split(optionsHeaders, ",") {
					if !isStandardHeader(strings.TrimSpace(strings.ToLower(headerName))) {
						optionsCustomHeaders = append(optionsCustomHeaders, headerName)
					}
				}
				requestHeaders = strings.Join(optionsCustomHeaders, ",")
			} else {
				// 非预检请求只检查自定义头（非标准头）
				var customHeaders []string
				for headerName := range r.Header {
					if !isStandardHeader(strings.TrimSpace(strings.ToLower(headerName))) {
						customHeaders = append(customHeaders, headerName)
					}
				}
				requestHeaders = strings.Join(customHeaders, ",")
			}

			// 只有当有自定义头时才需要验证
			if requestHeaders != "" && strings.TrimSpace(config.AllowHeaders) != "*" {
				// 验证请求中使用的自定义头是否在允许列表中
				allowedHeadersMapLower := make(map[string]bool)
				allowedHeaders := strings.Split(config.AllowHeaders, ",")

				// 将允许的头转换为小写 map，便于查找
				for _, header := range allowedHeaders {
					allowedHeadersMapLower[strings.TrimSpace(strings.ToLower(header))] = true
				}

				// 检查请求中的每个自定义头是否在允许列表中
				headersAllowed := true
				for _, header := range strings.Split(requestHeaders, ",") {
					header = strings.TrimSpace(strings.ToLower(header))
					if header == "" {
						continue
					}

					if !allowedHeadersMapLower[header] {
						headersAllowed = false
						break
					}
				}

				if !headersAllowed {
					log.Printf("[CORS] 请求头验证失败, 请求头: %s, 允许的请求头: %s (注意：自定义头大小写不敏感)", requestHeaders, config.AllowHeaders)
					httpx.SendResponse(w, http.StatusForbidden, "Headers not allowed", nil)
					return
				}
			}

			// ----------------------------------------------验证请求方法和请求头(结束)--------------------------------------------------

			// 如果是预检请求，设置响应头并返回
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", config.AllowMethods)

				// 直接返回预检请求中声明的头（已经验证过了）
				optionsHeaders := r.Header.Get("Access-Control-Request-Headers")
				if optionsHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", optionsHeaders)
				}

				w.Header().Set("Access-Control-Max-Age", config.MaxAge)
				httpx.SendResponse(w, http.StatusNoContent, "No Content", nil)
				return
			}

			// 3. 对于非预检请求，设置必要的 CORS 响应头
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isStandardHeader 判断是否是标准 HTTP 头（不需要验证的头）
func isStandardHeader(headerName string) bool {
	headerNameLower := strings.ToLower(headerName)

	// HTTP/1.1 标准请求头（RFC 7231）
	standardHeaders := map[string]bool{
		// 通用头（General Headers）
		"cache-control":     true,
		"connection":        true,
		"date":              true,
		"pragma":            true,
		"trailer":           true,
		"transfer-encoding": true,
		"upgrade":           true,
		"via":               true,
		"warning":           true,

		// 请求头（Request Headers）
		"accept":              true,
		"accept-charset":      true,
		"accept-encoding":     true,
		"accept-language":     true,
		"authorization":       true,
		"expect":              true,
		"from":                true,
		"host":                true,
		"if-match":            true,
		"if-modified-since":   true,
		"if-none-match":       true,
		"if-range":            true,
		"if-unmodified-since": true,
		"max-forwards":        true,
		"proxy-authorization": true,
		"range":               true,
		"referer":             true,
		"te":                  true,
		"user-agent":          true,

		// 实体头（Entity Headers）
		"content-encoding": true,
		"content-language": true,
		"content-length":   true,
		"content-location": true,
		"content-md5":      true,
		"content-range":    true,
		"content-type":     true,

		// CORS 相关头
		"origin":                         true,
		"access-control-request-method":  true,
		"access-control-request-headers": true,

		// 其他常见标准头
		"cookie": true,
		"dnt":    true,
	}

	return standardHeaders[headerNameLower]
}
