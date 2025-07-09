package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Phone       string    `json:"phone" gorm:"uniqueIndex;size:11;not null"`
	Nickname    string    `json:"nickname" gorm:"size:50"`
	AvatarURL   string    `json:"avatar_url" gorm:"column:avatar_url"`
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at"`
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
