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

	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

// Config HTTP server config
type Config struct {
	// Addr 服务器监听地址
	// 格式: "ip:port" 或 ":port"
	// 示例: ":8080" (监听所有网卡8080端口), "127.0.0.1:8080" (仅监听本地8080端口)
	// 影响: 决定服务器在哪个网络接口和端口上监听连接
	Addr string

	// ReadTimeout 读取超时时间
	// 作用: 控制从TCP连接建立到完整读取HTTP请求的最大时间
	// 触发条件: 客户端发送请求速度过慢、网络延迟高、请求体过大
	// 超时后果: 连接被强制关闭，返回408 Request Timeout
	// 配置建议:
	//   - 简单API: 15-30秒
	//   - 复杂业务: 30-60秒
	//   - 文件上传: 60-300秒
	ReadTimeout time.Duration

	// WriteTimeout 写入超时时间
	// 作用: 控制从开始写入HTTP响应到完成的最大时间
	// 触发条件: 响应数据过大、网络拥塞、客户端接收速度慢
	// 超时后果: 连接被强制关闭，客户端可能收到不完整响应
	// 配置建议:
	//   - 小响应: 15-30秒
	//   - 大响应: 30-60秒
	//   - 文件下载: 60-300秒
	WriteTimeout time.Duration

	// IdleTimeout 空闲超时时间 (HTTP Keepalive)
	// 作用: 控制HTTP keepalive连接的最大空闲时间
	// 工作原理: 客户端发送请求后，连接保持打开状态等待下一个请求
	// 超时后果: 空闲连接被关闭，下次请求需要重新建立连接
	// 性能影响:
	//   - 短超时: 连接复用率低，增加连接建立开销
	//   - 长超时: 连接复用率高，但占用更多资源
	// 配置建议:
	//   - 高并发API: 120-300秒
	//   - 普通Web服务: 60-120秒
	//   - 长连接服务: 300-600秒
	IdleTimeout time.Duration

	// MaxHeaderBytes 最大请求头大小
	// 作用: 限制HTTP请求头的最大字节数
	// 安全考虑: 防止恶意客户端发送超大请求头进行DoS攻击
	// 内存影响: 每个连接都会预分配此大小的缓冲区
	// 配置建议:
	//   - 标准Web: 1MB (1 << 20)
	//   - 需要大Cookie: 2-4MB
	//   - 安全要求高: 512KB-1MB
	MaxHeaderBytes int
}

// DefaultConfig 默认配置
// 这些配置针对8核16G+系统进行了优化，适合生产环境使用
// 配置特点:
//   - 30秒读写超时: 平衡性能和稳定性，适合大多数业务场景
//   - 3分钟keepalive: 提高连接复用率，减少连接建立开销
//   - 1MB请求头限制: 防止恶意攻击，内存使用合理
//
// 适用场景: 高并发API服务、Web应用、微服务等
var DefaultConfig = Config{
	Addr:           ":8080",           // 服务器监听地址，默认监听所有网卡的8080端口
	ReadTimeout:    30 * time.Second,  // 读取超时时间：从连接建立到读取完整个HTTP请求的最大时间
	WriteTimeout:   30 * time.Second,  // 写入超时时间：从开始写入HTTP响应到完成的最大时间
	IdleTimeout:    180 * time.Second, // 空闲超时时间：HTTP keepalive连接的最大空闲时间，影响连接复用效率
	MaxHeaderBytes: 1 << 20,           // 最大请求头大小：限制HTTP请求头的最大字节数，防止恶意大请求头攻击
}

// serverOption 服务器配置选项函数类型
// 用于在创建服务器时自定义配置参数
type serverOption func(*Server)

// WithAddr 设置服务器监听地址
// 参数: addr - 监听地址，格式为 "ip:port" 或 ":port"
// 示例: WithAddr(":9090"), WithAddr("127.0.0.1:8080")
// 用途: 自定义服务器监听的网络接口和端口
func WithAddr(addr string) serverOption {
	return func(s *Server) {
		s.config.Addr = addr
	}
}

