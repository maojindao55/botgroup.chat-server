package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project/models"
	"project/services"
)

// SchedulerRequest 调度请求结构
type SchedulerRequest struct {
	TaskName    string `json:"task_name" binding:"required"`
	Description string `json:"description"`
	CronExpr    string `json:"cron_expr" binding:"required"`
}

// SchedulerHandler 处理调度任务创建请求
func SchedulerHandler(c *gin.Context) {
	var req SchedulerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数无效",
		})
		return
	}

	// 调用服务层处理业务逻辑
	schedulerService := services.NewSchedulerService()
	task := models.Task{
		Name:        req.TaskName,
		Description: req.Description,
		CronExpr:    req.CronExpr,
	}

	id, err := schedulerService.CreateTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": id,
		"message": "任务创建成功",
	})
}

// GetSchedulerHandler 获取调度任务列表
func GetSchedulerHandler(c *gin.Context) {
	schedulerService := services.NewSchedulerService()
	tasks, err := schedulerService.GetAllTasks()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取任务列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}
