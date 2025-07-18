package api

import (
	"project/src/services/go-captcha/captdata"
	"project/src/services/go-captcha/checkdata"

	"github.com/gin-gonic/gin"
)

// CaptchaHandler 验证码处理器 - 直接调用原始函数
func CaptchaHandler(c *gin.Context) {
	captdata.GetClickBasicCaptData(c.Writer, c.Request)
}

// CaptchaCheckHandler 验证码检查处理器 - 直接调用原始函数
func CaptchaCheckHandler(c *gin.Context) {
	checkdata.CheckClickData(c.Writer, c.Request)
}
