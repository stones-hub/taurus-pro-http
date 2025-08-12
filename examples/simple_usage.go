package main

import (
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

// 在项目启动时调用此函数
func initGoRuntime() {
	log.Println("🔧 初始化Go运行时配置...")

	// 设置CPU核心数 (8核系统)
	runtime.GOMAXPROCS(8)

	// 设置GC参数 (平衡性能和内存)
	debug.SetGCPercent(150)

	// 设置内存限制 (12GB，留4GB给系统)
	debug.SetMemoryLimit(12 * 1024 * 1024 * 1024)

	log.Println("✅ Go运行时配置完成")
}

// 示例：如何在项目启动时设置Go运行时参数
func runServer() {
	// 1. 初始化Go运行时配置
	initGoRuntime()

	// 2. 创建HTTP服务器
	srv := server.New(server.Config{
		Addr:           ":8080",
		ReadTimeout:    60 * time.Second, // 8核16G优化
		WriteTimeout:   60 * time.Second,
		IdleTimeout:    300 * time.Second, // 5分钟keepalive
		MaxHeaderBytes: 1 << 20,           // 1MB
	})

	// 3. 启动服务器
	errChan := make(chan error, 1)
	srv.Start(errChan)

	log.Println("🚀 服务器已启动，端口: 8080")

	// 4. 等待服务器运行
	select {
	case err := <-errChan:
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 主函数
func main() {
	runServer()
}
