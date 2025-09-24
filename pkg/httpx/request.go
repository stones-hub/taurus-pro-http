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

// 修改于 2025-07-30
// author: yelei
// 注意：
// 1. 如果需要重复读取请求体数据， 请使用ParseStreamReusable
// 2. 凡是调用了r.Body.Close()的函数， 后续无法再读取请求体数据, 除非我们将r.Body重新被设置回io.NopCloser（io.NopCloser 会忽略关闭操作）
// 3. 不管是 io.ReadAll 还是 json.NewDecoder或者其他的读取方式， 读取后都会将偏移量移动到读取最后的位置，后续无法再读取之前已读取的数据
// 4. 我们将r.Body重新被设置回io.NopCloser, 虽然每次调用都能读取到数据，但是不建议这样做，因为每次重新设置都会创建新的内存缓冲区， 如果数据量很大， 会导致内存占用过高
package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetParam 获取 GET 提交的URL参数 或 POST 提交的表单(application/x-www-form-urlencoded)数据，兼容数组
func GetParams(r *http.Request, key string) ([]string, error) {
	// 解析查询参数
	if values, ok := r.URL.Query()[key]; ok {
		return values, nil
	}

	defer r.Body.Close()

	// 解析表单数据
	if err := r.ParseForm(); err == nil {
		if values, ok := r.Form[key]; ok {
			return values, nil
		}
	}

	return nil, fmt.Errorf("key %s not found", key)
}

// GetParam 获取 GET 提交的URL参数 或 POST 提交的表单(application/x-www-form-urlencoded)数据，不兼容数组
func GetParam(r *http.Request, key string) (string, error) {
	if res, err := GetParams(r, key); err != nil {
		return "", err
	} else {
		if len(res) == 0 {
			return "", fmt.Errorf("key %s not found", key)
		}

		return res[0], nil
	}
}

// ParseJson 获取非表单提交的 JSON 对象数据, 返回map
func ParseJson(r *http.Request) (map[string]interface{}, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("content type is not application/json")
	}

	defer r.Body.Close()
	var jsonData map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}

	return jsonData, nil
}

// ParseStreamReusable 获取请求体数据，并支持重复读取
// 读取后会将数据重新设置回请求体，以便后续可以再次读取
// 虽然每次调用都能读取到数据，但是不建议这样做，因为每次重新设置都会创建新的内存缓冲区， 如果数据量很大， 会导致内存占用过高
func ParseStreamReusable(r *http.Request) ([]byte, error) {
	defer r.Body.Close() // 这里关闭并不会影响后续读取， 因为我们将r.Body重新设置回io.NopCloser, NopCloser 会忽略关闭操作
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// 将读取的数据重新设置回请求体，以便后续可以再次读取
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// ParseStream 获取请求体数据（一次性读取，读取后无法再次读取）
func ParseStream(r *http.Request) ([]byte, error) {
	defer r.Body.Close()            // 这里会真正关闭原始的body，后续无法再读取
	body, err := io.ReadAll(r.Body) // 通过ReadAll读取后，偏移量会移动到末尾，后续无法再读取（哪怕不关闭r.Body，也无法再读取）
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return body, nil
}

// ParseJsonArray 获取非表单提交的  JSON 数组数据, 返回数组
func ParseJsonArray(r *http.Request) ([]interface{}, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("content type is not application/json")
	}

	defer r.Body.Close()
	var jsonArray []interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&jsonArray); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array body: %w", err)
	}

	return jsonArray, nil
}

// ParseJsonFlexible 获取非表单提交的  JSON 数据, 返回目标类型(不局限于数组还是对象)
func ParseJsonFlexible(r *http.Request, target interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return fmt.Errorf("content type is not application/json")
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to parse JSON body: %w", err)
	}

	return nil
}

// ParseText 获取非表单提交的纯文本数据
func ParseText(r *http.Request) (string, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		return "", fmt.Errorf("content type is not text/plain")
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read plain text body: %w", err)
	}

	return string(body), nil
}

// ParseMultipartFile 解析(multipart/form-data)表单上传的文件
func ParseMultipartFile(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	// 解析 multipart/form-data, 10MB 内存缓冲， 如果文件不上传完， 会报错， 所以当前函数只要返回没有错误， 就可以返回数据给客户端，不用等待
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max memory
		return nil, fmt.Errorf("failed to parse multipart form data: %w", err)
	}

	defer r.Body.Close()

	// 获取文件数据
	if files, ok := r.MultipartForm.File[key]; ok {
		return files, nil
	}

	return nil, fmt.Errorf("file key %s not found", key)
}

// ParseMultipartData 解析 multipart/form-data 请求，获取所有文件和参数数据
func ParseMultipartData(r *http.Request) (map[string][]*multipart.FileHeader, map[string][]string, error) {
	// 解析 multipart/form-data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max memory
		return nil, nil, fmt.Errorf("failed to parse multipart form data: %w", err)
	}
	defer r.Body.Close()

	// 获取所有文件数据
	files := r.MultipartForm.File

	// 获取参数数据
	params := r.MultipartForm.Value

	return files, params, nil
}

// GetPathParam 获取 URL 路径参数的值
// 适用于 Go 1.22+ 的动态路由，如 /video/{userid}/get
// 参数: r - HTTP 请求对象, key - 路径参数名
// 返回: 路径参数值, 错误信息
// 示例: userid := httpx.GetPathParam(r, "userid")
func GetPathParam(r *http.Request, key string) (string, error) {
	value := r.PathValue(key)
	if value == "" {
		return "", fmt.Errorf("path parameter %s not found", key)
	}
	return value, nil
}

// GetPathParamDefault 获取 URL 路径参数的值，如果不存在则返回默认值
// 参数: r - HTTP 请求对象, key - 路径参数名, defaultValue - 默认值
// 返回: 路径参数值或默认值
// 示例: userid := httpx.GetPathParamDefault(r, "userid", "unknown")
func GetPathParamDefault(r *http.Request, key, defaultValue string) string {
	value := r.PathValue(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// SaveUploadFiles 将文件数据存储到指定目录
func SaveUploadFiles(files []*multipart.FileHeader, destDir string) error {
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		// 创建目标文件
		destPath := filepath.Join(destDir, fileHeader.Filename)
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}
		defer destFile.Close()

		// 将上传的文件内容复制到目标文件
		if _, err := io.Copy(destFile, file); err != nil {
			return fmt.Errorf("failed to save file %s: %w", destPath, err)
		}
	}
	return nil
}
