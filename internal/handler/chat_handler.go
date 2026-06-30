package handler

import (
	"strconv"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天 HTTP 处理器
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器实例
func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		chatService: service.NewChatService(),
	}
}

// SendMessage 发送消息
// POST /api/chat/send
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req service.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.chatService.SendMessage(userID.(uint), &req); err != nil {
		errcode.Error(c, errcode.CodeBadRequest, err.Error())
		return
	}

	errcode.Success(c, "消息已发送")
}

// GetHistory 获取与某个用户的聊天记录
// GET /api/chat/history/:user_id?page=1&page_size=20
func (h *ChatHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	peerID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		errcode.Error(c, errcode.CodeBadRequest, "无效的用户ID")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	messages, total, err := h.chatService.GetHistory(userID.(uint), uint(peerID), page, pageSize)
	if err != nil {
		errcode.Error(c, errcode.CodeInternalError, err.Error())
		return
	}

	errcode.SuccessWithPagination(c, "获取成功", messages, total, page, pageSize)
}

// GetConversations 获取会话列表
// GET /api/chat/conversations
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID, _ := c.Get("user_id")

	conversations, err := h.chatService.GetConversations(userID.(uint))
	if err != nil {
		errcode.Error(c, errcode.CodeInternalError, err.Error())
		return
	}

	errcode.SuccessWithData(c, "获取成功", conversations)
}
