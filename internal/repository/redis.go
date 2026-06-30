package repository

import (
	"context"
	"fmt"
	"log"

	"video_feed/internal/config"

	"github.com/redis/go-redis/v9"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.RedisConfig) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("⚠ Redis 连接失败（服务仍可运行，排行榜等功能将不可用）: %v", err)
		return
	}

	log.Println("Redis 连接成功")
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() {
	if RDB != nil {
		if err := RDB.Close(); err != nil {
			log.Printf("关闭 Redis 连接失败: %v", err)
		}
		log.Println("Redis 连接已关闭")
	}
}
