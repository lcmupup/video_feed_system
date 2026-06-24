package model

import (
	"time"
)

// Like 点赞记录
type Like struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	// 为 user_id 和 video_id 这两个字段创建一个联合唯一索引，名字叫 idx_user_video，确保同一个用户不能重复点赞同一视频
	UserID  uint `gorm:"unique_index:idx_user_video;not null" json:"user_id"`  // 点赞用户ID
	VideoID uint `gorm:"unique_index:idx_user_video;not null" json:"video_id"` // 被点赞视频ID
}

// Favorite 收藏记录
type Favorite struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	// 为 user_id 和 video_id 这两个字段创建一个联合唯一索引，名字叫 idx_user_fav，确保同一个用户不能重复收藏同一视频
	UserID  uint `gorm:"unique_index:idx_user_fav;not null" json:"user_id"`  // 收藏用户ID
	VideoID uint `gorm:"unique_index:idx_user_fav;not null" json:"video_id"` // 被收藏视频ID
}
