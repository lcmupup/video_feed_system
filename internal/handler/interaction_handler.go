package handler

import (
	"net/http"
	"strconv"
	"video_feed/internal/service"

	"github.com/gin-gonic/gin"
)

// InteractionHandler 互动HTTP处理器
type InteractionHandler struct {
	interService *service.InteractionService
}

// NewInteractionHandler 创建互动处理器实例
func NewInteractionHandler() *InteractionHandler {
	return &InteractionHandler{
		interService: service.NewInteractionService(),
	}
}

// ========== 点赞相关 ==========

// LikeVideo 点赞视频
// POST /api/video/:id/like
func (h *InteractionHandler) LikeVideo(c *gin.Context) {
	// 获取用户ID和视频ID
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64) // 将字符串转为无符号整数
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	// 执行点赞服务
	err = h.interService.LikeVideo(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 返回正确响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "点赞成功",
	})
}

// UnlikeVideo 取消点赞
// DELETE /api/video/:id/like
func (h *InteractionHandler) UnlikeVideo(c *gin.Context) {
	// 获取用户ID和视频ID
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	// 执行取消点赞服务
	err = h.interService.UnlikeVideo(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 返回正确响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "取消点赞成功",
	})
}

// GetLikeStatus 获取点赞状态
// GET /api/video/:id/like
func (h *InteractionHandler) GetLikeStatus(c *gin.Context) {
	// 获取用户ID和视频ID
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	// 执行获取点赞状态服务
	status, err := h.interService.GetVideoLikeStatus(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 返回正确响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": status,
	})
}

// ========== 收藏相关 ==========

// FavoriteVideo 收藏视频
// POST /api/video/:id/favorite
func (h *InteractionHandler) FavoriteVideo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	err = h.interService.FavoriteVideo(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "收藏成功",
	})
}

// UnfavoriteVideo 取消收藏
// DELETE /api/video/:id/favorite
func (h *InteractionHandler) UnfavoriteVideo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	err = h.interService.UnfavoriteVideo(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "取消收藏成功",
	})
}

// GetFavoriteStatus 获取收藏状态
// GET /api/video/:id/favorite
func (h *InteractionHandler) GetFavoriteStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的视频ID",
		})
		return
	}

	status, err := h.interService.GetVideoFavoriteStatus(userID.(uint), uint(videoID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": status,
	})
}

// GetUserFavorites 获取用户收藏列表
// GET /api/user/favorites
func (h *InteractionHandler) GetUserFavorites(c *gin.Context) {
	// 回去用户ID
	userID, _ := c.Get("user_id")

	// 执行获取该用户收藏列表的服务
	videos, err := h.interService.GetUserFavorites(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 返回正确响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": videos,
	})
}
