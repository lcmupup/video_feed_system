package model

import (
	"time"
)

// Video 视频模型
type Video struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"` // json:"-"表示转成json时忽略这个字段

	Title       string `gorm:"size:100;not null" json:"title"`     // 视频标题
	Description string `gorm:"size:500" json:"description"`        // 视频描述
	FilePath    string `gorm:"size:255;not null" json:"file_path"` // 视频文件存储路径
	CoverPath   string `gorm:"size:255" json:"cover_path"`         // 封面图路径
	Duration    int64  `gorm:"default:0" json:"duration"`          // 视频时长（秒）
	FileSize    int64  `gorm:"default:0" json:"file_size"`         // 文件大小（字节）
	UserID      uint   `gorm:"index;not null" json:"user_id"`      // 上传用户ID
	ViewCount   int64  `gorm:"default:0" json:"view_count"`        // 观看次数
	LikeCount   int64  `gorm:"default:0" json:"like_count"`        // 点赞次数
	Status      int    `gorm:"default:0" json:"status"`            // 状态：0-待审核，1-已发布，2-下架
}
