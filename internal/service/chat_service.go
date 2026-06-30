package service

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"video_feed/internal/model"
	"video_feed/internal/repository"
)

// ChatService 聊天业务逻辑层
type ChatService struct {
	msgRepo  *repository.MessageRepository
	userRepo *repository.UserRepository
}

// NewChatService 创建聊天服务实例
func NewChatService() *ChatService {
	return &ChatService{
		msgRepo:  repository.NewMessageRepository(),
		userRepo: repository.NewUserRepository(),
	}
}

// SendMessageReq 发送消息请求
type SendMessageReq struct {
	ToUserID uint   `json:"to_user_id" binding:"required"`
	Content  string `json:"content" binding:"required,max=500"`
}

// ChatMessageBody RabbitMQ 消息体
type ChatMessageBody struct {
	FromUserID uint   `json:"from_user_id"`
	ToUserID   uint   `json:"to_user_id"`
	Content    string `json:"content"`
	Timestamp  string `json:"timestamp"`
}

// MessageInfo 消息响应
type MessageInfo struct {
	ID         uint   `json:"id"`
	FromUserID uint   `json:"from_user_id"`
	ToUserID   uint   `json:"to_user_id"`
	Content    string `json:"content"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
	FromUser   *SimpleUserInfo `json:"from_user,omitempty"`
}

// SimpleUserInfo 简化的用户信息
type SimpleUserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// SendMessage 发送消息：组装消息体 → 发布到 RabbitMQ
func (s *ChatService) SendMessage(fromUserID uint, req *SendMessageReq) error {
	// 1. 检查接收方是否存在
	if _, err := s.userRepo.FindByID(req.ToUserID); err != nil {
		return errors.New("接收用户不存在")
	}

	// 2. 不能给自己发消息
	if fromUserID == req.ToUserID {
		return errors.New("不能给自己发消息")
	}

	// 3. 组装消息体
	body := ChatMessageBody{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errors.New("消息序列化失败")
	}

	// 4. 发布到 RabbitMQ（异步，立刻返回）
	if err := repository.PublishChatMessage(jsonBody); err != nil {
		return errors.New("发送消息失败: " + err.Error())
	}

	return nil
}

// GetHistory 获取两个用户之间的聊天记录
func (s *ChatService) GetHistory(user1, user2 uint, page, pageSize int) ([]MessageInfo, int64, error) {
	messages, total, err := s.msgRepo.GetMessages(user1, user2, page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取聊天记录失败")
	}

	// 标记对方发来的消息为已读
	s.msgRepo.MarkAsRead(user2, user1)

	var infos []MessageInfo
	for _, m := range messages {
		info := MessageInfo{
			ID:         m.ID,
			FromUserID: m.FromUserID,
			ToUserID:   m.ToUserID,
			Content:    m.Content,
			IsRead:     m.IsRead == 1,
			CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if m.FromUser.ID != 0 {
			info.FromUser = &SimpleUserInfo{
				ID:       m.FromUser.ID,
				Username: m.FromUser.Username,
				Nickname: m.FromUser.Nickname,
				Avatar:   m.FromUser.Avatar,
			}
		}
		infos = append(infos, info)
	}

	if infos == nil {
		infos = []MessageInfo{}
	}

	return infos, total, nil
}

// GetConversations 获取用户的会话列表
func (s *ChatService) GetConversations(userID uint) ([]repository.ConversationInfo, error) {
	conversations, err := s.msgRepo.GetConversations(userID)
	if err != nil {
		return nil, errors.New("获取会话列表失败")
	}
	return conversations, nil
}

// SaveMessage 将消息持久化到 MySQL（Consumer 调用）
func SaveMessage(body *ChatMessageBody) (*model.Message, error) {
	msg := &model.Message{
		FromUserID: body.FromUserID,
		ToUserID:   body.ToUserID,
		Content:    body.Content,
		IsRead:     0,
	}

	repo := repository.NewMessageRepository()
	if err := repo.Create(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// StartConsumer 启动 RabbitMQ 消费者（在 goroutine 中运行）
// wsHandler 用于推送消息给在线用户，如果为 nil 则不推送
func StartConsumer(onMessage func(body *ChatMessageBody) error) {
	if repository.MQChannel == nil {
		log.Println("RabbitMQ 未连接，跳过消费者启动")
		return
	}

	msgs, err := repository.ConsumeChatMessages()
	if err != nil {
		log.Printf("启动聊天消费者失败: %v", err)
		return
	}

	log.Println("聊天消息消费者已启动")

	for d := range msgs {
		var body ChatMessageBody
		if err := json.Unmarshal(d.Body, &body); err != nil {
			log.Printf("消息解析失败: %v", err)
			d.Nack(false, false) // 解析失败，丢弃消息
			continue
		}

		// 回调处理（落库 + WebSocket 推送）
		if err := onMessage(&body); err != nil {
			log.Printf("处理消息失败: %v", err)
			d.Nack(false, true) // 处理失败，重新入队
			continue
		}

		d.Ack(false) // 处理成功，确认消息
	}
}
