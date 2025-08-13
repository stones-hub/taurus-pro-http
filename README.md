# Taurus Pro HTTP

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/stones-hub/taurus-pro-http)](https://goreportcard.com/report/github.com/stones-hub/taurus-pro-http)

Taurus Pro HTTP 是一个高性能、企业级的 Go HTTP 服务框架，专为构建现代化的 Web 应用和 API 服务而设计。该框架提供了完整的 HTTP 服务解决方案，包括请求处理、响应管理、中间件支持、WebSocket 通信、JWT 认证、MCP 协议支持等核心功能。

## ✨ 核心特性

### 🚀 高性能 HTTP 服务
- **标准化请求/响应处理**：统一的 API 响应格式，支持 JSON、XML、HTML 等多种内容类型
- **灵活的路由管理**：支持路由组、中间件链式调用、动态路由参数
- **智能请求解析**：自动解析 GET/POST 参数、JSON 数据、文件上传等
- **Range 请求支持**：完整的 HTTP Range 请求处理，支持大文件分片下载

### 🔐 安全与认证
- **JWT 认证中间件**：完整的 JWT Token 验证和管理
- **Token 存储接口**：支持自定义 Token 存储策略，实现分布式会话管理
- **CORS 中间件**：灵活的跨域资源共享配置
- **速率限制**：内置请求频率限制，防止 API 滥用

### 🌐 WebSocket 支持
- **实时通信**：完整的 WebSocket 连接管理和消息处理
- **房间管理**：支持多房间、广播消息、连接池管理
- **自动重连**：智能的连接恢复和错误处理机制

### 🔌 MCP 协议支持
- **多传输协议**：支持 stdio、SSE、Streamable HTTP 等多种传输方式
- **状态管理**：支持有状态和无状态两种运行模式
- **集群部署**：专为分布式部署场景优化

### 🛠️ 开发体验
- **中间件系统**：可插拔的中间件架构，支持自定义中间件开发
- **错误处理**：统一的错误码管理和异常处理机制
- **日志追踪**：内置请求追踪和日志记录
- **性能监控**：运行时性能优化和资源管理

## 📦 安装

### 环境要求
- Go 1.23.0 或更高版本
- 支持的操作系统：Linux、macOS、Windows

### 安装命令
```bash
go get github.com/stones-hub/taurus-pro-http
```

### 依赖管理
```bash
go mod tidy
```

## 🚀 快速开始

### 基础 HTTP 服务

```go
package main

import (
    "log"
    "net/http"

    "github.com/stones-hub/taurus-pro-http/pkg/httpx"
    "github.com/stones-hub/taurus-pro-http/pkg/router"
)

func main() {
    // 创建路由组
    apiGroup := router.RouteGroup{
        Prefix: "/api",
        Routes: []router.Router{
            {
                Path: "/hello",
                Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                    httpx.SendResponse(w, http.StatusOK, "Hello World!", nil)
                }),
            },
        },
    }

    // 添加路由组
    router.AddRouterGroup(apiGroup)

    // 加载所有路由
    mux := router.LoadRoutes()

    // 启动服务器
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### 使用内置服务器

```go
package main

import (
    "log"
    "time"

    "github.com/stones-hub/taurus-pro-http/pkg/server"
)

func main() {
    // 创建服务器配置
    srv := server.New(server.Config{
        Addr:           ":8080",
        ReadTimeout:    60 * time.Second,
        WriteTimeout:   60 * time.Second,
        IdleTimeout:    300 * time.Second,
        MaxHeaderBytes: 1 << 20,
    })

    // 启动服务器
    errChan := make(chan error, 1)
    srv.Start(errChan)

    log.Println("🚀 服务器已启动，端口: 8080")

    // 等待服务器运行
    select {
    case err := <-errChan:
        log.Fatalf("服务器启动失败: %v", err)
    }
}
```

## 📚 使用指南

### 请求处理

#### 1. 解析请求参数

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 获取 GET 参数
    if value, err := httpx.GetParam(r, "id"); err == nil {
        log.Printf("ID: %s", value)
    }

    // 获取多个同名参数
    if values, err := httpx.GetParams(r, "tags"); err == nil {
        log.Printf("Tags: %v", values)
    }

    // 解析 JSON 请求体
    if data, err := httpx.ParseJson(r); err == nil {
        log.Printf("JSON data: %+v", data)
    }
}
```

#### 2. 文件上传处理

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    // 解析多部分表单文件
    files, err := httpx.ParseMultipartFile(r, "file")
    if err != nil {
        httpx.SendResponse(w, httpx.StatusInvalidRequest, nil, nil)
        return
    }
    
    // 保存上传的文件
    if err := httpx.SaveUploadFiles(files, "./uploads"); err != nil {
        httpx.SendResponse(w, http.StatusInternalServerError, nil, nil)
        return
    }
    
    httpx.SendResponse(w, http.StatusOK, "Files uploaded successfully", nil)
}
```

#### 3. 发送响应

```go
func handleResponse(w http.ResponseWriter, r *http.Request) {
    // 发送 JSON 响应
    httpx.SendResponse(w, http.StatusOK, map[string]interface{}{
        "message": "Success",
        "data":    []string{"item1", "item2"},
    }, nil)

    // 发送 XML 响应
    httpx.SendResponse(w, http.StatusOK, data, map[string]string{
        "Content-Type": "application/xml",
    })

    // 发送 HTML 响应
    httpx.SendResponse(w, http.StatusOK, "<h1>Hello World</h1>", map[string]string{
        "Content-Type": "text/html",
    })
}
```

### 中间件使用

#### 1. JWT 认证中间件

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/middleware"
)

func main() {
    // 配置 JWT 中间件
    jwtConfig := &middleware.JWTConfig{
        TokenHeader:   "Authorization",
        JWTContextKey: "user_claims",
    }

    // 在路由中使用
    router.AddRouter(router.Router{
        Path: "/api/protected",
        Handler: protectedHandler,
        Middleware: []router.MiddlewareFunc{
            middleware.JWTMiddleware(jwtConfig),
        },
    })
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    // 从上下文中获取用户信息
    claims := r.Context().Value("user_claims").(*common.Claims)
    httpx.SendResponse(w, http.StatusOK, claims, nil)
}
```

