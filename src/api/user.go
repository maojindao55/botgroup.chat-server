package api

import (
	"net/http"
	"project/src/config"
	"project/src/models"
	"project/src/services"

	"github.com/gin-gonic/gin"
)

// UserInfoResponse 用户信息响应结构
type UserInfoResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *models.User `json:"data,omitempty"`
}

// UserUpdateRequest 用户更新请求结构
type UserUpdateRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

// UserUpdateResponse 用户更新响应结构
type UserUpdateResponse struct {
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

// UserUpdateHandler 更新用户信息处理器
func UserUpdateHandler(c *gin.Context) {
	// 从认证中间件中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, UserUpdateResponse{
			Success: false,
			Message: "用户认证失败",
		})
		return
	}

	// 类型断言
	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, UserUpdateResponse{
			Success: false,
			Message: "用户信息类型错误",
		})
		return
	}

	// 解析请求参数
	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, UserUpdateResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建用户服务
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)

	// 调用服务更新昵称
	err := userService.UpdateNickname(user.ID, req.Nickname)
	if err != nil {
		c.JSON(http.StatusBadRequest, UserUpdateResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	user.Nickname = req.Nickname
	// 返回成功响应
	c.JSON(http.StatusOK, UserUpdateResponse{
		Success: true,
		Message: "更新用户昵称成功",
		Data:    user,
	})
}
