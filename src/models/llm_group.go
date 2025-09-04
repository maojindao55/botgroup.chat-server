package models

import (
	"time"
)

// LlmGroup 群组模型
type LlmGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;index;comment:群组名称"`
	Description string    `json:"description" gorm:"type:text;comment:群组描述"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Characters []GroupCharacter `json:"characters,omitempty" gorm:"foreignKey:GID;references:ID"`
}

// TableName 设置表名
func (LlmGroup) TableName() string {
	return "llm_groups"
}

// LlmGroupCreateRequest 创建群组请求
type LlmGroupCreateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

// LlmGroupUpdateRequest 更新群组请求
type LlmGroupUpdateRequest struct {
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description" binding:"max=1000"`
}

// LlmGroupResponse 群组响应
type LlmGroupResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *LlmGroup `json:"data,omitempty"`
}

// LlmGroupListResponse 群组列表响应
type LlmGroupListResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    []LlmGroup `json:"data,omitempty"`
	Total   int64      `json:"total,omitempty"`
}
