package api

import (
	"net/http"

	"project/src/config"

	"github.com/gin-gonic/gin"
)

// InitHandler 返回应用程序初始化所需的配置信息
func InitHandler(c *gin.Context) {
	// 从配置中获取需要暴露给前端的配置信息
	initData := map[string]interface{}{
		"models":     config.AppConfig.LLMModels,
		"groups":     config.AppConfig.LLMGroups,
		"characters": config.AppConfig.LLMCharacters,
	}

	// 如果存在用户信息，则添加到响应中
	if user, exists := c.Get("user"); exists {
		initData["user"] = user
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    initData,
	})
}
