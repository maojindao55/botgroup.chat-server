package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"project/config"
	"project/models"
	"project/services"
)

// ScheduleRequest 调度请求结构
type ScheduleRequest struct {
	Message      string                 `json:"message" binding:"required"`
	History      []models.ChatMessage   `json:"history"`
	AvailableAIs []*config.LLMCharacter `json:"availableAIs" binding:"required"`
}

// ScheduleResponse 调度响应结构
type ScheduleResponse struct {
	SelectedAIs []string `json:"selectedAIs"`
}

// SchedulerHandler 处理AI调度请求
func SchedulerHandler(c *gin.Context) {
	var req ScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("调度请求参数无效: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	// 调用调度服务
	schedulerService := services.NewSchedulerService()
	selectedAIs, err := schedulerService.ScheduleAIResponses(req.Message, req.History, req.AvailableAIs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回选中的AI列表
	c.JSON(http.StatusOK, ScheduleResponse{
		SelectedAIs: selectedAIs,
	})
}
