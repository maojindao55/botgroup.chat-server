package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project/src/config"
	"project/src/constants"
	"project/src/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	upgrader websocket.Upgrader
	clients  map[string]*WSClient
	mutex    sync.RWMutex
}

// WSClient WebSocket客户端
type WSClient struct {
	ID         string
	SessionID  string
	Connection *websocket.Conn
	Send       chan []byte
	Service    *WebSocketService
}

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewWebSocketService 创建WebSocket服务实例
func NewWebSocketService() *WebSocketService {
	wsConfig := config.AppConfig.WebSocket

	return &WebSocketService{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  wsConfig.ReadBufferSize,
			WriteBufferSize: wsConfig.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				// 在生产环境中应该检查Origin
				return wsConfig.CheckOrigin
			},
		},
		clients: make(map[string]*WSClient),
	}
}

// HandleWebSocket 处理WebSocket连接升级
func (s *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取sessionID
	sessionID := r.URL.Path[len("/ws/auth/"):]
	if sessionID == "" {
		http.Error(w, "缺少会话ID", http.StatusBadRequest)
		return
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &WSClient{
		ID:         fmt.Sprintf("%s_%d", sessionID, time.Now().UnixNano()),
		SessionID:  sessionID,
		Connection: conn,
		Send:       make(chan []byte, 256),
		Service:    s,
	}

	// 注册客户端
	s.RegisterClient(client)

	// 启动客户端的读写goroutine
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket客户端连接成功: %s", client.ID)
}

// RegisterClient 注册客户端
func (s *WebSocketService) RegisterClient(client *WSClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[client.ID] = client
	log.Printf("注册WebSocket客户端: %s, 当前连接数: %d", client.ID, len(s.clients))
}

// UnregisterClient 注销客户端
func (s *WebSocketService) UnregisterClient(client *WSClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.clients[client.ID]; ok {
		close(client.Send)
		delete(s.clients, client.ID)
		log.Printf("注销WebSocket客户端: %s, 当前连接数: %d", client.ID, len(s.clients))
	}
}

// BroadcastToSession 向特定会话发送消息
func (s *WebSocketService) BroadcastToSession(sessionID string, message interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	sent := false
	for _, client := range s.clients {
		if client.SessionID == sessionID {
			select {
			case client.Send <- data:
				sent = true
			default:
				// 客户端发送缓冲区满，关闭连接
				s.UnregisterClient(client)
				client.Connection.Close()
			}
		}
	}

	if !sent {
		return fmt.Errorf("未找到会话ID对应的客户端: %s", sessionID)
	}

	return nil
}

// BroadcastToAll 向所有客户端发送消息
func (s *WebSocketService) BroadcastToAll(message interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	for _, client := range s.clients {
		select {
		case client.Send <- data:
		default:
			// 客户端发送缓冲区满，关闭连接
			s.UnregisterClient(client)
			client.Connection.Close()
		}
	}

	return nil
}

// GetClientCount 获取当前连接的客户端数量
func (s *WebSocketService) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.clients)
}

// GetClientsBySession 获取特定会话的客户端数量
func (s *WebSocketService) GetClientsBySession(sessionID string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	for _, client := range s.clients {
		if client.SessionID == sessionID {
			count++
		}
	}

	return count
}

// SendLoginResult 发送登录结果
func (s *WebSocketService) SendLoginResult(sessionID string, result *models.WebSocketLoginResult) error {
	return s.BroadcastToSession(sessionID, result)
}

// WritePump 处理向WebSocket连接写入消息
func (c *WSClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 服务关闭了发送通道
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送缓冲区中的额外消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 处理从WebSocket连接读取消息
func (c *WSClient) ReadPump() {
	defer func() {
		c.Service.UnregisterClient(c)
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(512)
	c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接异常关闭: %v", err)
			}
			break
		}

		// 处理客户端发送的消息
		c.handleMessage(message)
	}
}

// handleMessage 处理客户端消息
func (c *WSClient) handleMessage(message []byte) {
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("解析WebSocket消息失败: %v", err)
		return
	}

	switch wsMsg.Type {
	case "ping":
		// 回复pong
		response := WSMessage{
			Type: "pong",
			Data: "pong",
		}
		data, _ := json.Marshal(response)
		select {
		case c.Send <- data:
		default:
			// 发送缓冲区满
		}
	case "status_check":
		// 检查会话状态
		response := WSMessage{
			Type: "status_response",
			Data: map[string]interface{}{
				"session_id": c.SessionID,
				"connected":  true,
				"timestamp":  time.Now().Unix(),
			},
		}
		data, _ := json.Marshal(response)
		select {
		case c.Send <- data:
		default:
			// 发送缓冲区满
		}
	default:
		log.Printf("未知的WebSocket消息类型: %s", wsMsg.Type)
	}
}

// NotifyLoginSuccess 通知登录成功
func (s *WebSocketService) NotifyLoginSuccess(sessionID string, userInfo *models.UserInfo, token string, expiresIn int) error {
	result := &models.WebSocketLoginResult{
		Type: constants.WSMessageTypeLoginResult,
		Data: &models.WebSocketLoginData{
			Status:    constants.SessionStatusSuccess,
			Message:   "登录成功",
			UserInfo:  userInfo,
			Token:     token,
			ExpiresIn: expiresIn,
		},
	}

	return s.SendLoginResult(sessionID, result)
}

// NotifyLoginFailed 通知登录失败
func (s *WebSocketService) NotifyLoginFailed(sessionID string, message string) error {
	result := &models.WebSocketLoginResult{
		Type: constants.WSMessageTypeLoginResult,
		Data: &models.WebSocketLoginData{
			Status:    "failed",
			Message:   message,
			UserInfo:  nil,
			Token:     "",
			ExpiresIn: 0,
		},
	}

	return s.SendLoginResult(sessionID, result)
}

// NotifyLoginExpired 通知登录过期
func (s *WebSocketService) NotifyLoginExpired(sessionID string) error {
	result := &models.WebSocketLoginResult{
		Type: constants.WSMessageTypeLoginResult,
		Data: &models.WebSocketLoginData{
			Status:    constants.SessionStatusExpired,
			Message:   "登录会话已过期，请重新扫码",
			UserInfo:  nil,
			Token:     "",
			ExpiresIn: 0,
		},
	}

	return s.SendLoginResult(sessionID, result)
}

// CleanupExpiredConnections 清理过期连接
func (s *WebSocketService) CleanupExpiredConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 简单的心跳检测，向所有客户端发送ping
	for _, client := range s.clients {
		select {
		case client.Send <- []byte(`{"type":"ping","data":"heartbeat"}`):
		default:
			// 无法发送，说明连接可能有问题
			delete(s.clients, client.ID)
			client.Connection.Close()
		}
	}
}

// StartCleanupRoutine 启动清理例程
func (s *WebSocketService) StartCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			s.CleanupExpiredConnections()
		}
	}()
}
