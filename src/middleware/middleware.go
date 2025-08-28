package middleware

import (
	"fmt"
	"net/http"
	"project/src/config"
	"project/src/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqURI := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		c.Writer.Header().Set("X-Response-Time", latencyTime.String())

		// 打印日志
		gin.DefaultWriter.Write([]byte("[GIN] " + time.Now().Format("2006-01-02 15:04:05") +
			" | " + clientIP + " | " + reqMethod + " | " +
			reqURI + " | " + c.Errors.String() + " | " +
			time.Since(startTime).String() + " | Status: " +
			string(statusCode) + "\n"))
	}
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果auth_access为0，则不进行认证
		if config.AppConfig.AuthAccess == 0 {
			c.Next()
			return
		}
		// 获取Authorization头部
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少认证信息",
			})
			c.Abort()
			return
		}

		// 验证Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 创建用户服务并验证token
		userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
		user, err := userService.ValidateToken(token)
		if err != nil {
			fmt.Println("ValidateToken error", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "认证失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Next()
	}
}
