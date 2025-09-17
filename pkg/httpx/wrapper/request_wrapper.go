package wrapper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// RequestWrapper 请求包装器
type RequestWrapper struct {
	*http.Request
	body     *bytes.Buffer // 请求体
	bodySize int64         // 请求体大小
	maxSize  int64         // 最大允许的请求体大小
	bodyRead bool          // 是否已经读取了请求体
}

const (
	defaultBodyMaxSize = 1024 * 1024 // 1MB
)

// NewRequestWrapper 创建请求包装器
// maxSize: 最大允许的请求体大小（字节），0表示不限制
func NewRequestWrapper(r *http.Request, maxSize int64) *RequestWrapper {
	if maxSize <= 0 {
		maxSize = defaultBodyMaxSize
	}
	return &RequestWrapper{
		Request:  r,
		body:     &bytes.Buffer{},
		maxSize:  maxSize,
		bodySize: r.ContentLength,
	}
}

// ReadBody 读取请求体并检查大小限制
// 将请求体重的body数据读取到rw.body中， 并设置rw.bodyRead为true
func (rw *RequestWrapper) ReadBody() error {
	if rw.bodyRead {
		return nil // 已经读取过了
	}

	// 如果Content-Length为-1，说明是chunked编码，需要读取才能知道大小
	if rw.bodySize == -1 {
		body, err := io.ReadAll(rw.Request.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %w", err)
		}
		rw.body.Reset()
		rw.body.Write(body)
		rw.bodySize = int64(len(body))
	} else {
		// 有Content-Length，先检查大小
		if rw.maxSize > 0 && rw.bodySize > rw.maxSize {
			return fmt.Errorf("请求体大小超出限制: %d > %d 字节", rw.bodySize, rw.maxSize)
		}

		body, err := io.ReadAll(rw.Request.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %w", err)
		}
		rw.body.Reset()
		rw.body.Write(body)
	}

	// 检查实际读取的大小
	if rw.maxSize > 0 && rw.bodySize > rw.maxSize {
		return fmt.Errorf("请求体大小超出限制: %d > %d 字节", rw.bodySize, rw.maxSize)
	}

	// 将读取的数据重新设置回请求体，以便后续可以再次读取
	rw.Request.Body = io.NopCloser(bytes.NewReader(rw.body.Bytes()))
	rw.bodyRead = true

	return nil
}

// GetBodySize 获取请求体大小
func (rw *RequestWrapper) GetBodySize() (int64, error) {
	// 如果还没有读取请求体，返回Content-Length（可能为-1）
	if !rw.bodyRead {
		return 0, fmt.Errorf("请求体未读取")
	}
	return rw.bodySize, nil
}

// GetBody 获取请求体数据
func (rw *RequestWrapper) GetBody() ([]byte, error) {
	if !rw.bodyRead {
		if err := rw.ReadBody(); err != nil {
			return nil, err
		}
	}
	return rw.body.Bytes(), nil
}

// GetBodyString 获取请求体字符串
func (rw *RequestWrapper) GetBodyString() (string, error) {
	if !rw.bodyRead {
		if err := rw.ReadBody(); err != nil {
			return "", err
		}
	}
	return rw.body.String(), nil
}

// IsBodySizeExceeded 检查请求体大小是否超出限制
func (rw *RequestWrapper) IsBodySizeExceeded() bool {
	if rw.maxSize <= 0 {
		return false
	}
	// 如果还没有读取请求体，无法确定大小
	if !rw.bodyRead {
		return false
	}
	return rw.bodySize > rw.maxSize
}

// GetMaxSize 获取最大允许的请求体大小
func (rw *RequestWrapper) GetMaxSize() int64 {
	return rw.maxSize
}

// SetMaxSize 设置最大允许的请求体大小
func (rw *RequestWrapper) SetMaxSize(maxSize int64) {
	rw.maxSize = maxSize
}

// Reset 重置包装器状态
func (rw *RequestWrapper) Reset() {
	rw.body.Reset()
	rw.bodySize = 0
	rw.bodyRead = false
}

// Clone 克隆请求包装器（用于创建新的包装器实例）
func (rw *RequestWrapper) Clone() *RequestWrapper {
	clone := &RequestWrapper{
		Request:  rw.Request,
		body:     &bytes.Buffer{},
		maxSize:  rw.maxSize,
		bodySize: rw.bodySize,
		bodyRead: false,
	}
	// 如果原包装器已经读取了body，复制数据
	if rw.bodyRead {
		clone.body.Write(rw.body.Bytes())
		clone.bodyRead = true
	}
	return clone
}
