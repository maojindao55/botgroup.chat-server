package api

import (
	"fmt"
	"net/http"
	"project/models"
	"project/services"

	"github.com/gin-gonic/gin"
)

// ChatHandler 处理聊天请求，仅支持流式输出
func ChatHandler(c *gin.Context) {
	var req services.ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Flush()

	// 调用服务层处理流式业务逻辑
	chatService := services.NewChatService()
	err := chatService.ProcessMessageStream(models.ChatMessage{
		UserID:  req.UserID,
		Content: req.Message,
	}, req, c.Writer)

	if err != nil {
		fmt.Println("处理流式消息失败:", err)
		// 发送错误事件
		c.Writer.Write([]byte(fmt.Sprintf("data: {\"error\": \"%s\"}\n\n", err.Error())))
		c.Writer.Flush()
	}
}
