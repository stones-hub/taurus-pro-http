# HTTP 请求和响应包装器

这个包提供了完整的HTTP请求和响应包装器，允许你在中间件中对请求和响应进行拦截、修改和处理。

## 功能特性

### RequestWrapper 请求包装器
- 🔍 **请求体拦截**: 支持多次读取请求体
- 📝 **参数处理**: 便捷的查询参数和表单参数访问
- 🏷️ **头部管理**: 动态修改请求头
- 🍪 **Cookie处理**: 获取和设置Cookie
- 🔄 **数据转换**: JSON序列化和反序列化
- 📋 **格式检测**: 自动识别JSON和表单请求
- 🆔 **请求克隆**: 支持请求的完整复制

### ResponseWrapper 响应包装器
- 📤 **响应体拦截**: 收集响应数据到内存
- 🏷️ **头部管理**: 动态修改响应头
- 📊 **状态码控制**: 精确控制HTTP状态码
- 🔄 **流式支持**: 支持Flush、Hijack等高级功能
- 📋 **延迟发送**: 在最终处理完成后统一发送响应

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
        // 创建包装器
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 处理请求
        log.Printf("收到请求: %s", reqWrapper.GetBodyString())
        
        // 调用下一个处理器
        next.ServeHTTP(respWrapper, reqWrapper.Request)
        
        // 处理响应
        log.Printf("响应状态: %d", respWrapper.statusCode)
        
        // 发送响应
        respWrapper.SendResponse()
    })
}
```

### 请求处理示例

```go
// 解析JSON请求体
var data map[string]interface{}
if err := reqWrapper.GetJSONBody(&data); err != nil {
    // 处理错误
}

// 获取查询参数
page := reqWrapper.GetQueryParam("page")
size := reqWrapper.GetQueryParam("size")

// 获取请求头
userAgent := reqWrapper.GetUserAgent()
apiKey := reqWrapper.GetHeader("X-API-Key")

// 修改请求体
newData := map[string]interface{}{
    "original": data,
    "timestamp": time.Now().Unix(),
}
reqWrapper.SetJSONBody(newData)
```

### 响应处理示例

```go
// 设置响应头
respWrapper.Header().Set("Content-Type", "application/json")
respWrapper.Header().Set("X-Response-Time", time.Now().Format(time.RFC3339))

// 写入响应体
respWrapper.Write([]byte(`{"message": "success"}`))

// 设置状态码
respWrapper.WriteHeader(http.StatusOK)

