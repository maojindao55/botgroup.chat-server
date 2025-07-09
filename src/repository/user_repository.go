package repository

import (
	"errors"
	"fmt"
	"project/models"
	"strconv"
	"time"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserByPhone(phone string) (*models.User, error)
	CreateUser(phone, nickname string) (*models.User, error)
	UpdateLastLoginTime(userID uint) error
	GetUserByID(userID uint) (*models.User, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	users  []models.User // 临时存储，实际应使用数据库
	nextID uint
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() UserRepository {
	// 预置测试用户
	users := []models.User{
		{
			ID:          1,
			Phone:       "13800138000",
			Nickname:    "测试用户",
			AvatarURL:   "",
			Status:      1,
			CreatedAt:   time.Date(2025, 3, 26, 8, 39, 15, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 3, 26, 8, 39, 15, 0, time.UTC),
			LastLoginAt: time.Date(2025, 3, 26, 8, 39, 15, 0, time.UTC),
		},
	}

	return &userRepository{
		users:  users,
		nextID: 2, // 下一个ID从2开始
	}
}

// GetUserByPhone 根据手机号获取用户
func (r *userRepository) GetUserByPhone(phone string) (*models.User, error) {
	for _, user := range r.users {
		if user.Phone == phone {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

// CreateUser 创建新用户
func (r *userRepository) CreateUser(phone, nickname string) (*models.User, error) {
	// 检查用户是否已存在
	if existingUser, _ := r.GetUserByPhone(phone); existingUser != nil {
		return nil, errors.New("user already exists")
	}

	now := time.Now()
	user := models.User{
		ID:          r.nextID,
		Phone:       phone,
		Nickname:    nickname,
		AvatarURL:   "",
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: now,
	}

	r.users = append(r.users, user)
	r.nextID++

	return &user, nil
}

// UpdateLastLoginTime 更新最后登录时间
func (r *userRepository) UpdateLastLoginTime(userID uint) error {
	for i, user := range r.users {
		if user.ID == userID {
			r.users[i].LastLoginAt = time.Now()
			r.users[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return errors.New("user not found")
}

// GetUserByID 根据用户ID获取用户
func (r *userRepository) GetUserByID(userID uint) (*models.User, error) {
	for _, user := range r.users {
		if user.ID == userID {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetUserByIDString 根据字符串用户ID获取用户（用于JWT）
func (r *userRepository) GetUserByIDString(userIDStr string) (*models.User, error) {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}
	return r.GetUserByID(uint(userID))
}
