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

package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

// Config HTTP服务器配置
type Config struct {
	Addr         string        // 服务器地址，默认 ":8080"
	ReadTimeout  time.Duration // 读取超时时间，默认 15s
	WriteTimeout time.Duration // 写入超时时间，默认 15s
	IdleTimeout  time.Duration // 空闲超时时间，默认 30s
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Addr:         ":8080",
	ReadTimeout:  15 * time.Second,
	WriteTimeout: 15 * time.Second,
	IdleTimeout:  30 * time.Second,
}

type serverOption func(*Server)

func WithAddr(addr string) serverOption {
	return func(s *Server) {
		s.config.Addr = addr
	}
}

func WithReadTimeout(readTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.ReadTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.WriteTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.IdleTimeout = idleTimeout
	}
}

// Server HTTP服务器
type Server struct {
	*http.Server
	config Config
	router *router.RouterManager
}

func NewServer(options ...serverOption) *Server {
	srv := &Server{
		config: DefaultConfig,
		router: router.NewRouterManager(),
	}

	for _, option := range options {
		option(srv)
	}

	// 初始化 http.Server
	srv.Server = &http.Server{
		Addr:         srv.config.Addr,
		ReadTimeout:  srv.config.ReadTimeout,
		WriteTimeout: srv.config.WriteTimeout,
		IdleTimeout:  srv.config.IdleTimeout,
	}

	return srv
}

// New 创建新的服务器实例
func New(config Config) *Server {
	// 使用默认配置填充未指定的值
	if config.Addr == "" {
		config.Addr = DefaultConfig.Addr
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = DefaultConfig.ReadTimeout
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = DefaultConfig.WriteTimeout
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = DefaultConfig.IdleTimeout
	}

	// 创建服务器实例
	srv := &Server{
		Server: &http.Server{
			Addr:         config.Addr,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		config: config,
		router: router.NewRouterManager(),
	}

	return srv
}

// AddRouter 添加单个路由
func (s *Server) AddRouter(route router.Router) {
	s.router.AddRouter(route)
}

// AddRouterGroup 添加路由组
func (s *Server) AddRouterGroup(group router.RouteGroup) {
	s.router.AddRouterGroup(group)
}

// Start 启动服务器
func (s *Server) Start() error {
	// 加载所有路由
	s.Handler = s.router.LoadRoutes()

	// 启动服务器
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
