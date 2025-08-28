package middleware

import (
	"fmt"
	"net/http"
	"project/src/config"
	"project/src/services"
	"strconv"
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
			strconv.Itoa(statusCode) + "\n"))
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

// ChatRateLimitMiddleware Chat接口限流中间件
// 允许匿名用户每个IP每天调用指定次数的chat接口，超过后要求登录
func ChatRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果已经有用户信息（已登录），直接跳过限流检查
		if _, exists := c.Get("user"); exists || config.AppConfig.AuthAccess == 0 {
			c.Next()
			return
		}

		// 获取客户端真实IP（考虑反向代理）
		clientIP := getRealClientIP(c)

		// 调试日志：显示IP获取信息
		fmt.Printf("ChatRateLimitMiddleware Debug - ClientIP: %s, X-Real-IP: %s, X-Forwarded-For: %s, RemoteAddr: %s\n",
			clientIP, c.GetHeader("X-Real-IP"), c.GetHeader("X-Forwarded-For"), c.Request.RemoteAddr)

		// 创建KV服务实例
		kvService := services.NewKVService(config.AppConfig.Redis)

		// 构造Redis key
		rateLimitKey := fmt.Sprintf("chat_rate_limit:%s:%s", clientIP, time.Now().Format("2006-01-02"))

		// 获取当前访问次数
		countStr, err := kvService.Get(rateLimitKey)
		var count int = 0
		if err == nil && countStr != "" {
			if parsedCount, parseErr := strconv.Atoi(countStr); parseErr == nil {
				count = parsedCount
			}
		}

		// 检查是否超过限制（默认2次）
		maxAttempts := 20
		if config.AppConfig.ChatRateLimit > 0 {
			maxAttempts = config.AppConfig.ChatRateLimit
		}

		if count >= maxAttempts {
			fmt.Println("ChatRateLimitMiddleware", "IP: ", clientIP, "次数: ", count, "限制: ", maxAttempts)
			// 超过限制，要求登录
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": fmt.Sprintf("匿名用户每日chat调用次数已达上限(%d次)，请登录后继续使用", maxAttempts),
				"code":    "CHAT_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		// 增加访问次数
		count++
		countStr = strconv.Itoa(count)

		// 设置过期时间为当天结束（第二天0点）
		now := time.Now()
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		ttl := tomorrow.Sub(now)

		// 更新Redis中的计数
		kvService.Set(rateLimitKey, countStr, ttl)

		// 在响应头中添加剩余次数信息
		remaining := maxAttempts - count
		c.Header("X-Chat-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-Chat-Rate-Limit-Reset", tomorrow.Format(time.RFC3339))

		c.Next()
	}
}

// getRealClientIP 获取客户端真实IP地址
// 优先级：X-Real-IP > X-Forwarded-For > RemoteAddr
func getRealClientIP(c *gin.Context) string {
	// 1. 尝试从 X-Real-IP 获取（nginx设置的真实IP）
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 2. 尝试从 X-Forwarded-For 获取（可能有多个IP，取第一个）
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，格式为：client, proxy1, proxy2
		// 我们取第一个IP作为客户端IP
		if strings.Contains(forwardedFor, ",") {
			return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
		}
		return forwardedFor
	}

	// 3. 使用gin的默认方法作为后备
	return c.ClientIP()
}
