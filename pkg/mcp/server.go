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

package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/stones-hub/taurus-pro-http/pkg/router"
	httpServer "github.com/stones-hub/taurus-pro-http/pkg/server"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

type Transport string

var (
	TransportStdio          Transport = "stdio"           // 适合单机部署场景
	TransportSSE            Transport = "sse"             // 单机和集群都适合，但是在集群下需要维护有状态的session， Nginx同一个请求来源要路由到同一个服务上才可以
	TransportStreamableHTTP Transport = "streamable_http" // 适合集群部署场景
)

type Mode string

var (
	ModeStateful  Mode = "stateful"  // 保存上下文
	ModeStateless Mode = "stateless" // 不保存上下文
)

type MCPServer struct {
	Name       string             // 服务名称
	Version    string             // 服务版本
	Transport  Transport          // 传输协议
	Mode       Mode               // 模式
	server     *server.Server     // mcp server
	httpServer *httpServer.Server // http server
}

type McpServerOption func(*MCPServer)

func WithName(name string) McpServerOption {
	return func(s *MCPServer) {
		s.Name = name
	}
}

func WithVersion(version string) McpServerOption {
	return func(s *MCPServer) {
		s.Version = version
	}
}

func WithTransport(transport Transport) McpServerOption {
	return func(s *MCPServer) {
		s.Transport = transport
	}
}

func WithMode(mode Mode) McpServerOption {
	return func(s *MCPServer) {
		s.Mode = mode
	}
}

func WithHttpServer(httpServer *httpServer.Server) McpServerOption {
	return func(s *MCPServer) {
		s.httpServer = httpServer
	}
}

func New(options ...McpServerOption) (*MCPServer, func(), error) {
	// default options
	opts := &MCPServer{
		Name:       "mcp-server",
		Version:    "1.0.0",
		Transport:  TransportStreamableHTTP,
		Mode:       ModeStateless,
		httpServer: nil,
	}

	// apply options
	for _, option := range options {
		option(opts)
	}

	// check options to make sure the options are valid
	if opts.Transport != TransportStdio && opts.httpServer == nil {
		return nil, nil, fmt.Errorf("http server is required for non-stdio transport")
	}

	var stateMode transport.StateMode
	switch opts.Mode {
	case ModeStateful:
		stateMode = transport.Stateful
	case ModeStateless:
		stateMode = transport.Stateless
	default:
		stateMode = transport.Stateful
	}

	mcpTransport, mcpHandler := getTransport(opts.Transport, stateMode)

	mcpServer, err := server.NewServer(mcpTransport, server.WithServerInfo(protocol.Implementation{
		Name:    opts.Name,
		Version: opts.Version,
	}))

	if err != nil {
		log.Fatal(err)
	}

	opts.server = mcpServer

	switch h := mcpHandler.(type) {
	case nil:
		// stdio transport 不需要注册路由
	case *transport.SSEHandler:
		opts.httpServer.AddRouter(router.Router{
			Path:       "/sse",
			Handler:    h.HandleSSE(),
			Middleware: nil,
		})

		opts.httpServer.AddRouter(router.Router{
			Path:       "/message",
			Handler:    h.HandleMessage(),
			Middleware: nil,
		})
	case *transport.StreamableHTTPHandler:
		opts.httpServer.AddRouter(router.Router{
			Path:       "/mcp",
			Handler:    h.HandleMCP(),
			Middleware: nil,
		})
	default:
		log.Fatal(fmt.Errorf("unknown handler type: %T", mcpHandler))
	}

	return opts, func() {
		if err := opts.Shutdown(context.Background()); err != nil {
			log.Println(fmt.Errorf("failed to shutdown mcp server: %v", err))
		} else {
			log.Println("mcp server shutdown successfully !")
		}
	}, nil
}

func (s *MCPServer) Shutdown(ctx context.Context) error {

	return s.server.Shutdown(ctx)
}

// if you want to run stdio transport, you should run the server in the main thread
// for stdio transport, run the server in the main thread
func (s *MCPServer) Run() error {
	if s.Transport == TransportStdio {
		return s.server.Run()
	}
	return nil
}

func getTransport(transportName Transport, stateMode transport.StateMode) (transport.ServerTransport, interface{}) {
	var err error
	var t transport.ServerTransport
	var handler interface{}

	switch transportName {
	case TransportStdio:
		log.Println("start mcp server with stdio transport")
		t = transport.NewStdioServerTransport()
	case TransportSSE:
		log.Println("start mcp server with sse transport")
		var sseHandler *transport.SSEHandler
		t, sseHandler, err = transport.NewSSEServerTransportAndHandler("/message")
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create sse transport: %v", err))
		}
		handler = sseHandler
	case TransportStreamableHTTP:
		log.Println("start mcp server with streamable http transport")
		var streamableHandler *transport.StreamableHTTPHandler
		t, streamableHandler, err = transport.NewStreamableHTTPServerTransportAndHandler(transport.WithStreamableHTTPServerTransportAndHandlerOptionStateMode(stateMode))
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create streamable http transport: %v", err))
		}
		handler = streamableHandler
	default:
		log.Fatal(fmt.Errorf("unknown transport name: %s", transportName))
	}

	return t, handler
}

func (s *MCPServer) RegisterTool(tool *protocol.Tool, handler server.ToolHandlerFunc) {
	s.server.RegisterTool(tool, handler)
}

func (s *MCPServer) UnregisterTool(name string) {
	s.server.UnregisterTool(name)
}

func (s *MCPServer) RegisterPrompt(prompt *protocol.Prompt, handler server.PromptHandlerFunc) {
	s.server.RegisterPrompt(prompt, handler)
}

func (s *MCPServer) UnregisterPrompt(name string) {
	s.server.UnregisterPrompt(name)
}

func (s *MCPServer) RegisterResource(resource *protocol.Resource, handler server.ResourceHandlerFunc) {
	s.server.RegisterResource(resource, handler)
}

func (s *MCPServer) UnregisterResource(name string) {
	s.server.UnregisterResource(name)
}

func (s *MCPServer) RegisterResourceTemplate(resourceTemplate *protocol.ResourceTemplate, handler server.ResourceHandlerFunc) {
	s.server.RegisterResourceTemplate(resourceTemplate, handler)
}

func (s *MCPServer) UnregisterResourceTemplate(name string) {
	s.server.UnregisterResourceTemplate(name)
}
