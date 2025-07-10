package api

import (
	"net/http"
	"project/src/models"

	"github.com/gin-gonic/gin"
)

// UserInfoResponse 用户信息响应结构
type UserInfoResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *models.User `json:"data,omitempty"`
}

// UserInfoHandler 获取用户信息处理器
func UserInfoHandler(c *gin.Context) {
	// 从认证中间件中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, UserInfoResponse{
			Success: false,
			Message: "用户认证失败",
		})
		return
	}

	// 类型断言
	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, UserInfoResponse{
			Success: false,
			Message: "用户信息类型错误",
		})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, UserInfoResponse{
		Success: true,
		Message: "获取用户信息成功",
		Data:    user,
	})
}
