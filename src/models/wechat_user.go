package models

import (
	"time"
)

// WechatUser 微信用户模型
type WechatUser struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UID            uint      `json:"uid" gorm:"index;comment:关联用户ID"` // 关联users表的id
	OpenID         string    `json:"openid" gorm:"uniqueIndex;size:64;not null;comment:微信OpenID"`
	Nickname       string    `json:"nickname" gorm:"size:100;comment:微信昵称"`
	AvatarURL      string    `json:"avatar_url" gorm:"column:avatar_url;type:text;comment:微信头像URL"`
	SubscribeScene string    `json:"subscribe_scene" gorm:"size:50;comment:关注场景"`
	QRScene        string    `json:"qr_scene" gorm:"size:100;index;comment:二维码场景值"`
	SubscribeTime  time.Time `json:"subscribe_time" gorm:"index;autoCreateTime;comment:关注时间"`
	LastLoginAt    time.Time `json:"last_login_at" gorm:"autoCreateTime;comment:最后登录时间"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UID;references:ID"`
}

// TableName 设置表名
func (WechatUser) TableName() string {
	return "wechat_users"
}

// LoginSession Redis中存储的登录会话结构
type LoginSession struct {
	SessionID string `json:"session_id"` // 会话ID
	QRScene   string `json:"qr_scene"`   // 二维码场景值
	Status    string `json:"status"`     // pending|success|expired
	UserID    uint   `json:"user_id"`    // 登录成功后的用户ID
	OpenID    string `json:"openid"`     // 微信openid
	CreatedAt int64  `json:"created_at"` // 创建时间戳
	ExpiresAt int64  `json:"expires_at"` // 过期时间戳
}

// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	RedirectURI string `json:"redirect_uri,omitempty"` // 登录成功后的跳转地址（可选）
}

// WechatLoginResponse 微信登录响应
type WechatLoginResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *WechatLoginData `json:"data,omitempty"`
}

// WechatLoginData 微信登录数据
type WechatLoginData struct {
	QRUrl     string `json:"qr_url"`     // 二维码图片URL
	SessionID string `json:"session_id"` // 会话ID（用于WebSocket连接）
	QRScene   string `json:"qr_scene"`   // 二维码场景值
	ExpiresIn int    `json:"expires_in"` // 过期时间(秒)
}

// WebSocketLoginResult WebSocket推送的登录结果
type WebSocketLoginResult struct {
	Type string              `json:"type"` // login_result
	Data *WebSocketLoginData `json:"data"`
}

// WebSocketLoginData WebSocket登录数据
type WebSocketLoginData struct {
	Status    string    `json:"status"`     // success|failed|expired
	Message   string    `json:"message"`    // 状态描述
	UserInfo  *UserInfo `json:"user_info"`  // 用户信息
	Token     string    `json:"token"`      // JWT Token
	ExpiresIn int       `json:"expires_in"` // Token过期时间(秒)
}

// UserInfo 用户信息
type UserInfo struct {
	UserID    uint   `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	LoginType string `json:"login_type"` // wechat
}