// 最终发送响应
respWrapper.SendResponse()
```

## 应用场景

### 1. 日志记录
```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 记录请求信息
        log.Printf("请求: %s %s", r.Method, r.URL.Path)
        log.Printf("请求体: %s", reqWrapper.GetBodyString())
        
        next.ServeHTTP(respWrapper, reqWrapper.Request)
        
        // 记录响应信息
        duration := time.Since(start)
        log.Printf("响应: %d, 耗时: %v", respWrapper.statusCode, duration)
        
        respWrapper.SendResponse()
    })
}
```

### 2. 数据加密/解密
```go
func EncryptionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 解密请求体
        if reqWrapper.IsJSON() {
            encryptedBody := reqWrapper.GetBody()
            decryptedBody := decrypt(encryptedBody) // 你的解密函数
            reqWrapper.SetBody(decryptedBody)
        }
        
        next.ServeHTTP(respWrapper, reqWrapper.Request)
        
        // 加密响应体
        if len(respWrapper.body) > 0 {
            encryptedResponse := encrypt(respWrapper.body) // 你的加密函数
            respWrapper.body = encryptedResponse
        }
        
        respWrapper.SendResponse()
    })
}
```

### 3. 请求验证
```go
func ValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 验证API密钥
        apiKey := reqWrapper.GetHeader("X-API-Key")
        if !isValidAPIKey(apiKey) {
            respWrapper.WriteHeader(http.StatusUnauthorized)
            respWrapper.Write([]byte(`{"error": "无效的API密钥"}`))
            respWrapper.SendResponse()
            return
        }
        
        // 验证签名
        signature := reqWrapper.GetHeader("X-Signature")
        if !verifySignature(reqWrapper.GetBody(), signature) {
            respWrapper.WriteHeader(http.StatusUnauthorized)
            respWrapper.Write([]byte(`{"error": "签名验证失败"}`))
            respWrapper.SendResponse()
            return
        }
        
        next.ServeHTTP(respWrapper, reqWrapper.Request)
        respWrapper.SendResponse()
    })
}
```

### 4. 数据转换
```go
func TransformMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 转换请求数据
        if reqWrapper.IsJSON() {
            var data map[string]interface{}
            if err := reqWrapper.GetJSONBody(&data); err == nil {
                // 添加默认字段
                if _, exists := data["version"]; !exists {
                    data["version"] = "1.0"
                }
                reqWrapper.SetJSONBody(data)
            }
        }
        
        next.ServeHTTP(respWrapper, reqWrapper.Request)
        
        // 转换响应数据
        if len(respWrapper.body) > 0 {
            var response map[string]interface{}
            if err := json.Unmarshal(respWrapper.body, &response); err == nil {
                // 添加元数据
                response["_metadata"] = map[string]interface{}{
                    "server_time": time.Now().Unix(),
                    "request_id":  reqWrapper.GetHeader("X-Request-ID"),
                }
                if newBody, err := json.Marshal(response); err == nil {
                    respWrapper.body = newBody
                }
            }
        }
        
        respWrapper.SendResponse()
    })
}
```

### 5. 限流控制
```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    clients := make(map[string]int)
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqWrapper := wrapper.NewRequestWrapper(r)
        respWrapper := wrapper.NewResponseWrapper(w)
        
        // 获取客户端IP
        clientIP := reqWrapper.RemoteAddr
        if forwardedFor := reqWrapper.GetHeader("X-Forwarded-For"); forwardedFor != "" {
            clientIP = forwardedFor
        }
        
        // 检查限流
        if count, exists := clients[clientIP]; exists && count >= 100 {
            respWrapper.WriteHeader(http.StatusTooManyRequests)
            respWrapper.Write([]byte(`{"error": "请求过于频繁"}`))
            respWrapper.SendResponse()
            return
        }
        
        clients[clientIP]++
        next.ServeHTTP(respWrapper, reqWrapper.Request)
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
    handler = LoggingMiddleware(handler)      // 日志记录
    handler = ValidationMiddleware(handler)   // 请求验证
    handler = RateLimitMiddleware(handler)    // 限流控制
    handler = TransformMiddleware(handler)    // 数据转换
    handler = EncryptionMiddleware(handler)   // 加密解密
    
    return handler
}
```

## 注意事项

1. **内存使用**: 包装器会将请求和响应体加载到内存中，对于大文件上传需要特别注意
2. **性能影响**: 包装器会带来一定的性能开销，建议在需要时使用
3. **错误处理**: 确保在中间件中正确处理错误，避免响应未发送
4. **并发安全**: 包装器本身不是并发安全的，每个请求应该使用独立的包装器实例

## API 参考

### RequestWrapper 方法

| 方法 | 描述 |
|------|------|
| `GetBody()` | 获取请求体字节数组 |
| `GetBodyString()` | 获取请求体字符串 |
| `SetBody([]byte)` | 设置请求体 |
| `GetJSONBody(interface{})` | 解析JSON请求体 |
| `SetJSONBody(interface{})` | 设置JSON请求体 |
| `GetQueryParam(string)` | 获取查询参数 |
| `SetQueryParam(string, string)` | 设置查询参数 |
| `GetFormParam(string)` | 获取表单参数 |
| `SetFormParam(string, string)` | 设置表单参数 |
| `GetHeader(string)` | 获取请求头 |
| `SetHeader(string, string)` | 设置请求头 |
| `GetCookie(string)` | 获取Cookie |
| `IsJSON()` | 判断是否为JSON请求 |
| `IsForm()` | 判断是否为表单请求 |
| `Clone()` | 克隆请求包装器 |

### ResponseWrapper 方法

| 方法 | 描述 |
|------|------|
| `Write([]byte)` | 写入响应体 |
| `WriteHeader(int)` | 设置状态码 |
| `Header()` | 获取响应头 |
| `Flush()` | 刷新缓冲区 |
| `Hijack()` | 获取底层连接 |
| `SendResponse()` | 发送完整响应 |

## 许可证

MIT License 