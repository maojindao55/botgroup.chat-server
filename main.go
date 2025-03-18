package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"project/api"
	"project/config"
	"project/middleware"
)

func main() {
	// 加载配置
	config.LoadConfig()

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
		// 初始化接口
		apiGroup.GET("/init", api.InitHandler)
		// 聊天相关接口
		apiGroup.POST("/chat", api.ChatHandler)
		// 调度相关接口
		apiGroup.POST("/scheduler", api.SchedulerHandler)
	}
}
