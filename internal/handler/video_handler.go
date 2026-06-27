package handler

import (
	"net/http"
	"path/filepath"
	"strconv"

	"video_feed/internal/service"

	"github.com/gin-gonic/gin"
)

// VideoHandler 视频HTTP处理器
type VideoHandler struct {
	videoService *service.VideoService
}

// NewVideoHandler 创建视频处理器实例
func NewVideoHandler() *VideoHandler {
	return &VideoHandler{
		videoService: service.NewVideoService(),
	}
}

// Upload 视频上传接口
// POST /api/video/upload
func (h *VideoHandler) Upload(c *gin.Context) {
	// 1. 解析参数
	var req service.UploadVideoReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请检查上传参数: " + err.Error(),
		})
		return
	}

	// 2. 从JWT中获取用户ID（需要先登录）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}
	req.UserID = userID.(uint)

	// 3. 处理上传
	videoInfo, err := h.videoService.UploadVideo(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 4. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "上传成功",
		"data": videoInfo,
	})
}

// GetFeed 获取Feed流接口
// GET /api/feed?page=1&page_size=10
func (h *VideoHandler) GetFeed(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")           //第几页
	pageSizeStr := c.DefaultQuery("page_size", "10") //页面大小

	page, err := strconv.Atoi(pageStr) // 字符串转整数
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr) // 字符串转整数
	if err != nil || pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	videos, total, err := h.videoService.GetFeed(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"msg":       "获取成功",
		"data":      videos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Play 视频播放接口（新增）
// GET /api/video/play/:id
func (h *VideoHandler) Play(c *gin.Context) {
	// 1. 获取视频ID
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64) // 将字符串 videoIDStr 解析为一个无符号的 64 位整数
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	// 2. 从数据库获取视频信息
	video, err := h.videoService.GetVideoByID(uint(videoID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "视频不存在",
		})
		return
	}

	// 3. 获取视频文件的绝对路径
	videoPath := video.FilePath
	if !filepath.IsAbs(videoPath) { // 检查是否是绝对路径
		// 如果是相对路径，拼接当前工作目录
		absPath, _ := filepath.Abs(videoPath) // 将路径变成绝对路径
		videoPath = absPath
	}

	// 4. 设置响应头并返回视频文件
	c.Header("Content-Type", "video/mp4") // 告诉浏览器响应体的内容类型是 MP4 视频
	// filename="xxx.mp4" 表示建议的文件名，如果用户选择下载，浏览器会预填这个文件名
	c.Header("Content-Disposition", "inline; filename=\""+video.Title+".mp4\"") // 控制浏览器如何展示内容——是在页面内显示（inline），还是作为附件下载（attachment）
	c.File(videoPath)                                                           // 从本地文件系统读取视频文件，并将其内容作为响应体发送给客户端
}

// Search 搜索视频接口
// GET /api/video/search?keyword=xxx&page=1&page_size=10
func (h *VideoHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "搜索关键词不能为空",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	videos, total, err := h.videoService.SearchVideos(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"msg":       "搜索成功",
		"data":      videos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"keyword":   keyword,
	})
}
