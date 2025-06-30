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
	"log"
	"net/http"
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/common"
	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

// Config HTTP server config
type Config struct {
	Addr         string        // server address, default ":8080"
	ReadTimeout  time.Duration // read timeout, default 15s
	WriteTimeout time.Duration // write timeout, default 15s
	IdleTimeout  time.Duration // idle timeout, default 30s
}

// DefaultConfig default config
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

// Server HTTP server
type Server struct {
	*http.Server
	config Config
	router *router.RouterManager
}

// NewServer create a new server instance
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

// New create a new server instance
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

// AddRouter add a single router
func (s *Server) AddRouter(route router.Router) {
	s.router.AddRouter(route)
}

// AddRouterGroup add a router group
func (s *Server) AddRouterGroup(group router.RouteGroup) {
	s.router.AddRouterGroup(group)
}

// Get Server config
func (s *Server) GetConfig() Config {
	return s.config
}

// Start start server
func (s *Server) Start(errChan chan error) {
	// load all routes
	s.Handler = s.router.LoadRoutes()

	// start server
	go func() {
		log.Printf("%s🔗 -> Server is running on %s %s \n", common.Green, s.config.Addr, common.Reset)
		// when server startup failed, write error to errChan.
		// But http.ErrServerClosed is not an error,,because it is expected when the server is closed.
		// ListenAndServe is a blocking call
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("%s🔗 -> Server start failed on %s %s \n", common.Red, s.config.Addr, common.Reset)
			errChan <- err
		}
	}()
}

// Shutdown shutdown server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("%s🔗 -> Server is shutting down on %s %s \n", common.Yellow, s.config.Addr, common.Reset)
	return s.Server.Shutdown(ctx)
}
