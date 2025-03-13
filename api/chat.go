package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project/models"
	"project/services"
)

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

// ChatHandler 处理聊天请求
func ChatHandler(c *gin.Context) {
	var req ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	// 调用服务层处理业务逻辑
	chatService := services.NewChatService()
	response, err := chatService.ProcessMessage(models.ChatMessage{
		UserID:  req.UserID,
		Content: req.Message,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理消息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": response,
	})
}
