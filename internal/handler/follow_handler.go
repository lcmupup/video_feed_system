package handler

import (
	"strconv"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// FollowHandler 关注HTTP处理器
type FollowHandler struct {
	followService *service.FollowService
}

// NewFollowHandler 创建关注处理器实例
func NewFollowHandler() *FollowHandler {
	return &FollowHandler{
		followService: service.NewFollowService(),
	}
}

// FollowUser 关注用户
// POST /api/user/:id/follow
func (h *FollowHandler) FollowUser(c *gin.Context) {
	followerID, _ := c.Get("user_id")
	followingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的用户ID")
		return
	}

	err = h.followService.FollowUser(followerID.(uint), uint(followingID))
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.Success(c, "关注成功")
}

// UnfollowUser 取消关注
// DELETE /api/user/:id/follow
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	followerID, _ := c.Get("user_id")
	followingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的用户ID")
		return
	}

	err = h.followService.UnfollowUser(followerID.(uint), uint(followingID))
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.Success(c, "取消关注成功")
}

// GetFollowers 获取粉丝列表
// GET /api/user/:id/followers?page=1&page_size=10
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的用户ID")
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

	users, total, err := h.followService.GetFollowers(uint(userID), page, pageSize)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "获取成功", users, total, page, pageSize)
}

// GetFollowings 获取关注列表
// GET /api/user/:id/followings?page=1&page_size=10
func (h *FollowHandler) GetFollowings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的用户ID")
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

	users, total, err := h.followService.GetFollowings(uint(userID), page, pageSize)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "获取成功", users, total, page, pageSize)
}
