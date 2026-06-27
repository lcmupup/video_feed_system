package model

import (
	"time"
)

// Follow 关注关系模型
type Follow struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	FollowerID  uint `gorm:"index:idx_follower_following,unique;not null" json:"follower_id"`  // 关注者ID
	FollowingID uint `gorm:"index:idx_follower_following,unique;not null" json:"following_id"` // 被关注者ID
}
