package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/common"
	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
	"github.com/stones-hub/taurus-pro-http/pkg/middleware"
	"github.com/stones-hub/taurus-pro-http/pkg/router"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

// User 表示用户模型
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// UserHandler 处理用户相关的请求
type UserHandler struct {
	users map[uint]User
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	// 模拟用户数据
	users := map[uint]User{
		1: {ID: 1, Username: "admin", Role: "admin"},
		2: {ID: 2, Username: "user1", Role: "user"},
		3: {ID: 3, Username: "user2", Role: "user"},
	}
	return &UserHandler{users: users}
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]User, 0, len(h.users))
	for _, user := range h.users {
		users = append(users, user)
	}

	httpx.SendResponse(w, http.StatusOK, users, nil)
}

// GetUser 获取单个用户
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// 从请求中获取用户ID（实际应该从URL参数获取）
	userID := uint(1)

	user, ok := h.users[userID]
	if !ok {
		httpx.SendResponse(w, httpx.StatusInvalidRequest, nil, nil)
		return
	}

	httpx.SendResponse(w, http.StatusOK, user, nil)
}

// LoginHandler 处理登录请求
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟登录成功
	userID := uint(1)
	username := "admin"

	// 生成 token
	token, err := common.GenerateToken(userID, username)
	if err != nil {
		httpx.SendResponse(w, http.StatusInternalServerError, nil, nil)
		return
	}

	// 存储 token 到 Redis（这里只是示例，实际需要 Redis 支持）
	ua := r.Header.Get("User-Agent")
	fmt.Printf("Store token to Redis: user_id=%d, ua=%s, token=%s\n", userID, ua, token)

	httpx.SendResponse(w, http.StatusOK, map[string]string{"token": token}, nil)
}

func main() {
	// 创建服务器实例
	srv := server.NewServer(
		server.WithAddr(":8080"),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(15*time.Second),
		server.WithIdleTimeout(30*time.Second),
	)

	// 创建处理器
	userHandler := NewUserHandler()

	// 创建中间件
	// jwtMiddleware := middleware.JWTMiddleware(nil)             // 使用默认配置
	rateLimitMiddleware := middleware.RateLimitMiddleware(nil) // 使用默认配置
	corsMiddleware := middleware.CorsMiddleware(nil)           // 使用默认配置

	// 公开路由组
	srv.AddRouterGroup(router.RouteGroup{
		Prefix: "/api/v1/public",
		Middleware: []router.MiddlewareFunc{
			corsMiddleware,
			rateLimitMiddleware,
		},
		Routes: []router.Router{
			{
				Path:    "/login",
				Handler: http.HandlerFunc(userHandler.LoginHandler),
			},
		},
	})

	// 受保护的路由组
	srv.AddRouterGroup(router.RouteGroup{
		Prefix: "/api/v1",
		Middleware: []router.MiddlewareFunc{
			corsMiddleware,
			rateLimitMiddleware,
			// jwtMiddleware,
		},
		Routes: []router.Router{
			{
				Path:    "/users",
				Handler: http.HandlerFunc(userHandler.GetUsers),
			},
			{
				Path:    "/users/" + strconv.Itoa(1),
				Handler: http.HandlerFunc(userHandler.GetUser),
			},
		},
	})

	srv.AddRouter(router.Router{
		Path: "/home",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		}),
	})

	// 启动服务器
	errChan := make(chan error, 1)
	srv.Start(errChan)

	// wait for server start failed or timeout
	if err := <-errChan; err != nil {
		log.Printf("Server start failed (%s) on %s \n", err.Error(), srv.GetConfig().Addr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
