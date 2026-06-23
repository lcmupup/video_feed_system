package handler

//封装用户的路由函数
import (
	"net/http"
	"path/filepath"
	"video_feed/internal/service"

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
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"code": 409,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "注册成功",
		"data": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
		},
	})
}

// Login 用户登录接口
// POST /api/user/login
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": resp.Token,
			"user": gin.H{
				"user_id":  resp.User.ID,
				"username": resp.User.Username,
				"nickname": resp.User.Nickname,
			},
		},
	})
}

// GetProfile 获取当前登录用户的简介信息（需要Token）
// GET /api/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从JWT中间件中获取用户ID
	userID, _ := c.Get("user_id")

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
		},
	})
}

// UpdateProfile 更新用户个人信息
// PUT /api/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 1. 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 2. 解析请求参数
	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}
	req.UserID = userID.(uint)

	// 3. 更新个人信息
	userInfo, err := h.userService.UpdateProfile(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": userInfo,
	})
}

// UpdateAvatar 更新用户头像
// POST /api/user/avatar
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	// 1. 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 2. 获取上传的头像文件
	file, err := c.FormFile("avatar") // 从前端传来的form表单中取出key为avatar的文件字段
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请选择头像文件",
		})
		return
	}

	// 3. 验证文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "仅支持 JPG/JPEG/PNG/GIF 格式的图片",
		})
		return
	}

	// 4. 保存头像
	avatarURL, err := h.userService.UpdateAvatar(userID.(uint), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "头像上传失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "头像更新成功",
		"data": gin.H{
			"avatar": avatarURL,
		},
	})
}
