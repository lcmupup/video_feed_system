package repository

import (
	"fmt"
	"log"

	"video_feed/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MQConn 全局 RabbitMQ 连接
var MQConn *amqp.Connection

// MQChannel 全局 RabbitMQ Channel
var MQChannel *amqp.Channel

// InitRabbitMQ 初始化 RabbitMQ 连接
func InitRabbitMQ(cfg *config.RabbitMQConfig) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	var err error
	MQConn, err = amqp.Dial(dsn)
	if err != nil {
		log.Printf("⚠ RabbitMQ 连接失败（服务仍可运行，聊天等功能将不可用）: %v", err)
		return
	}

	MQChannel, err = MQConn.Channel()
	if err != nil {
		log.Printf("⚠ RabbitMQ 创建 Channel 失败: %v", err)
		return
	}

	log.Println("RabbitMQ 连接成功")
}

// CloseRabbitMQ 关闭 RabbitMQ 连接
func CloseRabbitMQ() {
	if MQChannel != nil {
		if err := MQChannel.Close(); err != nil {
			log.Printf("关闭 RabbitMQ Channel 失败: %v", err)
		}
	}
	if MQConn != nil {
		if err := MQConn.Close(); err != nil {
			log.Printf("关闭 RabbitMQ 连接失败: %v", err)
		}
	}
	log.Println("RabbitMQ 连接已关闭")
}
