package model

import (
	"time"
)

// Comment 评论模型
type Comment struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Content  string `gorm:"type:text;not null" json:"content"`   // 评论内容
	UserID   uint   `gorm:"index;not null" json:"user_id"`       // 评论用户ID
	VideoID  uint   `gorm:"index;not null" json:"video_id"`      // 所属视频ID
	ParentID *uint  `gorm:"index;default:null" json:"parent_id"` // 父评论ID（支持回复）

	// 关联（不存数据库，仅用于查询）
	User    User      `gorm:"foreignKey:UserID" json:"user"`      // 评论用户信息，告诉 GORM 用 Comment 的 UserID 字段去关联 User 表
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies"` // 子评论列表，告诉 GORM 用 Comment 的 ParentID 字段去关联 []Comment 切片
}
