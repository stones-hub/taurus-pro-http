# HTTP 响应包装器

这个包提供了 HTTP 响应包装器，专门用于中间件场景中的响应拦截、修改和延迟发送。

## 功能特性

- 📤 **响应体拦截**: 收集响应数据到内存，支持修改
- 🏷️ **头部管理**: 动态修改响应头
- 📊 **状态码控制**: 精确控制 HTTP 状态码
- 📋 **延迟发送**: 在最终处理完成后统一发送响应
- 🔄 **便捷方法**: 提供 JSON、文本、错误响应的便捷方法

## 核心设计理念

**专注于响应拦截和延迟发送**，让中间件能够：
1. 拦截业务逻辑的响应
2. 修改响应数据、状态码、头部
3. 在合适的时机统一发送响应

## 快速开始

### 基本使用

```go
package main

import (
    "net/http"
    "github.com/your-project/pkg/httpx/wrapper"
)

func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 创建响应包装器
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用下一个处理器
        next.ServeHTTP(respWrapper, r)
        
        // 处理响应
        log.Printf("响应状态: %d", respWrapper.GetStatusCode())
        log.Printf("响应体: %s", respWrapper.GetBodyString())
        
        // 发送响应
        respWrapper.SendResponse()
    })
}
```

## 应用场景

### 1. 日志记录中间件

记录所有响应的详细信息，包括状态码、响应体、响应时间等。

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 记录响应信息
        duration := time.Since(start)
        log.Printf("响应: %d, 耗时: %v, 大小: %d bytes", 
            respWrapper.GetStatusCode(), 
            duration, 
            len(respWrapper.GetBody()))
        
        // 发送响应
        respWrapper.SendResponse()
    })
}
```

### 2. 数据加密中间件

对响应体进行加密处理，保护敏感数据。

```go
func EncryptionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 加密响应体
        if len(respWrapper.GetBody()) > 0 {
            encryptedBody := encrypt(respWrapper.GetBody()) // 你的加密函数
            respWrapper.SetBody(encryptedBody)
            
            // 设置加密标识头
            respWrapper.Header().Set("X-Encrypted", "true")
        }
        
        respWrapper.SendResponse()
    })
}
```

### 3. 响应数据转换中间件

统一处理响应数据格式，添加元数据或转换数据结构。

```go
func TransformMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 转换响应数据
        if respWrapper.GetStatusCode() == 200 {
            var response map[string]interface{}
            if err := json.Unmarshal(respWrapper.GetBody(), &response); err == nil {
                // 添加元数据
                response["_metadata"] = map[string]interface{}{
                    "server_time": time.Now().Unix(),
                    "request_id":  r.Header.Get("X-Request-ID"),
                    "version":     "1.0",
                }
                
                // 重新设置响应体
                if newBody, err := json.Marshal(response); err == nil {
                    respWrapper.SetBody(newBody)
                }
            }
        }
        
        respWrapper.SendResponse()
    })
}
```

### 4. 错误处理中间件

统一处理错误响应格式，确保错误信息的一致性。

```go
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 统一错误格式
        statusCode := respWrapper.GetStatusCode()
        if statusCode >= 400 {
            errorResponse := map[string]interface{}{
                "error":     getErrorMessage(statusCode),
                "code":      statusCode,
                "timestamp": time.Now().Unix(),
                "path":      r.URL.Path,
            }
            respWrapper.RespondWithJSON(statusCode, errorResponse)
        } else {
            respWrapper.SendResponse()
        }
    })
}

func getErrorMessage(statusCode int) string {
    switch statusCode {
    case 400: return "请求参数错误"
    case 401: return "未授权访问"
    case 403: return "禁止访问"
    case 404: return "资源不存在"
    case 500: return "服务器内部错误"
    default: return "未知错误"
    }
}
```

### 5. 响应压缩中间件

对响应体进行压缩处理，减少网络传输量。

```go
func CompressionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 检查是否支持压缩
        if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            originalBody := respWrapper.GetBody()
            if len(originalBody) > 1024 { // 只压缩大于1KB的响应
                compressedBody := compressGzip(originalBody) // 你的压缩函数
                respWrapper.SetBody(compressedBody)
                respWrapper.Header().Set("Content-Encoding", "gzip")
                respWrapper.Header().Set("Content-Length", strconv.Itoa(len(compressedBody)))
            }
        }
        
        respWrapper.SendResponse()
    })
}
```

### 6. 限流中间件

基于响应状态码进行限流控制。

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    clients := make(map[string]int)
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 获取客户端IP
        clientIP := getClientIP(r)
        
        // 检查限流
        if count, exists := clients[clientIP]; exists && count >= 100 {
            respWrapper.RespondWithError(429, errors.New("请求过于频繁"))
            return
        }
        
        // 调用业务逻辑
        next.ServeHTTP(respWrapper, r)
        
        // 根据响应状态码更新限流计数
        if respWrapper.GetStatusCode() == 200 {
            clients[clientIP]++
        }
        
        respWrapper.SendResponse()
    })
}
```

## 中间件链组合

```go
func SetupMiddlewareChain() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/example", ExampleHandler)
    
    // 按顺序应用中间件
    var handler http.Handler = mux
    handler = LoggingMiddleware(handler)        // 日志记录
    handler = ErrorHandlingMiddleware(handler)  // 错误处理
    handler = CompressionMiddleware(handler)    // 响应压缩
    handler = TransformMiddleware(handler)      // 数据转换
    handler = EncryptionMiddleware(handler)     // 数据加密
    
    return handler
}
```

## API 参考

### 核心方法

| 方法 | 描述 |
|------|------|
| `Write([]byte) (int, error)` | 写入响应体（拦截到内存） |
| `WriteHeader(int)` | 设置状态码 |
| `Header() Header` | 获取响应头 |
| `SendResponse()` | 发送完整响应 |

### 数据访问

| 方法 | 描述 |
|------|------|
| `GetBody() []byte` | 获取响应体字节数组 |
| `GetBodyString() string` | 获取响应体字符串 |
| `SetBody([]byte)` | 设置响应体 |
| `GetStatusCode() int` | 获取状态码 |

### 便捷方法

| 方法 | 描述 |
|------|------|
| `Respond(int, []byte)` | 发送响应（便捷方法） |
| `RespondWithJSON(int, interface{})` | 发送JSON响应 |
| `RespondWithText(int, string)` | 发送文本响应 |
| `RespondWithError(int, error)` | 发送错误响应 |
| `Reset()` | 重置包装器 |

## 注意事项

1. **内存使用**: 包装器会将响应体加载到内存中，对于大文件下载需要特别注意
2. **性能影响**: 包装器会带来一定的性能开销，建议在需要时使用
3. **错误处理**: 确保在中间件中正确处理错误，避免响应未发送
4. **并发安全**: 包装器本身不是并发安全的，每个请求应该使用独立的包装器实例
5. **必须调用 SendResponse()**: 包装器只是拦截响应，必须调用 `SendResponse()` 才会真正发送

## 许可证

MIT License
