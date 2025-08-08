package repository

import (
	"project/src/config"
	"project/src/models"
	"time"

	"gorm.io/gorm"
)

// WechatUserRepository 微信用户仓库接口
type WechatUserRepository interface {
	GetWechatUserByOpenID(openID string) (*models.WechatUser, error)
	CreateWechatUser(wechatUser *models.WechatUser) error
	UpdateWechatUser(wechatUser *models.WechatUser) error
	GetWechatUserByUID(uid uint) (*models.WechatUser, error)
	UpdateLastLoginTime(openID string) error
	DeleteWechatUser(openID string) error
}

// wechatUserRepository 微信用户仓库实现
type wechatUserRepository struct {
	db *gorm.DB
}

// NewWechatUserRepository 创建微信用户仓库实例
func NewWechatUserRepository() WechatUserRepository {
	return &wechatUserRepository{
		db: config.GetDB(),
	}
}

// GetWechatUserByOpenID 根据OpenID获取微信用户
func (r *wechatUserRepository) GetWechatUserByOpenID(openID string) (*models.WechatUser, error) {
	var wechatUser models.WechatUser
	err := r.db.Where("openid = ?", openID).First(&wechatUser).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	return &wechatUser, err
}

// CreateWechatUser 创建微信用户
func (r *wechatUserRepository) CreateWechatUser(wechatUser *models.WechatUser) error {
	return r.db.Create(wechatUser).Error
}

// UpdateWechatUser 更新微信用户信息
func (r *wechatUserRepository) UpdateWechatUser(wechatUser *models.WechatUser) error {
	return r.db.Save(wechatUser).Error
}

// GetWechatUserByUID 根据用户ID获取微信用户
func (r *wechatUserRepository) GetWechatUserByUID(uid uint) (*models.WechatUser, error) {
	var wechatUser models.WechatUser
	err := r.db.Where("uid = ?", uid).First(&wechatUser).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	return &wechatUser, err
}

// UpdateLastLoginTime 更新最后登录时间
func (r *wechatUserRepository) UpdateLastLoginTime(openID string) error {
	return r.db.Model(&models.WechatUser{}).
		Where("openid = ?", openID).
		Update("last_login_at", time.Now()).Error
}

// DeleteWechatUser 删除微信用户
func (r *wechatUserRepository) DeleteWechatUser(openID string) error {
	return r.db.Where("openid = ?", openID).Delete(&models.WechatUser{}).Error
}
