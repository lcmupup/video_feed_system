package model

import (
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	Username string `gorm:"uniqueIndex" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`                  // 存储用户头像URL的字段
	Bio      string `gorm:"size:500" json:"bio"`     // 个人简介
	Gender   int    `gorm:"default:0" json:"gender"` // 性别（0-未知,1-男,2-女）
	Birthday string `gorm:"size:10" json:"birthday"` // 生日
}

// TableName 自定义表名
func (u User) TableName() string {
	return "users"
}
