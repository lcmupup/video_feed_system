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

// ========== 聊天消息队列 ==========

const (
	chatExchange   = "chat.exchange"
	chatQueue      = "chat.queue"
	chatRoutingKey = "chat.message"
)

// DeclareChatQueue 声明聊天相关的 Exchange 和 Queue
func DeclareChatQueue() error {
	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	// 声明 Direct Exchange
	if err := MQChannel.ExchangeDeclare(
		chatExchange, // name
		"direct",     // type
		true,         // durable（持久化）
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("声明 Exchange 失败: %w", err)
	}

	// 声明 Queue
	if _, err := MQChannel.QueueDeclare(
		chatQueue, // name
		true,      // durable（持久化）
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	); err != nil {
		return fmt.Errorf("声明 Queue 失败: %w", err)
	}

	// 绑定 Queue 到 Exchange
	if err := MQChannel.QueueBind(
		chatQueue,
		chatRoutingKey,
		chatExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("绑定 Queue 失败: %w", err)
	}

	log.Println("聊天消息队列声明成功")
	return nil
}

// PublishChatMessage 发布聊天消息到 RabbitMQ
func PublishChatMessage(body []byte) error {
	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	return MQChannel.Publish(
		chatExchange,   // exchange
		chatRoutingKey, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // 消息持久化
			Body:         body,
		},
	)
}

// ========== 视频上传队列 ==========

const (
	videoUploadExchange   = "video.exchange"
	videoUploadQueue      = "video.upload.queue"
	videoUploadRoutingKey = "video.upload"
)

// DeclareVideoUploadQueue 声明视频上传相关的 Exchange 和 Queue
func DeclareVideoUploadQueue() error {
	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	if err := MQChannel.ExchangeDeclare(
		videoUploadExchange, "direct",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("声明视频 Exchange 失败: %w", err)
	}

	if _, err := MQChannel.QueueDeclare(
		videoUploadQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("声明视频 Queue 失败: %w", err)
	}

	if err := MQChannel.QueueBind(
		videoUploadQueue, videoUploadRoutingKey, videoUploadExchange, false, nil,
	); err != nil {
		return fmt.Errorf("绑定视频 Queue 失败: %w", err)
	}

	log.Println("视频上传队列声明成功")
	return nil
}

// PublishVideoUpload 发布视频上传消息到 RabbitMQ
func PublishVideoUpload(body []byte) error {
	if MQChannel == nil {
		return fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	return MQChannel.Publish(
		videoUploadExchange,
		videoUploadRoutingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// ConsumeVideoUploads 消费视频上传消息
func ConsumeVideoUploads() (<-chan amqp.Delivery, error) {
	if MQChannel == nil {
		return nil, fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	return MQChannel.Consume(
		videoUploadQueue,
		"",
		false, // 手动ACK
		false, false, false, nil,
	)
}

// ConsumeChatMessages 消费聊天消息，返回一个只读 channel
func ConsumeChatMessages() (<-chan amqp.Delivery, error) {
	if MQChannel == nil {
		return nil, fmt.Errorf("RabbitMQ Channel 未初始化")
	}

	return MQChannel.Consume(
		chatQueue, // queue
		"",        // consumer tag（空表示自动生成）
		false,     // auto-ack（手动ACK，确保消息不丢）
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
}
