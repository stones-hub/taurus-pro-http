package wrapper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RequestWrapper 完整的Request封装示例
type RequestWrapper struct {
	*http.Request
	body         []byte
	headers      http.Header
	queryParams  url.Values
	formParams   url.Values
	bodyRead     bool
	originalBody io.ReadCloser
}

// NewRequestWrapper 创建新的请求包装器
func NewRequestWrapper(r *http.Request) *RequestWrapper {
	// 复制请求以避免修改原始请求
	reqCopy := *r
	reqCopy.Header = make(http.Header)
	for k, v := range r.Header {
		reqCopy.Header[k] = v
	}

	// 复制URL以避免修改原始URL
	if r.URL != nil {
		reqCopy.URL = &url.URL{}
		*reqCopy.URL = *r.URL
	}

	return &RequestWrapper{
		Request:      &reqCopy,
		headers:      make(http.Header),
		queryParams:  make(url.Values),
		formParams:   make(url.Values),
		originalBody: r.Body,
	}
}

// Body 重写Body方法 - 拦截请求体读取
// 用途：拦截处理器读取的请求体，收集到内存中以便后续处理
// 重要性：核心方法，必须实现。这是实现请求处理（如解密、验证、日志）的关键
// 返回：请求体的io.ReadCloser接口
func (rw *RequestWrapper) Body() io.ReadCloser {
	if !rw.bodyRead {
		rw.readBody()
	}
	return io.NopCloser(bytes.NewReader(rw.body))
}

// readBody 读取并缓存请求体
// 用途：将原始请求体读取到内存中，支持多次读取
// 重要性：内部方法，确保请求体可以被多次访问
func (rw *RequestWrapper) readBody() {
	if rw.originalBody != nil {
		bodyBytes, err := io.ReadAll(rw.originalBody)
		if err == nil {
			rw.body = bodyBytes
		}
		rw.originalBody.Close()
	}
	rw.bodyRead = true
}

// GetBody 获取请求体字节数组
// 用途：直接获取请求体的字节数据，用于处理和分析
// 重要性：便捷方法，提供对请求体数据的直接访问
// 返回：请求体的字节数组
func (rw *RequestWrapper) GetBody() []byte {
	if !rw.bodyRead {
		rw.readBody()
	}
	return rw.body
}

// SetBody 设置请求体
// 用途：修改请求体内容，用于请求转换或处理
// 重要性：核心方法，支持请求体的动态修改
// 参数：body - 新的请求体字节数组
func (rw *RequestWrapper) SetBody(body []byte) {
	rw.body = body
	rw.bodyRead = true
	// 更新Content-Length头部
	rw.Header.Set("Content-Length", strconv.Itoa(len(body)))
}

// GetBodyString 获取请求体字符串
// 用途：以字符串形式获取请求体，便于文本处理
// 重要性：便捷方法，适用于JSON、XML等文本格式的请求体
// 返回：请求体的字符串表示
func (rw *RequestWrapper) GetBodyString() string {
	return string(rw.GetBody())
}

// GetJSONBody 解析JSON请求体
// 用途：将JSON格式的请求体解析为Go结构体
// 重要性：便捷方法，适用于API开发中的JSON处理
// 参数：v - 目标结构体指针
// 返回：解析错误
func (rw *RequestWrapper) GetJSONBody(v interface{}) error {
	body := rw.GetBody()
	return json.Unmarshal(body, v)
}

// SetJSONBody 设置JSON请求体
// 用途：将Go结构体序列化为JSON并设置为请求体
// 重要性：便捷方法，适用于API开发中的JSON处理
// 参数：v - 要序列化的结构体
// 返回：序列化错误
func (rw *RequestWrapper) SetJSONBody(v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.SetBody(body)
	rw.Header.Set("Content-Type", "application/json")
	return nil
}

// GetQueryParam 获取URL查询参数
// 用途：获取URL中的查询参数值
// 重要性：便捷方法，提供对查询参数的访问
// 参数：key - 参数名
// 返回：参数值，如果不存在则返回空字符串
func (rw *RequestWrapper) GetQueryParam(key string) string {
	if rw.queryParams == nil {
		rw.queryParams = rw.URL.Query()
	}
	return rw.queryParams.Get(key)
}

// SetQueryParam 设置URL查询参数
// 用途：添加或修改URL查询参数
// 重要性：便捷方法，支持动态修改查询参数
// 参数：key - 参数名，value - 参数值
func (rw *RequestWrapper) SetQueryParam(key, value string) {
	if rw.queryParams == nil {
		rw.queryParams = rw.URL.Query()
	}
	rw.queryParams.Set(key, value)
	rw.URL.RawQuery = rw.queryParams.Encode()
}

// GetFormParam 获取表单参数
// 用途：获取POST表单中的参数值
// 重要性：便捷方法，适用于表单处理
// 参数：key - 参数名
// 返回：参数值，如果不存在则返回空字符串
func (rw *RequestWrapper) GetFormParam(key string) string {
	if rw.formParams == nil {
		rw.parseForm()
	}
	return rw.formParams.Get(key)
}

// SetFormParam 设置表单参数
// 用途：添加或修改表单参数
// 重要性：便捷方法，支持动态修改表单参数
// 参数：key - 参数名，value - 参数值
func (rw *RequestWrapper) SetFormParam(key, value string) {
	if rw.formParams == nil {
		rw.formParams = make(url.Values)
	}
	rw.formParams.Set(key, value)
}

