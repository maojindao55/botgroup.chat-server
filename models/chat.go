package models

import (
	"time"
)

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatResponse 聊天响应模型
type ChatResponse struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MessageID uint      `json:"message_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
