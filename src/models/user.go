package models

import (
	"fmt"
	"project/src/config"
	"strings"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Phone       string    `json:"phone" gorm:"uniqueIndex;size:11;not null;charset:utf8mb4;collation:utf8mb4_unicode_ci"`
	Nickname    string    `json:"nickname" gorm:"size:50;charset:utf8mb4;collation:utf8mb4_unicode_ci"`
	AvatarURL   string    `json:"avatar_url" gorm:"column:avatar_url;type:text;charset:utf8mb4;collation:utf8mb4_unicode_ci"`
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	LastLoginAt time.Time `json:"last_login_at" gorm:"autoCreateTime"`
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

// AfterFind GORM Hook: 查询后自动添加头像URL前缀
func (u *User) AfterFind(tx *gorm.DB) error {
	if u.AvatarURL != "" && config.AppConfig.Cloudflare.ImagePrefix != "" {
		// 检查URL是否已经包含前缀，避免重复添加
		if !strings.HasPrefix(u.AvatarURL, config.AppConfig.Cloudflare.ImagePrefix) {
			u.AvatarURL = fmt.Sprintf(config.AppConfig.Cloudflare.ImagePrefix, u.AvatarURL)
		}
	}
	return nil
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *UserData `json:"data,omitempty"`
}

// UserData 用户数据
type UserData struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