// parseForm 解析表单数据
// 用途：解析请求体中的表单数据
// 重要性：内部方法，支持表单参数的处理
func (rw *RequestWrapper) parseForm() {
	if rw.formParams == nil {
		rw.formParams = make(url.Values)
	}

	contentType := rw.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		body := rw.GetBodyString()
		if parsed, err := url.ParseQuery(body); err == nil {
			rw.formParams = parsed
		}
	}
}

// GetHeader 获取请求头
// 用途：获取指定请求头的值
// 重要性：便捷方法，提供对请求头的访问
// 参数：key - 头部名称
// 返回：头部值，如果不存在则返回空字符串
func (rw *RequestWrapper) GetHeader(key string) string {
	return rw.Header.Get(key)
}

// SetHeader 设置请求头
// 用途：添加或修改请求头
// 重要性：便捷方法，支持动态修改请求头
// 参数：key - 头部名称，value - 头部值
func (rw *RequestWrapper) SetHeader(key, value string) {
	rw.Header.Set(key, value)
}

// AddHeader 添加请求头
// 用途：添加请求头，支持多个同名头部
// 重要性：便捷方法，适用于需要多个同名头部的场景
// 参数：key - 头部名称，value - 头部值
func (rw *RequestWrapper) AddHeader(key, value string) {
	rw.Header.Add(key, value)
}

// GetCookie 获取Cookie
// 用途：获取指定名称的Cookie
// 重要性：便捷方法，提供对Cookie的访问
// 参数：name - Cookie名称
// 返回：Cookie对象，如果不存在则返回nil
func (rw *RequestWrapper) GetCookie(name string) *http.Cookie {
	for _, cookie := range rw.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// SetCookie 设置Cookie
// 用途：添加或修改Cookie
// 重要性：便捷方法，支持动态修改Cookie
// 参数：cookie - Cookie对象
func (rw *RequestWrapper) SetCookie(cookie *http.Cookie) {
	rw.Header.Add("Set-Cookie", cookie.String())
}

// GetUserAgent 获取User-Agent
// 用途：获取客户端的User-Agent信息
// 重要性：便捷方法，常用于客户端识别
// 返回：User-Agent字符串
func (rw *RequestWrapper) GetUserAgent() string {
	return rw.GetHeader("User-Agent")
}

// GetContentType 获取Content-Type
// 用途：获取请求的Content-Type
// 重要性：便捷方法，用于判断请求体格式
// 返回：Content-Type字符串
func (rw *RequestWrapper) GetContentType() string {
	return rw.GetHeader("Content-Type")
}

// IsJSON 判断是否为JSON请求
// 用途：判断请求体是否为JSON格式
// 重要性：便捷方法，用于请求格式判断
// 返回：是否为JSON格式
func (rw *RequestWrapper) IsJSON() bool {
	contentType := rw.GetContentType()
	return strings.Contains(contentType, "application/json")
}

// IsForm 判断是否为表单请求
// 用途：判断请求体是否为表单格式
// 重要性：便捷方法，用于请求格式判断
// 返回：是否为表单格式
func (rw *RequestWrapper) IsForm() bool {
	contentType := rw.GetContentType()
	return strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}

// Clone 克隆请求包装器
// 用途：创建请求包装器的副本，避免修改原始请求
// 重要性：核心方法，支持请求的复制和修改
// 返回：新的请求包装器实例
func (rw *RequestWrapper) Clone() *RequestWrapper {
	// 创建新的请求副本
	reqCopy := *rw.Request
	reqCopy.Header = make(http.Header)
	for k, v := range rw.Request.Header {
		reqCopy.Header[k] = v
	}

	// 复制URL
	if rw.Request.URL != nil {
		reqCopy.URL = &url.URL{}
		*reqCopy.URL = *rw.Request.URL
	}

	// 创建新的包装器
	newWrapper := &RequestWrapper{
		Request:     &reqCopy,
		headers:     make(http.Header),
		queryParams: make(url.Values),
		formParams:  make(url.Values),
		bodyRead:    rw.bodyRead,
	}

	// 复制数据
	copy(newWrapper.body, rw.body)
	for k, v := range rw.headers {
		newWrapper.headers[k] = v
	}
	for k, v := range rw.queryParams {
		newWrapper.queryParams[k] = v
	}
	for k, v := range rw.formParams {
		newWrapper.formParams[k] = v
	}

	return newWrapper
}

// ToHTTPRequest 转换为标准HTTP请求
// 用途：将包装器转换为标准的http.Request，用于外部调用
// 重要性：核心方法，支持与外部库的兼容
// 返回：标准的http.Request对象
func (rw *RequestWrapper) ToHTTPRequest() *http.Request {
	req := *rw.Request

	// 设置修改后的头部
	for k, v := range rw.headers {
		req.Header[k] = v
	}

	// 设置修改后的URL
	if rw.queryParams != nil {
		req.URL.RawQuery = rw.queryParams.Encode()
	}

	// 设置修改后的请求体
	if rw.bodyRead && len(rw.body) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(rw.body))
		req.ContentLength = int64(len(rw.body))
	}

	return &req
}
