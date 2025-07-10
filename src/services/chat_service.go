package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"project/src/config"
	"project/src/models"
	"project/src/repository"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// ChatService 聊天服务接口
type ChatService interface {
	ProcessMessageStream(message models.ChatMessage, req ChatRequest, writer http.ResponseWriter) error
	GetChatHistory(userID string) ([]models.ChatMessage, error)
}

// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Message          string               `json:"message"`
	UserID           string               `json:"user_id"`
	Model            string               `json:"model"`
	CustomPrompt     string               `json:"custom_prompt"`
	AIName           string               `json:"aiName"`
	History          []models.ChatMessage `json:"history"`
	Index            int                  `json:"index"`
	ResponseCallback func(string)
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

// ProcessMessageStream 处理聊天消息并以流式方式返回响应
func (s *chatService) ProcessMessageStream(message models.ChatMessage, req ChatRequest, writer http.ResponseWriter) error {
	// 设置消息时间戳
	message.Timestamp = time.Now()

	//根据res.model选择不同的模型的apikey， 需要判断是否存在
	provider := config.AppConfig.LLMModels[req.Model]
	if provider == "" {
		return errors.New("model not found:" + req.Model)
	}
	apiKey := config.AppConfig.LLMProviders[provider].APIKey
	baseURL := config.AppConfig.LLMProviders[provider].BaseURL
	if apiKey == "" || baseURL == "" {
		return errors.New("api key is empty, model:" + req.Model)
	}

	// 创建 sashabaranov/go-openai 客户端
	aconfig := openai.DefaultConfig(apiKey)
	aconfig.BaseURL = baseURL
	client := openai.NewClientWithConfig(aconfig)

	systemPrompt := ""
	if req.CustomPrompt != "" {
		//替换#name#
		llmSystemPrompt := strings.Replace(config.AppConfig.LLMSystemPrompt, "#name#", req.AIName, -1)
		systemPrompt = req.CustomPrompt + "\n" + llmSystemPrompt
	}

	// 构建消息数组
	var chatMessages []openai.ChatCompletionMessage

	// 添加系统消息
	if systemPrompt != "" {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	history := req.History
	// 添加历史消息，最多取最近10条
	historyLimit := 10
	if len(history) > historyLimit {
		history = history[len(history)-historyLimit:]
	}

	for _, msg := range history {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		})
	}

	userMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Content,
	}

	// 添加当前用户消息
	if req.Index == 0 {
		chatMessages = append(chatMessages, userMessage)
	} else {
		// 在历史消息的倒数第index位置插入用户消息
		insertPosition := len(chatMessages) - req.Index
		if insertPosition >= 0 {
			// 创建一个新的消息数组
			newChatMessages := make([]openai.ChatCompletionMessage, 0, len(chatMessages)+1)
			// 复制前面的消息
			newChatMessages = append(newChatMessages, chatMessages[:insertPosition]...)
			// 插入用户消息
			newChatMessages = append(newChatMessages, userMessage)
			// 添加剩余的消息
			newChatMessages = append(newChatMessages, chatMessages[insertPosition:]...)
			chatMessages = newChatMessages
		} else {
			// 如果索引超出范围，直接添加到消息列表开头
			chatMessages = append([]openai.ChatCompletionMessage{userMessage}, chatMessages...)
		}
	}

	// 创建上下文
	ctx := context.Background()

	// 创建流式聊天完成请求
	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: chatMessages,
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		log.Println("error:", err, "model:", req.Model)
		return err
	}
	defer stream.Close()

	// 处理流式响应并直接发送到客户端
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				// 构造 SSE 事件
				data := fmt.Sprintf("data: %s\n\n",
					fmt.Sprintf("{\"content\": %q}", content))

				// 发送到客户端
				_, err := writer.Write([]byte(data))
				if err != nil {
					return err
				}
				writer.(http.Flusher).Flush()
			}
		}
	}

	return nil
}

// GetChatHistory 获取聊天历史
func (s *chatService) GetChatHistory(userID string) ([]models.ChatMessage, error) {
	return s.repo.GetMessagesByUserID(userID)
}
