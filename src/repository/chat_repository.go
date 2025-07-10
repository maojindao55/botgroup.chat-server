package repository

import (
	"errors"

	"project/src/models"
)

// ChatRepository 聊天仓库接口
type ChatRepository interface {
	SaveMessage(message models.ChatMessage) error
	GetMessagesByUserID(userID string) ([]models.ChatMessage, error)
}

// chatRepository 聊天仓库实现
type chatRepository struct {
	// 这里可以添加数据库连接等
	messages []models.ChatMessage // 临时存储，实际应使用数据库
}

// NewChatRepository 创建聊天仓库实例
func NewChatRepository() ChatRepository {
	return &chatRepository{
		messages: make([]models.ChatMessage, 0),
	}
}

// SaveMessage 保存消息
func (r *chatRepository) SaveMessage(message models.ChatMessage) error {
	// 模拟ID自增
	message.ID = uint(len(r.messages) + 1)
	r.messages = append(r.messages, message)
	return nil
}

// GetMessagesByUserID 根据用户ID获取消息
func (r *chatRepository) GetMessagesByUserID(userID string) ([]models.ChatMessage, error) {
	var result []models.ChatMessage

	for _, msg := range r.messages {
		if msg.UserID == userID {
			result = append(result, msg)
		}
	}

	if len(result) == 0 {
		return nil, errors.New("未找到该用户的消息")
	}

	return result, nil
}
