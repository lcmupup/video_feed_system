package model

import (
	"time"
)

// Message 聊天消息模型（无软删除，用硬删除）
type Message struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	FromUserID uint   `gorm:"index;not null" json:"from_user_id"`
	ToUserID   uint   `gorm:"index;not null" json:"to_user_id"`
	Content    string `gorm:"type:text;not null" json:"content"`
	IsRead     int    `gorm:"default:0" json:"is_read"` // 0=未读, 1=已读（用 int 避免 GORM v1 对 bool 的 AutoMigrate bug）

	// 关联（GORM v1 通过命名约定自动关联，不显式声明 foreignKey 避免 AutoMigrate 死循环）
	FromUser User `json:"from_user,omitempty"`
	ToUser   User `json:"to_user,omitempty"`
}
