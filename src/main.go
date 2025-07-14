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
	// API路由组
	apiGroup := r.Group("/api")
	{

		// // 短信相关接口
		// apiGroup.POST("/sms/send", api.SendSMSHandler)
		// apiGroup.POST("/sms/send-template", api.SendSMSWithTemplateHandler)
		// 用户登录相关接口
		apiGroup.POST("/login", api.LoginHandler)
		apiGroup.POST("/sendcode", api.SendCodeHandler) // 测试用接口

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
