package services

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"project/src/config"
	"project/src/models"

	openai "github.com/sashabaranov/go-openai"
)

// SchedulerService 调度服务接口
type SchedulerService interface {
	ScheduleAIResponses(message string, history []models.ChatMessage, availableAIs []*config.LLMCharacter) ([]string, error)
}

// schedulerService 调度服务实现
type schedulerService struct{}

// NewSchedulerService 创建调度服务实例
func NewSchedulerService() SchedulerService {
	return &schedulerService{}
}

// SchedulerRequest 调度请求结构体
type SchedulerRequest struct {
	Message      string                 `json:"message"`
	History      []models.ChatMessage   `json:"history"`
	AvailableAIs []*config.LLMCharacter `json:"available_ais"`
}

// SchedulerResponse 调度响应结构体
type SchedulerResponse struct {
	SelectedAIs []string `json:"selected_ais"`
}

// ScheduleAIResponses 调度AI响应
func (s *schedulerService) ScheduleAIResponses(message string, history []models.ChatMessage, availableAIs []*config.LLMCharacter) ([]string, error) {
	// 1. 收集所有可用的标签
	allTags := make(map[string]bool)
	for _, ai := range availableAIs {
		for _, tag := range ai.Tags {
			allTags[tag] = true
		}
	}

	// 将map转换为切片
	tagsList := make([]string, 0, len(allTags))
	for tag := range allTags {
		tagsList = append(tagsList, tag)
	}

	// 2. 使用AI模型分析消息并匹配标签
	matchedTags, err := s.analyzeMessageWithAI(message, tagsList, history)
	if err != nil {
		log.Printf("分析消息失败: %v", err)
		// 即使分析失败，我们也继续执行，只是没有标签匹配
	}

	log.Printf("匹配的标签: %v", matchedTags)

	// 如果含有"文字游戏"标签，则需要全员参与
	if containsTag(matchedTags, "文字游戏") {
		selectedAIs := make([]string, 0, len(availableAIs))
		for _, ai := range availableAIs {
			selectedAIs = append(selectedAIs, ai.ID)
		}
		return selectedAIs, nil
	}

	// 3. 计算每个AI的匹配分数
	aiScores := make(map[string]int)
	messageLower := strings.ToLower(message)

	for _, ai := range availableAIs {
		if len(ai.Tags) == 0 {
			continue
		}

		score := 0
		// 标签匹配分数
		for _, tag := range matchedTags {
			if containsTag(ai.Tags, tag) {
				score += 2 // 每个匹配的标签得2分
			}
		}

		// 直接提到AI名字额外加分
		if strings.Contains(messageLower, strings.ToLower(ai.Name)) {
			score += 5
		}

		// 历史对话相关性加分
		recentHistory := getRecentHistory(history, 5) // 只看最近5条消息
		for _, hist := range recentHistory {
			if hist.Name == ai.Name && len(hist.Content) > 0 {
				score += 1 // 最近有参与对话的AI加分
			}
		}

		if score > 0 {
			aiScores[ai.ID] = score
		}
	}

	// 4. 根据分数排序选择AI
	sortedAIs := sortAIsByScore(aiScores)

	// 5. 如果没有匹配到任何AI，随机选择1-2个
	if len(sortedAIs) == 0 {
		log.Println("没有匹配到任何AI，随机选择1-2个")
		maxResponders := min(2, len(availableAIs))
		numResponders := rand.Intn(maxResponders) + 1

		shuffledAIs := shuffleAIs(availableAIs)
		selectedAIs := make([]string, 0, numResponders)
		for i := 0; i < numResponders && i < len(shuffledAIs); i++ {
			selectedAIs = append(selectedAIs, shuffledAIs[i].ID)
		}

		return selectedAIs, nil
	}

	// 6. 限制最大回复数量
	const maxResponders = 9
	if len(sortedAIs) > maxResponders {
		sortedAIs = sortedAIs[:maxResponders]
	}

	return sortedAIs, nil
}

// analyzeMessageWithAI 使用AI分析消息并返回匹配的标签
func (s *schedulerService) analyzeMessageWithAI(message string, allTags []string, history []models.ChatMessage) ([]string, error) {
	// 获取调度器AI配置
	schedulerAIConfig := config.AppConfig.LLMCharacters[0]
	if schedulerAIConfig == nil || schedulerAIConfig.Personality != "sheduler" {
		return nil, errors.New("调度器AI配置未找到")
	}

	// 获取模型配置
	modelConfig := config.AppConfig.LLMModels[schedulerAIConfig.Model]
	if modelConfig == "" {
		return nil, errors.New("模型配置未找到")
	}

	// 获取API密钥和基础URL
	provider := modelConfig
	apiKey := config.AppConfig.LLMProviders[provider].APIKey
	baseURL := config.AppConfig.LLMProviders[provider].BaseURL
	if apiKey == "" || baseURL == "" {
		return nil, errors.New("API密钥或基础URL未配置")
	}

	// 创建OpenAI客户端
	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.BaseURL = baseURL
	client := openai.NewClientWithConfig(openaiConfig)

	// 构建提示词
	prompt := schedulerAIConfig.CustomPrompt
	tagsStr := strings.Join(allTags, ", ")
	prompt = strings.ReplaceAll(prompt, "#allTags#", tagsStr)

	// 构建消息数组
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
	}

	// 添加历史消息，最多取最近10条
	historyLimit := 10
	if len(history) > historyLimit {
		history = history[len(history)-historyLimit:]
	}

	for _, msg := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		})
	}

	// 添加当前用户消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	// 创建上下文
	ctx := context.Background()

	// 发送请求
	completion, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    schedulerAIConfig.Model,
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	// 解析响应
	if len(completion.Choices) == 0 {
		return []string{}, nil
	}

	content := completion.Choices[0].Message.Content
	matchedTags := strings.Split(content, ",")
	for i, tag := range matchedTags {
		matchedTags[i] = strings.TrimSpace(tag)
	}

	return matchedTags, nil
}

// 辅助函数

// containsTag 检查标签列表是否包含特定标签
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// getRecentHistory 获取最近的历史消息
func getRecentHistory(history []models.ChatMessage, limit int) []models.ChatMessage {
	if len(history) <= limit {
		return history
	}
	return history[len(history)-limit:]
}

// sortAIsByScore 根据分数排序AI
func sortAIsByScore(aiScores map[string]int) []string {
	type aiScore struct {
		id    string
		score int
	}

	scores := make([]aiScore, 0, len(aiScores))
	for id, score := range aiScores {
		scores = append(scores, aiScore{id: id, score: score})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]string, len(scores))
	for i, s := range scores {
		result[i] = s.id
	}

	return result
}

// shuffleAIs 随机打乱AI列表
func shuffleAIs(ais []*config.LLMCharacter) []*config.LLMCharacter {
	result := make([]*config.LLMCharacter, len(ais))
	copy(result, ais)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
