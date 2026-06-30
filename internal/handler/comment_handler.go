package handler

import (
	"strconv"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

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
	userID, _ := c.Get("user_id")

	var req service.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的视频ID")
		return
	}
	req.VideoID = uint(videoID)
	req.UserID = userID.(uint)

	commentInfo, err := h.commentService.CreateComment(&req)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithData(c, "评论成功", commentInfo)
}

// DeleteComment 删除评论
// DELETE /api/comment/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID, _ := c.Get("user_id")

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的评论ID")
		return
	}

	err = h.commentService.DeleteComment(uint(commentID), userID.(uint))
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.Success(c, "删除成功")
}

// GetComments 获取视频评论列表
// GET /api/video/:id/comments?page=1&page_size=10
func (h *CommentHandler) GetComments(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的视频ID")
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

	comments, total, err := h.commentService.GetCommentsByVideoID(uint(videoID), page, pageSize)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "获取成功", comments, total, page, pageSize)
}
