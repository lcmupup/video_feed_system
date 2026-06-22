package repository

import (
	"fmt"
	"log"
	"video_feed/internal/config"
	"video_feed/internal/model"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)
	var err error
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err) // 连接失败直接退出
	}
	// 自动迁移（建表）
	DB.AutoMigrate(&model.User{}, &model.Video{})
	log.Println("数据库连接成功，表已自动创建")
}
