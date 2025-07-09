package api

import (
	"net/http"
	"project/config"
	"project/services"
	"project/utils"

	"github.com/gin-gonic/gin"
)

// SendSMSRequest 发送短信请求结构
type SendSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// SendSMSWithTemplateRequest 发送模板短信请求结构
type SendSMSWithTemplateRequest struct {
	Phone         string            `json:"phone" binding:"required"`
	TemplateCode  string            `json:"template_code" binding:"required"`
	TemplateParam map[string]string `json:"template_param"`
}

// SMSResponse 短信发送响应结构
type SMSResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendSMSHandler 发送验证码短信
func SendSMSHandler(c *gin.Context) {
	var req SendSMSRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SMSResponse{
			Success: false,
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, SMSResponse{
			Success: false,
			Message: "手机号格式无效",
		})
		return
	}

	// 验证验证码格式
	if !utils.IsValidCode(req.Code) {
		c.JSON(http.StatusBadRequest, SMSResponse{
			Success: false,
			Message: "验证码格式无效",
		})
		return
	}

	// 创建SMS服务
	smsService := services.NewSMSService(config.AppConfig.SMS)

	// 发送短信
	if err := smsService.SendSMS(req.Phone, req.Code); err != nil {
		c.JSON(http.StatusInternalServerError, SMSResponse{
			Success: false,
			Message: "短信发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SMSResponse{
		Success: true,
		Message: "短信发送成功",
	})
}

// SendSMSWithTemplateHandler 发送模板短信
func SendSMSWithTemplateHandler(c *gin.Context) {
	var req SendSMSWithTemplateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SMSResponse{
			Success: false,
			Message: "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, SMSResponse{
			Success: false,
			Message: "手机号格式无效",
		})
		return
	}

	// 创建SMS服务
	smsService := services.NewSMSService(config.AppConfig.SMS)

	// 发送短信
	if err := smsService.SendSMSWithTemplate(req.Phone, req.TemplateCode, req.TemplateParam); err != nil {
		c.JSON(http.StatusInternalServerError, SMSResponse{
			Success: false,
			Message: "短信发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SMSResponse{
		Success: true,
		Message: "短信发送成功",
	})
}
