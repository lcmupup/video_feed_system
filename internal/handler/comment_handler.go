package handler

import (
	"net/http"
	"strconv"
	"video_feed/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentHandler 评论处理器
type CommentHandler struct {
	commentService *service.CommentService
}

// NewCommentHandler 创建评论处理器实例
func NewCommentHandler() *CommentHandler {
	return &CommentHandler{
		commentService: service.NewCommentService(),
	}
}

// CreateComment 发表评论
// POST /api/video/:id/comment
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID, _ := c.Get("user_id") // jwt里面已经把用户id存到了上下文里，这里可以直接调用

	var req service.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 从URL中获取视频ID
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}
	req.VideoID = uint(videoID)
	req.UserID = userID.(uint)

	commentInfo, err := h.commentService.CreateComment(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "评论成功",
		"data": commentInfo,
	})
}

// DeleteComment 删除评论
// DELETE /api/comment/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID, _ := c.Get("user_id")

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的评论ID",
		})
		return
	}

	err = h.commentService.DeleteComment(uint(commentID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// GetComments 获取视频评论列表
// GET /api/video/:id/comments?page=1&page_size=10
func (h *CommentHandler) GetComments(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	// 提取URL里的Query String
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 检查page和pageSize的合法范围
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	// 执行获取视频的评论列表服务
	comments, total, err := h.commentService.GetCommentsByVideoID(uint(videoID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"msg":       "获取成功",
		"data":      comments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
