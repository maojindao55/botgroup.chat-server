package services

import (
	"time"

	"project/models"
	"project/repository"
)

// ChatService 聊天服务接口
type ChatService interface {
	ProcessMessage(message models.ChatMessage) (string, error)
	GetChatHistory(userID string) ([]models.ChatMessage, error)
}

// chatService 聊天服务实现
type chatService struct {
	repo repository.ChatRepository
}

// NewChatService 创建聊天服务实例
func NewChatService() ChatService {
	return &chatService{
		repo: repository.NewChatRepository(),
	}
}

// ProcessMessage 处理聊天消息
func (s *chatService) ProcessMessage(message models.ChatMessage) (string, error) {
	// 设置消息时间戳
	message.Timestamp = time.Now()

	// 保存消息到数据库
	if err := s.repo.SaveMessage(message); err != nil {
		return "", err
	}

	// 这里可以添加更复杂的消息处理逻辑，如调用AI服务等
	response := "收到您的消息: " + message.Content

	return response, nil
}

// GetChatHistory 获取聊天历史
func (s *chatService) GetChatHistory(userID string) ([]models.ChatMessage, error) {
	return s.repo.GetMessagesByUserID(userID)
}