#### 2. 速率限制中间件

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/middleware"
)

func main() {
    // 配置速率限制
    rateLimitConfig := &middleware.RateLimitConfig{
        RequestsPerMinute: 100,
        BurstSize:         20,
    }

    router.AddRouter(router.Router{
        Path: "/api/limited",
        Handler: handler,
        Middleware: []router.MiddlewareFunc{
            middleware.RateLimitMiddleware(rateLimitConfig),
        },
    })
}
```

#### 3. CORS 中间件

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/middleware"
)

func main() {
    // 配置 CORS
    corsConfig := &middleware.CORSConfig{
        AllowedOrigins: []string{"http://localhost:3000", "https://example.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    }

    // 全局应用 CORS
    router.Use(middleware.CORSMiddleware(corsConfig))
}
```

### WebSocket 使用

#### 1. 基础 WebSocket 处理

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/wsocket"
)

func main() {
    // 初始化 WebSocket
    wsocket.Initialize()

    // 添加 WebSocket 路由
    router.AddRouter(router.Router{
        Path: "/ws",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            wsocket.HandleWebSocket(w, r, messageHandler)
        }),
    })
}

func messageHandler(conn *websocket.Conn, messageType int, message []byte) error {
    // 处理接收到的消息
    log.Printf("Received: %s", string(message))
    
    // 发送响应消息
    return conn.WriteMessage(messageType, []byte("Message received"))
}
```

#### 2. 房间管理

```go
func main() {
    // 创建 WebSocket Hub
    hub := wsocket.NewWebSocketHub()
    
    // 启动 Hub
    go hub.Run()

    router.AddRouter(router.Router{
        Path: "/ws/room",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            wsocket.HandleWebSocketRoom(w, r, messageHandler, hub, "default")
        }),
    })
}
```

### MCP 服务器

#### 1. 创建 MCP 服务器

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/mcp"
)

func main() {
    // 创建 MCP 服务器
    mcpServer, cleanup, err := mcp.New(
        mcp.WithName("my-mcp-server"),
        mcp.WithVersion("1.0.0"),
        mcp.WithTransport(mcp.TransportStreamableHTTP),
        mcp.WithMode(mcp.ModeStateless),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // 启动 MCP 服务器
    if err := mcpServer.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## 🏗️ 项目结构

```
taurus-pro-http/
├── bin/                    # 可执行文件
├── docs/                   # 文档
├── examples/               # 使用示例
├── pkg/                    # 核心包
│   ├── common/            # 通用功能
│   │   ├── jwt.go         # JWT 工具
│   │   └── rate_limiter.go # 速率限制器
│   ├── httpx/             # HTTP 扩展
│   │   ├── request.go      # 请求处理
│   │   ├── response.go     # 响应处理
│   │   └── wrapper/        # 包装器
│   ├── middleware/         # 中间件
│   │   ├── cors.go         # CORS 中间件
│   │   ├── jwt.go          # JWT 中间件
│   │   ├── rate_limit.go   # 速率限制中间件
│   │   └── recovery.go     # 恢复中间件
│   ├── router/             # 路由管理
│   │   └── router.go       # 路由核心
│   ├── server/             # 服务器
│   │   └── server.go       # HTTP 服务器
│   ├── wsocket/            # WebSocket
│   │   ├── broadcast.go     # 广播功能
│   │   ├── handler.go       # 处理器
│   │   └── websocket.go     # WebSocket 核心
│   └── mcp/                # MCP 协议
│       └── server.go       # MCP 服务器
├── scripts/                # 系统优化脚本
├── go.mod                  # Go 模块文件
├── go.sum                  # 依赖校验文件
└── README.md               # 项目说明
```

## 🔧 配置说明

### 服务器配置

```go
type Config struct {
    Addr           string        // 监听地址
    ReadTimeout    time.Duration // 读取超时
    WriteTimeout   time.Duration // 写入超时
    IdleTimeout    time.Duration // 空闲超时
    MaxHeaderBytes int           // 最大头部大小
}
```

### JWT 配置

```go
type JWTConfig struct {
    TokenHeader   string        // Token 头部键名
    TokenStore    TokenStore    // Token 存储实现
    JWTContextKey JWTContextKey // 上下文键名
}
```

### CORS 配置

```go
type CORSConfig struct {
    AllowedOrigins []string // 允许的源
    AllowedMethods []string // 允许的方法
    AllowedHeaders []string // 允许的头部
    ExposedHeaders []string // 暴露的头部
    AllowCredentials bool   // 允许凭证
    MaxAge           int    // 预检请求缓存时间
}
```

## 🚀 性能优化

### 系统优化脚本

项目提供了针对不同硬件配置的系统优化脚本：

- `scripts/optimize_system_8c16g.sh` - 8核16G系统优化
- `scripts/optimize_system_16c32g.sh` - 16核32G系统优化
- `scripts/optimize_system_safe.sh` - 安全优化配置

### Go 运行时优化

```go
func initGoRuntime() {
    // 设置CPU核心数
    runtime.GOMAXPROCS(8)
    
    // 设置GC参数
    debug.SetGCPercent(150)
    
    // 设置内存限制
    debug.SetMemoryLimit(12 * 1024 * 1024 * 1024)
}
```

## 📖 API 参考

### 状态码定义

```go
const (
    StatusInvalidRequest = 1001 // 无效请求
    StatusInvalidParams  = 1002 // 无效参数
    StatusUnauthorized   = 1003 // 未授权
)
```

### 响应格式

```go
type Response struct {
    Code    int         `json:"code"`    // 状态码
    Message string      `json:"message"` // 状态消息
    Data    interface{} `json:"data"`    // 响应数据
}
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/stones-hub/taurus-pro-http.git

# 进入项目目录
cd taurus-pro-http

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 构建项目
go build ./...
```

## 📄 许可证

本项目采用 Apache License 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系我们

- **作者**: yelei
- **邮箱**: 61647649@qq.com
- **项目地址**: https://github.com/stones-hub/taurus-pro-http

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和用户！

---

**Taurus Pro HTTP** - 构建高性能 HTTP 服务的首选框架 🚀
