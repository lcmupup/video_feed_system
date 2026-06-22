package handler

//封装用户的路由函数
import (
	"net/http"
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
