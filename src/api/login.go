package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"project/src/config"
	"project/src/models"
	"project/src/services"
	"project/src/utils"

	"github.com/gin-gonic/gin"
)

// LoginHandler 用户登录处理器
func LoginHandler(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.UserLoginResponse{
			Success: false,
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 创建用户服务
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)

	// 执行登录
	userData, err := userService.Login(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.UserLoginResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 登录成功
	c.JSON(http.StatusOK, models.UserLoginResponse{
		Success: true,
		Message: "登录成功",
		Data:    userData,
	})
}

// SendCodeRequest 发送验证码请求结构
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// SendCodeResponse 发送验证码响应结构
type SendCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		Code string `json:"code,omitempty"` // 仅在开发环境返回
	} `json:"data,omitempty"`
}

// SendCodeHandler 发送验证码处理器
func SendCodeHandler(c *gin.Context) {
	var req SendCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendCodeResponse{
			Success: false,
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, SendCodeResponse{
			Success: false,
			Message: "手机号格式无效",
		})
		return
	}

	// 生成6位随机验证码
	verificationCode, err := generateVerificationCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, SendCodeResponse{
			Success: false,
			Message: "生成验证码失败: " + err.Error(),
		})
		return
	}

	// 创建短信服务
	smsService := services.NewSMSService(config.AppConfig.SMS)

	// 发送短信验证码
	if err := smsService.SendSMS(req.Phone, verificationCode); err != nil {
		c.JSON(http.StatusInternalServerError, SendCodeResponse{
			Success: false,
			Message: "短信发送失败: " + err.Error(),
		})
		return
	}

	// 创建用户服务并将验证码存储到KV中
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
	if err := userService.SetSMSCode(req.Phone, verificationCode); err != nil {
		c.JSON(http.StatusInternalServerError, SendCodeResponse{
			Success: false,
			Message: "存储验证码失败: " + err.Error(),
		})
		return
	}

	// 构建响应
	response := SendCodeResponse{
		Success: true,
		Message: "验证码发送成功",
	}

	// 在开发环境下返回验证码（仅用于测试）
	// 可以通过环境变量或配置来控制
	if config.AppConfig.Server.Port == "8080" { // 简单的开发环境判断，可以改为更精确的判断
		response.Data = &struct {
			Code string `json:"code,omitempty"`
		}{
			Code: verificationCode,
		}
	}

	c.JSON(http.StatusOK, response)
}

// generateVerificationCode 生成6位随机验证码
func generateVerificationCode() (string, error) {
	// 生成 100000 到 999999 之间的随机数
	min := int64(100000)
	max := int64(999999)

	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return "", fmt.Errorf("生成随机数失败: %v", err)
	}

	code := min + n.Int64()
	return fmt.Sprintf("%06d", code), nil
}
