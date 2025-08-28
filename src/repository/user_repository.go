package repository

import (
	"errors"
	"fmt"
	"project/src/config"
	"project/src/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserByPhone(phone string) (*models.User, error)
	GetUserByOpenID(openID string) (*models.User, error)
	CreateUser(phone, openID, nickname string) (*models.User, error)
	UpdateLastLoginTime(userID uint) error
	GetUserByID(userID uint) (*models.User, error)
	GetUserByIDString(userIDStr string) (*models.User, error)
	UpdateUserNickname(userID uint, nickname string) error
	UpdateUserAvatar(userID uint, avatarURL string) error
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() UserRepository {
	return &userRepository{
		db: config.GetDB(),
	}
}

// GetUserByPhone 根据手机号获取用户
func (r *userRepository) GetUserByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// GetUserByOpenID 根据OpenID获取用户
func (r *userRepository) GetUserByOpenID(openID string) (*models.User, error) {
	if openID == "" {
		return nil, errors.New("user not found")
	}
	var user models.User
	err := r.db.Where("openid = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// CreateUser 创建新用户
func (r *userRepository) CreateUser(phone, openID, nickname string) (*models.User, error) {
	// 检查用户是否已存在（通过 OpenID）
	if openID != "" {
		if existingUser, _ := r.GetUserByOpenID(openID); existingUser != nil {
			return nil, errors.New("user already exists")
		}
	}

	// 如果有手机号，也检查手机号是否已存在
	if phone != "" {
		if existingUser, _ := r.GetUserByPhone(phone); existingUser != nil {
			return nil, errors.New("user already exists")
		}
	}

	now := time.Now()
	user := models.User{
		Phone:       phone,
		OpenID:      nil, // 先设为 nil
		Nickname:    nickname,
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: now,
	}

	// 如果 openID 不为空，设置指针
	if openID != "" {
		user.OpenID = &openID
	}

	// 创建用户
	err := r.db.Create(&user).Error
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	return &user, nil
}

// UpdateLastLoginTime 更新最后登录时间
func (r *userRepository) UpdateLastLoginTime(userID uint) error {
	err := r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"updated_at":    gorm.Expr("NOW()"),
		}).Error

	if err != nil {
		return fmt.Errorf("更新最后登录时间失败: %v", err)
	}

	return nil
}

// GetUserByID 根据用户ID获取用户
func (r *userRepository) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// GetUserByIDString 根据字符串用户ID获取用户（用于JWT）
func (r *userRepository) GetUserByIDString(userIDStr string) (*models.User, error) {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}
	return r.GetUserByID(uint(userID))
}

// UpdateUserNickname 更新用户昵称
func (r *userRepository) UpdateUserNickname(userID uint, nickname string) error {
	err := r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"nickname":   nickname,
			"updated_at": gorm.Expr("NOW()"),
		}).Error

	if err != nil {
		return fmt.Errorf("更新用户昵称失败: %v", err)
	}

	return nil
}

// UpdateUserAvatar 更新用户头像
func (r *userRepository) UpdateUserAvatar(userID uint, avatarURL string) error {
	err := r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"avatar_url": avatarURL,
			"updated_at": gorm.Expr("NOW()"),
		}).Error

	if err != nil {
		return fmt.Errorf("更新用户头像失败: %v", err)
	}

	return nil
}
