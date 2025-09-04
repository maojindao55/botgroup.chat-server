package models

import (
	"fmt"
	"project/src/config"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GroupCharacter 群组角色模型
type GroupCharacter struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	GID          uint      `json:"gid" gorm:"not null;index;comment:群组ID，关联llm_groups表的id字段"`
	Name         string    `json:"name" gorm:"size:100;not null;index;comment:角色名称"`
	Personality  string    `json:"personality" gorm:"size:100;not null;default:'';comment:角色性格描述"`
	Model        string    `json:"model" gorm:"size:50;index;comment:AI模型名称"`
	Avatar       string    `json:"avatar" gorm:"type:text;comment:角色头像URL"`
	CustomPrompt string    `json:"custom_prompt" gorm:"type:text;comment:自定义提示词"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Group *LlmGroup `json:"group,omitempty" gorm:"foreignKey:GID;references:ID"`
}

// TableName 设置表名
func (GroupCharacter) TableName() string {
	return "group_characters"
}

// AfterFind GORM Hook: 查询后自动添加头像URL前缀
func (gc *GroupCharacter) AfterFind(tx *gorm.DB) error {
	if gc.Avatar != "" && config.AppConfig.Cloudflare.ImagePrefix != "" {
		// 检查URL是否已经包含前缀，避免重复添加
		if !strings.HasPrefix(gc.Avatar, config.AppConfig.Cloudflare.ImagePrefix) {
			gc.Avatar = fmt.Sprintf(config.AppConfig.Cloudflare.ImagePrefix, gc.Avatar)
		}
	}
	return nil
}

// GroupCharacterCreateRequest 创建群组角色请求
type GroupCharacterCreateRequest struct {
	GID          uint   `json:"gid" binding:"required"`
	Name         string `json:"name" binding:"required,max=100"`
	Personality  string `json:"personality" binding:"max=100"`
	Model        string `json:"model" binding:"max=50"`
	Avatar       string `json:"avatar" binding:"max=500"`
	CustomPrompt string `json:"custom_prompt" binding:"max=2000"`
}

// GroupCharacterUpdateRequest 更新群组角色请求
type GroupCharacterUpdateRequest struct {
	Name         string `json:"name" binding:"max=100"`
	Personality  string `json:"personality" binding:"max=100"`
	Model        string `json:"model" binding:"max=50"`
	Avatar       string `json:"avatar" binding:"max=500"`
	CustomPrompt string `json:"custom_prompt" binding:"max=2000"`
}

// GroupCharacterResponse 群组角色响应
type GroupCharacterResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    *GroupCharacter `json:"data,omitempty"`
}

// GroupCharacterListResponse 群组角色列表响应
type GroupCharacterListResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    []GroupCharacter `json:"data,omitempty"`
	Total   int64            `json:"total,omitempty"`
}

// GroupCharactersByGroupResponse 按群组获取角色列表响应
type GroupCharactersByGroupResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Group   *LlmGroup        `json:"group,omitempty"`
	Data    []GroupCharacter `json:"data,omitempty"`
	Total   int64            `json:"total,omitempty"`
}
