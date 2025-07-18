package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"regexp"
	"time"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	hash := md5.Sum([]byte(time.Now().String()))
	return hex.EncodeToString(hash[:]) + string(timestamp)
}

// GenerateRandomCode 生成指定长度的随机数字验证码
func GenerateRandomCode(length int) string {
	if length <= 0 {
		length = 6
	}

	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < length; i++ {
		code += string(rune('0' + rand.Intn(10)))
	}
	return code
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseCronExpression 解析Cron表达式
func ParseCronExpression(expr string) (bool, error) {
	// 这里可以添加Cron表达式解析逻辑
	// 简单实现，实际应使用cron库
	return true, nil
}

// IsValidPhone 验证手机号格式（中国大陆）
func IsValidPhone(phone string) bool {
	// 中国大陆手机号正则表达式
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidCode 验证验证码格式（4-8位数字）
func IsValidCode(code string) bool {
	codeRegex := regexp.MustCompile(`^\d{4,8}$`)
	return codeRegex.MatchString(code)
}
