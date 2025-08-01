package utils

import (
	"errors"
	"fmt"
	"project/src/config"
	"strings"
)

// ValidateWechatConfig 验证微信配置
func ValidateWechatConfig() error {
	wechatConfig := config.AppConfig.Wechat

	// 检查必需的配置项
	if wechatConfig.AppID == "" {
		return errors.New("微信AppID未配置，请设置WECHAT_APP_ID环境变量")
	}

	if wechatConfig.AppSecret == "" {
		return errors.New("微信AppSecret未配置，请设置WECHAT_APP_SECRET环境变量")
	}

	if wechatConfig.Token == "" {
		return errors.New("微信Token未配置，请设置WECHAT_TOKEN环境变量")
	}

	if wechatConfig.CallbackURL == "" {
		return errors.New("微信回调URL未配置，请设置WECHAT_CALLBACK_URL环境变量")
	}

	// 验证配置格式
	if !strings.HasPrefix(wechatConfig.AppID, "wx") {
		return errors.New("微信AppID格式错误，应以wx开头")
	}

	if len(wechatConfig.AppSecret) != 32 {
		return errors.New("微信AppSecret格式错误，应为32位字符串")
	}

	if !strings.HasPrefix(wechatConfig.CallbackURL, "https://") {
		return errors.New("微信回调URL必须使用HTTPS协议")
	}

	// 检查默认值
	if wechatConfig.QRExpiresIn <= 0 {
		return errors.New("二维码过期时间必须大于0")
	}

	if wechatConfig.SessionExpiresIn <= 0 {
		return errors.New("会话过期时间必须大于0")
	}

	return nil
}

// ValidateWebSocketConfig 验证WebSocket配置
func ValidateWebSocketConfig() error {
	wsConfig := config.AppConfig.WebSocket

	if wsConfig.ReadBufferSize <= 0 {
		return errors.New("WebSocket读取缓冲区大小必须大于0")
	}

	if wsConfig.WriteBufferSize <= 0 {
		return errors.New("WebSocket写入缓冲区大小必须大于0")
	}

	return nil
}

// ValidateRedisConfig 验证Redis配置
func ValidateRedisConfig() error {
	redisConfig := config.AppConfig.Redis

	if redisConfig.Host == "" {
		return errors.New("Redis主机地址未配置")
	}

	if redisConfig.Port == "" {
		return errors.New("Redis端口未配置")
	}

	return nil
}

// ValidateAllConfigs 验证所有配置
func ValidateAllConfigs() error {
	validators := []struct {
		name      string
		validator func() error
	}{
		{"微信配置", ValidateWechatConfig},
		{"WebSocket配置", ValidateWebSocketConfig},
		{"Redis配置", ValidateRedisConfig},
	}

	var errors []string
	for _, v := range validators {
		if err := v.validator(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", v.name, err.Error()))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// GetWechatConfigSummary 获取微信配置摘要（用于日志）
func GetWechatConfigSummary() string {
	wechatConfig := config.AppConfig.Wechat
	return fmt.Sprintf("AppID: %s, Token: %s***, CallbackURL: %s, QRExpires: %ds, SessionExpires: %ds",
		wechatConfig.AppID,
		wechatConfig.Token[:min(len(wechatConfig.Token), 8)],
		wechatConfig.CallbackURL,
		wechatConfig.QRExpiresIn,
		wechatConfig.SessionExpiresIn)
}

// min 函数用于获取两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
