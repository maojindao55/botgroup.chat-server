package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"project/src/config"
	"project/src/services"
	"project/src/services/go-captcha/captdata"
	"project/src/services/go-captcha/checkdata"
	"project/src/utils"

	"github.com/gin-gonic/gin"
)

// CaptchaCheckRequest 验证码检查请求结构
type CaptchaCheckRequest struct {
	Dots      string `json:"dots" form:"dots" binding:"required"`
	Key       string `json:"key" form:"key" binding:"required"`
	ExtraData string `json:"extraData" form:"extraData" binding:"required"`
}

// ExtraDataStruct 解析extraData的结构
type ExtraDataStruct struct {
	Phone string `json:"phone" binding:"required"`
}

// CaptchaResponse 验证码响应结构
type CaptchaResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// CaptchaHandler 验证码处理器 - 直接调用原始函数
func CaptchaHandler(c *gin.Context) {
	captdata.GetClickBasicCaptData(c.Writer, c.Request)
}

// CaptchaCheckHandler 验证码检查处理器 - 验证码通过后发送短信
func CaptchaCheckHandler(c *gin.Context) {
	var req CaptchaCheckRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, CaptchaResponse{
			Code:    1,
			Message: "请求参数无效: " + err.Error(),
			Success: false,
		})
		return
	}

	// 解析extraData中的JSON字符串
	var extraData ExtraDataStruct
	if err := json.Unmarshal([]byte(req.ExtraData), &extraData); err != nil {
		c.JSON(http.StatusBadRequest, CaptchaResponse{
			Code:    1,
			Message: "extraData格式无效: " + err.Error(),
			Success: false,
		})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(extraData.Phone) {
		c.JSON(http.StatusBadRequest, CaptchaResponse{
			Code:    1,
			Message: "手机号格式无效",
			Success: false,
		})
		return
	}

	// 验证验证码 - 使用现有的CheckClickData函数
	isValid := validateCaptchaWithExisting(req.Dots, req.Key)
	if !isValid {
		c.JSON(http.StatusBadRequest, CaptchaResponse{
			Code:    1,
			Message: "验证码验证失败",
			Success: false,
		})
		return
	}

	// 生成6位随机验证码
	smsCode := utils.GenerateRandomCode(6)

	// 创建SMS服务并发送短信
	smsService := services.NewSMSService(config.AppConfig.SMS)
	if err := smsService.SendSMS(extraData.Phone, smsCode); err != nil {
		c.JSON(http.StatusInternalServerError, CaptchaResponse{
			Code:    1,
			Message: "短信发送失败: " + err.Error(),
			Success: false,
		})
		return
	}

	// 创建用户服务来存储验证码
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)

	// 将验证码存储到缓存中，用于后续登录验证
	if err := userService.SetSMSCode(extraData.Phone, smsCode); err != nil {
		c.JSON(http.StatusInternalServerError, CaptchaResponse{
			Code:    1,
			Message: "验证码存储失败: " + err.Error(),
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, CaptchaResponse{
		Code:    0,
		Message: "验证码验证通过，短信已发送",
		Success: true,
	})
}

// validateCaptchaWithExisting 使用现有的CheckClickData函数验证验证码
func validateCaptchaWithExisting(dots, key string) bool {
	// 创建一个模拟的请求来调用CheckClickData
	formData := "dots=" + dots + "&key=" + key
	req, err := http.NewRequest("POST", "/captcha/check", bytes.NewBufferString(formData))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 创建一个ResponseWriter来捕获响应
	responseBuffer := &bytes.Buffer{}
	writer := &responseWriter{Buffer: responseBuffer}

	// 调用现有的CheckClickData函数
	checkdata.CheckClickData(writer, req)

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(responseBuffer.Bytes(), &response); err != nil {
		return false
	}

	// 检查响应码，0表示成功
	if code, ok := response["code"].(float64); ok && code == 0 {
		return true
	}

	return false
}

// responseWriter 实现http.ResponseWriter接口用于捕获响应
type responseWriter struct {
	*bytes.Buffer
	statusCode int
}

func (w *responseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
