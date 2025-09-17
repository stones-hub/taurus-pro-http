package wrapper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// 注意: 并未提供Close函数来清理资源，因为谁负责调用 Close()？第一个中间件？但后续中间件还在使用，最后一个中间件？但不知道哪个是最后一个

// ResponseWrapper 响应包装器
// 用途：拦截响应数据，支持中间件处理
type ResponseWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

// NewResponseWrapper 创建响应包装器
func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
	}
}

// Write 拦截写入操作
func (rw *ResponseWrapper) Write(data []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.body.Write(data)
}

// WriteHeader 拦截状态码设置
func (rw *ResponseWrapper) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
	}
}

// GetStatusCode 获取状态码
func (rw *ResponseWrapper) GetStatusCode() int {
	return rw.statusCode
}

// GetBody 获取响应体
func (rw *ResponseWrapper) GetBody() []byte {
	return rw.body.Bytes()
}

// GetBodyString 获取响应体字符串
func (rw *ResponseWrapper) GetBodyString() string {
	return rw.body.String()
}

// SetBody 设置响应体
func (rw *ResponseWrapper) SetBody(body []byte) {
	rw.body.Reset()
	rw.body.Write(body)
}

// SendResponse 发送响应
func (rw *ResponseWrapper) SendResponse() {
	if rw.statusCode != 0 {
		rw.ResponseWriter.WriteHeader(rw.statusCode)
	}
	if rw.body.Len() > 0 {
		rw.ResponseWriter.Write(rw.body.Bytes())
	}
}

// Respond 发送响应（便捷方法）
func (rw *ResponseWrapper) Respond(statusCode int, data []byte) {
	rw.body.Reset()
	rw.body.Write(data)
	rw.WriteHeader(statusCode)
	rw.SendResponse()
}

// RespondWithJSON 发送JSON响应
func (rw *ResponseWrapper) RespondWithJSON(statusCode int, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		rw.Respond(500, []byte(`{"error": "JSON序列化失败"}`))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Respond(statusCode, jsonData)
}

// RespondWithText 发送文本响应
func (rw *ResponseWrapper) RespondWithText(statusCode int, text string) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.Respond(statusCode, []byte(text))
}

// RespondWithError 发送错误响应
func (rw *ResponseWrapper) RespondWithError(statusCode int, err error) {
	errorResponse := map[string]interface{}{
		"error":     err.Error(),
		"timestamp": time.Now().Unix(),
	}
	rw.RespondWithJSON(statusCode, errorResponse)
}

// Reset 重置包装器
func (rw *ResponseWrapper) Reset() {
	rw.body.Reset() // bytes.Buffer 会自动清空
	rw.statusCode = 0
	rw.written = false
}
