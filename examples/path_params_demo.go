package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
	"github.com/stones-hub/taurus-pro-http/pkg/middleware"
	"github.com/stones-hub/taurus-pro-http/pkg/router"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

// VideoHandler 处理视频相关的请求
type VideoHandler struct {
	videos map[string]Video
}

// Video 表示视频模型
type Video struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	UserID   string `json:"user_id"`
	Duration int    `json:"duration"`
}

// NewVideoHandler 创建视频处理器
func NewVideoHandler() *VideoHandler {
	// 模拟视频数据
	videos := map[string]Video{
		"1001": {ID: "1001", Title: "视频1", UserID: "1001", Duration: 120},
		"1002": {ID: "1002", Title: "视频2", UserID: "1002", Duration: 180},
		"1003": {ID: "1003", Title: "视频3", UserID: "1001", Duration: 90},
	}
	return &VideoHandler{videos: videos}
}

// GetVideo 获取指定用户的视频
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	// 使用新的路径参数获取函数
	userid, err := httpx.GetPathParam(r, "userid")
	if err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": "缺少用户ID参数",
		})
		return
	}

	// 获取视频ID参数
	videoID, err := httpx.GetPathParam(r, "videoId")
	if err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": "缺少视频ID参数",
		})
		return
	}

	// 验证用户ID和视频ID的匹配
	video, exists := h.videos[videoID]
	if !exists {
		httpx.SendResponse(w, http.StatusNotFound, nil, map[string]string{
			"error": "视频不存在",
		})
		return
	}

	if video.UserID != userid {
		httpx.SendResponse(w, http.StatusForbidden, nil, map[string]string{
			"error": "无权访问该视频",
		})
		return
	}

	httpx.SendResponse(w, http.StatusOK, video, nil)
}

// GetUserVideos 获取用户的所有视频
func (h *VideoHandler) GetUserVideos(w http.ResponseWriter, r *http.Request) {
	// 使用带默认值的路径参数获取函数
	userid := httpx.GetPathParamDefault(r, "userid", "unknown")

	// 获取查询参数（分页）
	pageStr, _ := httpx.GetParam(r, "page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}

	// 模拟分页查询
	var userVideos []Video
	for _, video := range h.videos {
		if video.UserID == userid {
			userVideos = append(userVideos, video)
		}
	}

	// 简单的分页逻辑
	pageSize := 10
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(userVideos) {
		userVideos = []Video{}
	} else {
		if end > len(userVideos) {
			end = len(userVideos)
		}
		userVideos = userVideos[start:end]
	}

	response := map[string]interface{}{
		"videos": userVideos,
		"page":   page,
		"total":  len(h.videos),
	}

	httpx.SendResponse(w, http.StatusOK, response, nil)
}

// CreateVideo 创建新视频
func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID
	userid, err := httpx.GetPathParam(r, "userid")
	if err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": "缺少用户ID参数",
		})
		return
	}

	// 解析请求体
	var videoData struct {
		Title    string `json:"title"`
		Duration int    `json:"duration"`
	}

	if err := httpx.ParseJsonFlexible(r, &videoData); err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": "无效的JSON数据",
		})
		return
	}

	// 创建新视频
	videoID := strconv.FormatInt(time.Now().Unix(), 10)
	newVideo := Video{
		ID:       videoID,
		Title:    videoData.Title,
		UserID:   userid,
		Duration: videoData.Duration,
	}

	h.videos[videoID] = newVideo

	httpx.SendResponse(w, http.StatusCreated, newVideo, nil)
}

func runPathParamsDemo() {
	// 创建服务器实例
	srv := server.NewServer(
		server.WithAddr(":8080"),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(15*time.Second),
		server.WithIdleTimeout(30*time.Second),
	)

	// 创建处理器
	videoHandler := NewVideoHandler()

	// 创建中间件
	corsMiddleware := middleware.CorsMiddleware(nil)

	// 视频相关的路由组
	srv.AddRouterGroup(router.RouteGroup{
		Prefix: "/api/v1",
		Middleware: []router.MiddlewareFunc{
			corsMiddleware,
		},
		Routes: []router.Router{
			// 获取指定用户的指定视频
			{
				Path:    "/video/{userid}/{videoId}",
				Handler: http.HandlerFunc(videoHandler.GetVideo),
			},
			// 获取用户的所有视频
			{
				Path:    "/user/{userid}/videos",
				Handler: http.HandlerFunc(videoHandler.GetUserVideos),
			},
			// 创建新视频
			{
				Path:    "/user/{userid}/video",
				Handler: http.HandlerFunc(videoHandler.CreateVideo),
			},
		},
	})

	// 健康检查路由
	srv.AddRouter(router.Router{
		Path: "/health",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.SendResponse(w, http.StatusOK, map[string]string{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			}, nil)
		}),
	})

	// 启动服务器
	errChan := make(chan error, 1)
	srv.Start(errChan)

	log.Println("🚀 路径参数演示服务器已启动，端口: 8080")
	log.Println("📝 测试路径参数功能:")
	log.Println("   GET  /api/v1/video/{userid}/{videoId}  - 获取指定视频")
	log.Println("   GET  /api/v1/user/{userid}/videos      - 获取用户视频列表")
	log.Println("   POST /api/v1/user/{userid}/video       - 创建新视频")
	log.Println("   GET  /health                           - 健康检查")
	log.Println("")
	log.Println("🔍 测试示例:")
	log.Println("   curl http://localhost:8080/api/v1/video/1001/1001")
	log.Println("   curl http://localhost:8080/api/v1/user/1001/videos")
	log.Println("   curl -X POST http://localhost:8080/api/v1/user/1001/video -H 'Content-Type: application/json' -d '{\"title\":\"新视频\",\"duration\":200}'")

	// 等待服务器运行
	if err := <-errChan; err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
