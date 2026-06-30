package handler

import (
	"path/filepath"
	"strconv"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

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
	var req service.UploadVideoReq
	if err := c.ShouldBind(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "请检查上传参数: "+err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errcode.Error(c, errcode.CodeUnauthorized, "请先登录")
		return
	}
	req.UserID = userID.(uint)

	videoInfo, err := h.videoService.UploadVideo(&req)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithData(c, "上传成功", videoInfo)
}

// GetFeed 获取Feed流接口
// GET /api/feed?page=1&page_size=10
func (h *VideoHandler) GetFeed(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	videos, total, err := h.videoService.GetFeed(page, pageSize)
	if err != nil {
		errcode.Error(c, errcode.CodeInternalError, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "获取成功", videos, total, page, pageSize)
}

// Play 视频播放接口
// GET /api/video/play/:id
func (h *VideoHandler) Play(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的视频ID")
		return
	}

	video, err := h.videoService.GetVideoByID(uint(videoID))
	if err != nil {
		errcode.Error(c, errcode.CodeNotFound, "视频不存在")
		return
	}

	videoPath := video.FilePath
	if !filepath.IsAbs(videoPath) {
		absPath, _ := filepath.Abs(videoPath)
		videoPath = absPath
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", "inline; filename=\""+video.Title+".mp4\"")
	c.File(videoPath)
}

// Search 搜索视频接口
// GET /api/video/search?keyword=xxx&page=1&page_size=10
func (h *VideoHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		errcode.Error(c, errcode.CodeBadRequest, "搜索关键词不能为空")
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
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "搜索成功", videos, total, page, pageSize)
}