// WithReadTimeout 设置读取超时时间
// 参数: readTimeout - 读取超时时间，建议根据业务复杂度调整
// 配置建议:
//   - 简单API: 15-30秒
//   - 复杂业务逻辑: 30-60秒
//   - 文件上传: 60-300秒
//
// 用途: 防止慢客户端或网络问题影响服务器性能
func WithReadTimeout(readTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout 设置写入超时时间
// 参数: writeTimeout - 写入超时时间，建议根据响应大小调整
// 配置建议:
//   - 小响应: 15-30秒
//   - 大响应: 30-60秒
//   - 文件下载: 60-300秒
//
// 用途: 防止网络拥塞或客户端问题导致响应写入超时
func WithWriteTimeout(writeTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.WriteTimeout = writeTimeout
	}
}

// WithIdleTimeout 设置空闲超时时间 (HTTP Keepalive)
// 参数: idleTimeout - 空闲超时时间，影响连接复用效率
// 配置建议:
//   - 高并发API: 120-300秒
//   - 普通Web服务: 60-120秒
//   - 长连接服务: 300-600秒
//
// 用途: 优化连接复用，减少连接建立/关闭开销
func WithIdleTimeout(idleTimeout time.Duration) serverOption {
	return func(s *Server) {
		s.config.IdleTimeout = idleTimeout
	}
}

// WithMaxHeaderBytes 设置最大请求头大小
// 参数: maxHeaderBytes - 请求头最大字节数，建议使用 1 << 20 (1MB)
// 安全考虑: 防止恶意大请求头攻击
// 内存影响: 每个连接都会预分配此大小的缓冲区
// 用途: 限制请求头大小，平衡安全性和功能性
func WithMaxHeaderBytes(maxHeaderBytes int) serverOption {
	return func(s *Server) {
		s.config.MaxHeaderBytes = maxHeaderBytes
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
		Addr:           srv.config.Addr,
		ReadTimeout:    srv.config.ReadTimeout,
		WriteTimeout:   srv.config.WriteTimeout,
		IdleTimeout:    srv.config.IdleTimeout,
		MaxHeaderBytes: srv.config.MaxHeaderBytes,
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
	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = DefaultConfig.MaxHeaderBytes
	}

	// 创建服务器实例
	srv := &Server{
		Server: &http.Server{
			Addr:           config.Addr,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			IdleTimeout:    config.IdleTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes,
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
		log.Printf("Server is running on %s \n", s.config.Addr)
		// when server startup failed, write error to errChan.
		// But http.ErrServerClosed is not an error,,because it is expected when the server is closed.
		// ListenAndServe is a blocking call
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server start failed on %s \n", s.config.Addr)
			errChan <- err
		}
	}()
}

// Shutdown shutdown server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("Server is shutting down on %s \n", s.config.Addr)
	return s.Server.Shutdown(ctx)
}

/*
并发性能优化建议：

1. 系统层面优化：
   - 增加文件描述符限制：ulimit -n 65535
   - 优化内核参数：net.core.somaxconn, net.ipv4.tcp_max_syn_backlog
   - 调整 TCP 参数：net.ipv4.tcp_fin_timeout, net.ipv4.tcp_tw_reuse

2. Go 运行时优化：
   - 设置 GOMAXPROCS 为 CPU 核心数
   - 调整 GC 参数：GOGC=100
   - 使用 pprof 分析性能瓶颈

3. 应用层面优化：
   - 使用连接池管理数据库连接
   - 避免在请求处理中使用全局锁
   - 使用 goroutine 池处理并发请求
   - 实现请求限流和熔断机制

4. 网络层面优化：
   - 启用 TCP 快速打开
   - 调整 TCP 缓冲区大小
   - 使用 SO_REUSEPORT 支持多进程监听

5. 监控和调优：
   - 监控 goroutine 数量
   - 监控内存分配和 GC 情况
   - 使用 netstat 查看连接状态
*/
