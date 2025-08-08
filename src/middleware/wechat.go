package middleware

import (
	"net/http"
	"project/src/config"
	"project/src/services"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 简单的内存限流器
type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int           // 每分钟最大请求数
	window   time.Duration // 时间窗口
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// 启动清理goroutine
	go rl.cleanup()

	return rl
}

func (rl *rateLimiter) isAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 获取该IP的请求历史
	requests, exists := rl.requests[ip]
	if !exists {
		rl.requests[ip] = []time.Time{now}
		return true
	}

	// 过滤掉超出时间窗口的请求
	validRequests := make([]time.Time, 0)
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		for ip, requests := range rl.requests {
			validRequests := make([]time.Time, 0)
			for _, req := range requests {
				if req.After(cutoff) {
					validRequests = append(validRequests, req)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validRequests
			}
		}
		rl.mutex.Unlock()
	}
}

// 全局限流器实例
var (
	qrCodeLimiter   = newRateLimiter(10, time.Minute)  // 二维码生成：每分钟10次
	callbackLimiter = newRateLimiter(100, time.Minute) // 微信回调：每分钟100次
	statusLimiter   = newRateLimiter(60, time.Minute)  // 状态查询：每分钟60次
)

// WechatQRCodeRateLimit 二维码生成限流中间件
func WechatQRCodeRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !qrCodeLimiter.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求频率过高，请稍后再试",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// WechatCallbackRateLimit 微信回调限流中间件
func WechatCallbackRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !callbackLimiter.isAllowed(ip) {
			c.String(http.StatusTooManyRequests, "Rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

// WechatStatusRateLimit 状态查询限流中间件
func WechatStatusRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !statusLimiter.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "查询频率过高，请稍后再试",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// WechatSignatureVerify 微信签名验证中间件（仅用于回调接口）
func WechatSignatureVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对POST请求进行签名验证（事件回调）
		// GET请求的签名验证在处理函数中进行（服务器验证）
		if c.Request.Method == http.MethodPost {
			signature := c.Query("signature")
			timestamp := c.Query("timestamp")
			nonce := c.Query("nonce")

			if signature == "" || timestamp == "" || nonce == "" {
				c.String(http.StatusBadRequest, "Missing required parameters")
				c.Abort()
				return
			}

			// 创建回调服务进行签名验证
			kvService := services.NewKVService(config.AppConfig.Redis)
			sessionService := services.NewSessionService(kvService)
			userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
			callbackService := services.NewWechatCallbackService(sessionService, userService)

			if !callbackService.VerifySignature(signature, timestamp, nonce) {
				c.String(http.StatusUnauthorized, "Invalid signature")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// WechatCORS 微信专用CORS中间件（处理微信服务器的请求）
func WechatCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 微信服务器请求不需要CORS，但我们添加一些安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

// WechatRequestLogger 微信请求专用日志中间件
func WechatRequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return "[WECHAT] " + time.Now().Format("2006-01-02 15:04:05") +
			" | " + param.ClientIP + " | " + param.Method + " | " +
			param.Path + " | " + param.ErrorMessage + " | " +
			param.Latency.String() + " | Status: " +
			string(rune(param.StatusCode)) + "\n"
	})
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 对于API响应，不缓存敏感信息
		if c.Request.URL.Path != "/health" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}
