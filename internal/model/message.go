package model

import (
	"time"
)

// Message 聊天消息模型
type Message struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`

	FromUserID uint   `gorm:"index:idx_conversation;not null" json:"from_user_id"` // 发送者ID
	ToUserID   uint   `gorm:"index:idx_conversation;index:idx_to_user;not null" json:"to_user_id"` // 接收者ID
	Content    string `gorm:"type:text;not null" json:"content"`                   // 消息内容
	IsRead     bool   `gorm:"default:false" json:"is_read"`                        // 是否已读

	// 关联（不存数据库，仅用于 Preload 查询）
	FromUser User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}
