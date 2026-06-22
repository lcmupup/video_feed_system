package handler

import (
	"net/http"
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
