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

package router

import (
	"log"
	"net/http"
)

// Router holds the configuration for a route, including its handler and middleware
type Router struct {
	Path       string
	Handler    http.Handler
	Middleware []MiddlewareFunc
}

// RouteGroup holds a group of routes with a common prefix and middleware
type RouteGroup struct {
	Prefix     string
	Middleware []MiddlewareFunc
	Routes     []Router
}

// RouterManager manages all routes and route groups
type RouterManager struct {
	routes          []Router
	routeGroups     []RouteGroup
	registeredPaths map[string]bool // Track registered paths
}

// NewRouterManager creates a new RouterManager
func NewRouterManager() *RouterManager {
	return &RouterManager{
		routes:          []Router{},
		routeGroups:     []RouteGroup{},
		registeredPaths: make(map[string]bool),
	}
}

// AddRouter adds a single route to the manager
func (rm *RouterManager) AddRouter(route Router) {
	rm.routes = append(rm.routes, route)
}

// AddRouterGroup adds a route group to the manager
func (rm *RouterManager) AddRouterGroup(group RouteGroup) {
	rm.routeGroups = append(rm.routeGroups, group)
}

// LoadRoutes loads all routes and route groups into a ServeMux
func (rm *RouterManager) LoadRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	// Load individual routes
	for _, route := range rm.routes {
		if rm.registeredPaths[route.Path] {
			log.Printf("Warning: Path %s is already registered, skipping.\n", route.Path)
			continue
		}
		handler := ChainMiddleware(route.Handler, route.Middleware...)
		mux.Handle(route.Path, handler)
		rm.registeredPaths[route.Path] = true
	}
	// Load route groups
	for _, group := range rm.routeGroups {
		for _, route := range group.Routes {
			// Combine group and route middleware, maintaining order
			allMiddleware := append(group.Middleware, route.Middleware...)
			handler := ChainMiddleware(route.Handler, allMiddleware...)
			// Ensure the path is correctly formatted
			fullPath := group.Prefix + route.Path
			if fullPath == "" || rm.registeredPaths[fullPath] {
				log.Printf("Warning: Path %s is already registered, skipping.\n", fullPath)
				continue // Skip if the full path is empty or already registered
			}
			mux.Handle(fullPath, handler)
			rm.registeredPaths[fullPath] = true
		}
	}
	return mux
}

// MiddlewareFunc defines a function to process middleware
type MiddlewareFunc func(http.Handler) http.Handler

// ChainMiddleware applies a list of middleware functions to an http.Handler
func ChainMiddleware(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	// 反转一下，调用顺序不对
	/*
	   for _, middleware := range middlewares {
	       handler = middleware(handler)
	   }
	*/
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

/*
Go 1.22+ 动态路由支持说明
========================

从 Go 1.22 开始，http.ServeMux 支持动态路径参数和 HTTP 方法匹配。

1. 动态路径参数语法
   - 使用 {paramName} 语法定义路径参数
   - 示例：/video/{userid}/get、/user/{id}/profile/{section}
   - 参数名区分大小写，建议使用小写字母和下划线

2. 路径参数获取
   - 在处理器中使用 r.PathValue("paramName") 获取参数值
   - 如果参数不存在，PathValue 返回空字符串
   - 建议使用 httpx.GetPathParam() 进行错误处理

3. HTTP 方法匹配
   - 支持在路由模式中指定 HTTP 方法
   - 语法：METHOD /path/pattern
   - 示例：GET /api/users/{id}、POST /api/users、PUT /api/users/{id}

4. 路由匹配优先级
   - 更具体的路径优先匹配
   - 例如：/users/{id} 比 /users/{id}/profile 更通用
   - 避免路径冲突，确保路由模式唯一性

5. 注意事项
   - 路径参数值不包含前导或尾随斜杠
   - 路径参数值已进行 URL 解码
   - 路径参数名不能包含特殊字符，只能使用字母、数字、下划线
   - 避免在路径参数中使用连字符，建议使用下划线

6. 使用示例
   ```go
   // 路由配置
   srv.AddRouter(router.Router{
       Path:    "/video/{userid}/get",
       Handler: http.HandlerFunc(videoHandler),
   })

   // 处理器中获取参数
   func videoHandler(w http.ResponseWriter, r *http.Request) {
       userid, err := httpx.GetPathParam(r, "userid")
       if err != nil {
           // 处理参数缺失错误
           return
       }
       // 使用 userid...
   }
   ```


*/
