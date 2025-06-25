# Taurus Pro HTTP

Taurus Pro HTTP 是 Taurus 框架的 HTTP 组件，提供了强大的 HTTP 服务功能。

## 特性

- 标准化的请求/响应处理
- 灵活的路由管理
- 中间件支持
- 文件上传/下载
- JSON/XML/HTML 响应
- Range 请求支持

## 安装

```bash
go get github.com/stones-hub/taurus-pro-http
```

## 快速开始

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

## 使用示例

### 1. 处理 JSON 请求

```go
func handleJSON(w http.ResponseWriter, r *http.Request) {
    data, err := httpx.ParseJson(r)
    if err != nil {
        httpx.SendResponse(w, httpx.StatusInvalidRequest, nil, nil)
        return
    }
    httpx.SendResponse(w, http.StatusOK, data, nil)
}
```

### 2. 文件上传

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    files, err := httpx.ParseMultipartFile(r, "file")
    if err != nil {
        httpx.SendResponse(w, httpx.StatusInvalidRequest, nil, nil)
        return
    }
    
    if err := httpx.SaveUploadFiles(files, "./uploads"); err != nil {
        httpx.SendResponse(w, http.StatusInternalServerError, nil, nil)
        return
    }
    
    httpx.SendResponse(w, http.StatusOK, "Files uploaded successfully", nil)
}
```

### 3. 使用中间件

```go
func LoggerMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// 在路由中使用
router.AddRouter(router.Router{
    Path: "/api",
    Handler: handler,
    Middleware: []router.MiddlewareFunc{LoggerMiddleware},
})
```

## 许可证

Apache-2.0 license
