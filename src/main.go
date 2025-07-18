package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"project/src/api"
	"project/src/config"
	"project/src/middleware"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	config.InitDatabase()

	// 创建Gin引擎
	r := gin.Default()

	// 设置信任代理
	r.SetTrustedProxies(nil)

	// 注册中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 注册路由
	registerRoutes(r)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func registerRoutes(r *gin.Engine) {
	// 根路径响应
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "BotGroup Chat API Server",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 简单健康检查端点（用于Docker健康检查）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "服务正常运行",
		})
	})

	// 详细健康检查端点（包含数据库检查）
	r.GET("/health/detailed", func(c *gin.Context) {
		// 检查数据库连接
		sqlDB, err := config.DB.DB()
		if err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "数据库连接失败",
				"error":   err.Error(),
			})
			return
		}

		// 测试数据库连接
		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "数据库ping失败",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "ok",
			"message":  "服务正常运行",
			"database": "connected",
		})
	})

	// API路由组
	apiGroup := r.Group("/api")
	{

		// // 短信相关接口
		// apiGroup.POST("/sms/send", api.SendSMSHandler)
		// apiGroup.POST("/sms/send-template", api.SendSMSWithTemplateHandler)
		// 用户登录相关接口
		apiGroup.POST("/login", api.LoginHandler)
		apiGroup.POST("/sendcode", api.SendCodeHandler) // 测试用接口
		apiGroup.GET("/captcha", api.CaptchaHandler)
		apiGroup.GET("/captcha/check", api.CaptchaCheckHandler)

		// 需要认证的用户接口
		userGroup := apiGroup.Group("/")
		userGroup.Use(middleware.AuthMiddleware())
		{
			// 初始化接口
			userGroup.GET("/init", api.InitHandler)
			// 聊天相关接口
			userGroup.POST("/chat", api.ChatHandler)
			// 调度相关接口
			userGroup.POST("/scheduler", api.SchedulerHandler)
			// 用户相关接口
			userGroup.GET("/user/info", api.UserInfoHandler)
			userGroup.POST("/user/update", api.UserUpdateHandler)
			// 上传相关接口
			userGroup.POST("/user/upload", api.UploadHandler)
		}
	}
}
