package wrapper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
)

// ResponseWrapper 完整的ResponseWriter封装示例
type ResponseWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
	written    bool
}

// NewResponseWrapper 创建新的响应包装器
func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
		headers:        make(http.Header),
	}
}

// Write 重写Write方法 - 收集响应体
// 用途：拦截处理器写入的所有响应数据，收集到内存中而不是立即发送
// 重要性：核心方法，必须实现。这是实现响应处理（如加密、压缩、日志）的关键
// 参数：data - 要写入的字节数据
// 返回：写入的字节数和错误信息
func (rw *ResponseWrapper) Write(data []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}

	// 限制响应体大小
	if len(data) > 1024*1024 {
		return 0, errors.New("response body is too large, max size is 1MB")
	}

	rw.body.Write(data)
	return len(data), nil
}

// Reset 完全重置响应包装器
// 用途：清空所有响应数据，给外部重新设置的机会
// 重要性：核心方法，用于错误处理和响应重写
// 执行操作：清空body、headers、statusCode，重置written状态
func (rw *ResponseWrapper) Reset() {
	rw.body.Reset()
	rw.headers = make(http.Header)
	rw.statusCode = 0
	rw.written = false
}

// WriteHeader 重写WriteHeader方法 - 记录状态码
// 用途：拦截处理器设置的状态码，记录但不立即发送HTTP头
// 重要性：核心方法，必须实现。确保状态码在最终处理（如加密）后正确发送
// 参数：statusCode - HTTP状态码（如200、404、500等）
func (rw *ResponseWrapper) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
	}
}

// Header 重写Header方法 - 收集响应头
// 用途：拦截处理器设置的响应头，收集到本地而不是直接设置到原始ResponseWriter
// 重要性：核心方法，必须实现。确保响应头在最终处理（如加密）后正确发送
// 返回：响应头集合，处理器可以设置Content-Type、Cookie、CORS等头部
func (rw *ResponseWrapper) Header() http.Header {
	return rw.headers
}

// Flush 实现Flusher接口
// 用途：支持流式响应，当处理器调用Flush时立即发送缓冲数据
// 重要性：功能增强方法，按需实现。保持与流式响应的兼容性
// 应用场景：实时日志、SSE（Server-Sent Events）、聊天应用
func (rw *ResponseWrapper) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现Hijacker接口
// 用途：获取底层TCP连接，用于WebSocket升级、HTTP/2推送等高级功能
// 重要性：功能增强方法，按需实现。支持协议升级和长连接
// 应用场景：WebSocket、HTTP/2服务器推送、自定义协议
// 返回：原始TCP连接、缓冲读写器、错误信息
func (rw *ResponseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// CloseNotify 实现CloseNotifier接口（已弃用）
// 注意：http.CloseNotifier已被弃用，新代码应使用Request.Context()
// 用途：监听客户端连接关闭事件，用于长连接场景
// 重要性：兼容性方法，已弃用。建议使用Request.Context()替代
// 应用场景：长轮询、实时通知、聊天应用（使用Context替代）
// 返回：连接关闭通知通道，当客户端断开时通道会收到信号
func (rw *ResponseWrapper) CloseNotify() <-chan bool {
	// 注意：http.CloseNotifier已被弃用，建议使用Request.Context()
	// 新代码应该使用：r.Context().Done() 来监听连接关闭
	if notifier, ok := rw.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	return nil
}

// Push 实现Pusher接口 (HTTP/2)
// 用途：HTTP/2服务器推送，主动向客户端推送资源
// 重要性：性能优化方法，高级功能。提高页面加载性能
// 应用场景：CSS/JS预加载、图片预推送、资源优化
// 参数：target - 推送的资源路径，opts - 推送选项（如头部信息）
// 返回：推送操作的结果
func (rw *ResponseWrapper) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// ReadFrom 实现io.ReaderFrom接口
// 用途：高效文件传输，使用零拷贝技术直接从文件读取到网络
// 重要性：性能优化方法，高级功能。比io.Copy更高效
// 应用场景：大文件下载、视频流、图片传输
// 参数：r - 数据源（如文件、网络连接等）
// 返回：传输的字节数和错误信息
func (rw *ResponseWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	if readerFrom, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(r)
	}
	return 0, http.ErrNotSupported
}

// GetStatusCode 获取状态码
func (rw *ResponseWrapper) GetStatusCode() int {
	return rw.statusCode
}

// GetHeaders 获取响应头
func (rw *ResponseWrapper) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for key, values := range rw.headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// GetBody 获取响应体
func (rw *ResponseWrapper) GetBody() []byte {
	return rw.body.Bytes()
}

// GetBodyString 获取响应体字符串
func (rw *ResponseWrapper) GetBodyString() string {
	return rw.body.String()
}

// SendResponse 发送收集到的响应
// 用途：将收集到的响应头、状态码和响应体发送到客户端
// 重要性：核心方法，必须实现。这是响应处理的最终步骤
// 执行顺序：1. 设置响应头 2. 发送状态码 3. 发送响应体
// 注意：WriteHeader只能调用一次，必须在Write之前调用
func (rw *ResponseWrapper) SendResponse() {
	// 复制收集的头部到原始ResponseWriter
	for key, values := range rw.headers {
		for _, value := range values {
			rw.ResponseWriter.Header().Add(key, value)
		}
	}

	// 发送状态码和响应体
	if rw.statusCode != 0 {
		rw.ResponseWriter.WriteHeader(rw.statusCode)
	}

	if len(rw.body.Bytes()) > 0 {
		rw.ResponseWriter.Write(rw.body.Bytes())
	}
}

// Respond 优雅的响应发送方法
// 用途：一次性设置并发送完整的HTTP响应
// 重要性：便捷方法，提供统一的响应发送接口
// 参数：statusCode - HTTP状态码，headers - 响应头部，data - 响应数据
func (rw *ResponseWrapper) Respond(statusCode int, headers map[string]string, data []byte) {
	// 清空之前的响应数据
	rw.Reset()

	// 设置状态码
	rw.WriteHeader(statusCode)

	// 设置响应头部
	for key, value := range headers {
		rw.Header().Set(key, value)
	}

	// 写入响应数据
	if len(data) > 0 {
		rw.Write(data)
	}

	// 发送响应
	rw.SendResponse()
}

// RespondWithJSON 发送JSON格式响应
// 用途：便捷的JSON响应发送方法
// 重要性：常用方法，适用于API开发
// 参数：statusCode - HTTP状态码，headers - 响应头部，data - JSON数据
func (rw *ResponseWrapper) RespondWithJSON(statusCode int, headers map[string]string, data interface{}) {
	// 确保Content-Type为application/json
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	// 序列化JSON数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		// 序列化失败，发送错误响应
		rw.Respond(statusCode, headers, []byte(`{"error": "JSON序列化失败"}`))
		return
	}

	// 发送JSON响应
	rw.Respond(statusCode, headers, jsonData)
}

// RespondWithText 发送文本格式响应
// 用途：便捷的文本响应发送方法
// 重要性：常用方法，适用于简单文本响应
// 参数：statusCode - HTTP状态码，headers - 响应头部，text - 文本内容
func (rw *ResponseWrapper) RespondWithText(statusCode int, headers map[string]string, text string) {
	// 确保Content-Type为text/plain
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "text/plain"

	// 发送文本响应
	rw.Respond(statusCode, headers, []byte(text))
}
