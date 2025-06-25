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
