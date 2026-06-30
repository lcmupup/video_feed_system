package handler

import (
	"path/filepath"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// Register 用户注册接口
// POST /api/user/register
func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		errcode.Error(c, errcode.CodeConflict, err.Error())
		return
	}

	errcode.SuccessWithData(c, "注册成功", gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

// Login 用户登录接口
// POST /api/user/login
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		errcode.Error(c, errcode.CodeUnauthorized, err.Error())
		return
	}

	errcode.SuccessWithData(c, "登录成功", gin.H{
		"token": resp.Token,
		"user": gin.H{
			"user_id":  resp.User.ID,
			"username": resp.User.Username,
			"nickname": resp.User.Nickname,
		},
	})
}

// GetProfile 获取当前登录用户的简介信息（需要Token）
// GET /api/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		errcode.Error(c, errcode.CodeNotFound, "用户不存在")
		return
	}

	errcode.SuccessWithData(c, "获取成功", gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})
}

// UpdateProfile 更新用户个人信息
// PUT /api/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errcode.Error(c, errcode.CodeUnauthorized, "请先登录")
		return
	}

	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	req.UserID = userID.(uint)

	userInfo, err := h.userService.UpdateProfile(&req)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.SuccessWithData(c, "更新成功", userInfo)
}

// UpdateAvatar 更新用户头像
// POST /api/user/avatar
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errcode.Error(c, errcode.CodeUnauthorized, "请先登录")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "请选择头像文件")
		return
	}

	// 验证文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		errcode.Error(c, errcode.CodeBadRequest, "仅支持 JPG/JPEG/PNG/GIF 格式的图片")
		return
	}

	avatarURL, err := h.userService.UpdateAvatar(userID.(uint), file)
	if err != nil {
		errcode.Error(c, errcode.CodeInternalError, "头像上传失败: "+err.Error())
		return
	}

	errcode.SuccessWithData(c, "头像更新成功", gin.H{
		"avatar": avatarURL,
	})
}
