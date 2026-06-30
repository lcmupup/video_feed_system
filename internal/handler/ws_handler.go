package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"video_feed/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSHandler WebSocket 连接管理器
type WSHandler struct {
	onlineUsers map[uint]*websocket.Conn
	mu          sync.RWMutex
}

// DefaultWSHandler 全局 WebSocket 管理器（供 Consumer 推送消息使用）
var DefaultWSHandler *WSHandler

// NewWSHandler 创建 WebSocket 管理器
func NewWSHandler() *WSHandler {
	ws := &WSHandler{
		onlineUsers: make(map[uint]*websocket.Conn),
	}
	DefaultWSHandler = ws
	return ws
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（开发环境）
	},
}

// HandleWebSocket 处理 WebSocket 连接
// GET /api/ws/chat?token=xxx
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	// 1. 从 Query 参数中获取 Token 并校验
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "缺少Token"})
		return
	}

	claims, err := middleware.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token无效或已过期"})
		return
	}

	userID := claims.UserID

	// 2. 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败 (user_id=%d): %v", userID, err)
		return
	}

	// 3. 注册在线用户
	h.mu.Lock()
	// 如果该用户之前已在线，关闭旧连接
	if oldConn, ok := h.onlineUsers[userID]; ok {
		oldConn.Close()
	}
	h.onlineUsers[userID] = conn
	h.mu.Unlock()

	log.Printf("用户 %d 已上线 (当前在线: %d)", userID, len(h.onlineUsers))

	// 4. 读取循环（处理心跳和连接关闭）
	go h.readPump(conn, userID)
}

// readPump 读取 WebSocket 消息（处理 ping/pong 和连接关闭）
func (h *WSHandler) readPump(conn *websocket.Conn, userID uint) {
	defer func() {
		h.mu.Lock()
		delete(h.onlineUsers, userID)
		h.mu.Unlock()
		conn.Close()
		log.Printf("用户 %d 已下线 (当前在线: %d)", userID, len(h.onlineUsers))
	}()

	// 设置读超时（60秒无消息则断开）
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// 收到 pong 消息时刷新读超时
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 每30秒发送一次 ping 保活
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket 异常关闭 (user_id=%d): %v", userID, err)
			}
			break
		}
	}
}

// SendToUser 向指定用户推送消息（Consumer 调用）
func (h *WSHandler) SendToUser(userID uint, message interface{}) error {
	h.mu.RLock()
	conn, ok := h.onlineUsers[userID]
	h.mu.RUnlock()

	if !ok {
		return nil // 用户不在线，消息已落库，上线后可通过历史记录查看
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, jsonData)
}
