package utils

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	hash := md5.Sum([]byte(time.Now().String()))
	return hex.EncodeToString(hash[:]) + string(timestamp)
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
