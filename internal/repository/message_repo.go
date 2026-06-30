package repository

import (
	"video_feed/internal/model"
)

// MessageRepository 聊天消息数据访问层
type MessageRepository struct{}

// NewMessageRepository 创建消息仓库实例
func NewMessageRepository() *MessageRepository {
	return &MessageRepository{}
}

// Create 保存消息
func (r *MessageRepository) Create(msg *model.Message) error {
	return DB.Create(msg).Error
}

// GetMessages 获取两个用户之间的聊天记录（分页，按时间倒序）
func (r *MessageRepository) GetMessages(user1, user2 uint, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := DB.Model(&model.Message{}).
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			user1, user2, user2, user1)

	query.Count(&total)

	err := query.
		Preload("FromUser").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&messages).Error

	return messages, total, err
}

// ConversationInfo 会话信息
type ConversationInfo struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	LastMessage string `json:"last_message"`
	LastTime    string `json:"last_time"`
	UnreadCount int64  `json:"unread_count"`
}

// GetConversations 获取用户的会话列表（最近联系人 + 最后一条消息 + 未读数）
func (r *MessageRepository) GetConversations(userID uint) ([]ConversationInfo, error) {
	// 子查询：找到每个会话的最后一条消息ID
	// 策略：分别查"我发的"和"我收的"，去重后取每个对方用户的最后一条消息
	rows, err := DB.Raw(`
		SELECT
			other_user_id,
			u.username,
			u.nickname,
			u.avatar,
			m.content AS last_message,
			m.created_at AS last_time,
			(SELECT COUNT(*) FROM messages
			 WHERE to_user_id = ? AND from_user_id = other_user_id AND is_read = 0
			) AS unread_count
		FROM (
			SELECT
				CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END AS other_user_id,
				MAX(id) AS max_id
			FROM messages
			WHERE from_user_id = ? OR to_user_id = ?
			GROUP BY other_user_id
		) AS latest
		JOIN messages m ON m.id = latest.max_id
		JOIN users u ON u.id = latest.other_user_id
		ORDER BY m.created_at DESC
	`, userID, userID, userID, userID).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationInfo
	for rows.Next() {
		var c ConversationInfo
		var lastTime interface{}
		if err := rows.Scan(&c.UserID, &c.Username, &c.Nickname, &c.Avatar,
			&c.LastMessage, &lastTime, &c.UnreadCount); err != nil {
			continue
		}
		// 格式化时间
		if t, ok := lastTime.([]byte); ok {
			c.LastTime = string(t)
		}
		conversations = append(conversations, c)
	}

	if conversations == nil {
		conversations = []ConversationInfo{}
	}

	return conversations, nil
}

// MarkAsRead 标记消息为已读
func (r *MessageRepository) MarkAsRead(fromUserID, toUserID uint) error {
	return DB.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", fromUserID, toUserID, 0).
		Update("is_read", true).Error
}
