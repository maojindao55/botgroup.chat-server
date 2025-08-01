package main

import (
	"fmt"
	"project/src/config"
	"project/src/utils"
)

// 配置测试工具
// 运行方式: go run scripts/test_config.go
func main() {
	fmt.Println("=== 微信登录配置测试工具 ===")
	fmt.Println()

	// 加载配置
	fmt.Println("正在加载配置...")
	config.LoadConfig()
	fmt.Println("✅ 配置加载完成")
	fmt.Println()

	// 验证配置
	fmt.Println("正在验证配置...")
	if err := utils.ValidateAllConfigs(); err != nil {
		fmt.Printf("❌ 配置验证失败:\n%s\n", err.Error())
		return
	}
	fmt.Println("✅ 配置验证通过")
	fmt.Println()

	// 显示配置摘要
	fmt.Println("=== 配置摘要 ===")
	fmt.Printf("微信配置: %s\n", utils.GetWechatConfigSummary())
	fmt.Printf("WebSocket配置: ReadBuffer=%d, WriteBuffer=%d, CheckOrigin=%t\n",
		config.AppConfig.WebSocket.ReadBufferSize,
		config.AppConfig.WebSocket.WriteBufferSize,
		config.AppConfig.WebSocket.CheckOrigin)
	fmt.Printf("Redis配置: %s:%s, DB=%d\n",
		config.AppConfig.Redis.Host,
		config.AppConfig.Redis.Port,
		config.AppConfig.Redis.DB)
	fmt.Printf("JWT Secret: %s***\n", config.AppConfig.JWTSecret[:min(len(config.AppConfig.JWTSecret), 8)])
	fmt.Println()

	// 配置检查建议
	fmt.Println("=== 配置检查建议 ===")
	checkWechatConfig()
	checkWebSocketConfig()
	checkRedisConfig()

	fmt.Println()
	fmt.Println("🎉 配置测试完成!")
}

func checkWechatConfig() {
	wechat := config.AppConfig.Wechat

	// 检查是否使用默认值
	if wechat.AppID == "wx1234567890abcdef" {
		fmt.Println("⚠️  微信AppID使用默认值，请在生产环境中设置真实值")
	}

	if wechat.AppSecret == "1234567890abcdef12345678" {
		fmt.Println("⚠️  微信AppSecret使用默认值，请在生产环境中设置真实值")
	}

	if wechat.Token == "your_wechat_token_here" {
		fmt.Println("⚠️  微信Token使用默认值，请设置复杂的自定义Token")
	}

	if wechat.CallbackURL == "https://your-domain.com/api/auth/wechat/callback" {
		fmt.Println("⚠️  微信回调URL使用默认值，请设置真实的域名")
	}

	// 检查过期时间设置
	if wechat.QRExpiresIn < 300 {
		fmt.Println("⚠️  二维码过期时间较短，建议设置为300秒以上")
	}

	if wechat.SessionExpiresIn < 600 {
		fmt.Println("⚠️  会话过期时间较短，建议设置为600秒以上")
	}
}

func checkWebSocketConfig() {
	ws := config.AppConfig.WebSocket

	if ws.ReadBufferSize < 1024 {
		fmt.Println("ℹ️  WebSocket读取缓冲区较小，建议设置为1024以上")
	}

	if ws.WriteBufferSize < 1024 {
		fmt.Println("ℹ️  WebSocket写入缓冲区较小，建议设置为1024以上")
	}

	if !ws.CheckOrigin {
		fmt.Println("⚠️  WebSocket未启用源检查，生产环境建议启用")
	}
}

func checkRedisConfig() {
	redis := config.AppConfig.Redis

	if redis.Host == "localhost" || redis.Host == "127.0.0.1" {
		fmt.Println("ℹ️  Redis使用本地地址，确保在正确的环境中运行")
	}

	if redis.Password == "" {
		fmt.Println("⚠️  Redis未设置密码，生产环境建议设置密码")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
