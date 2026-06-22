package model

import (
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Username  string    `gorm:"uniqueIndex" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"` // 存储用户头像URL的字段
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (u User) TableName() string {
	return "users"
}
