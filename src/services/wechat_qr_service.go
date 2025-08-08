package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"project/src/config"
	"project/src/constants"
	"project/src/models"
	"strings"
	"time"
)

// WechatQRService 微信二维码服务
type WechatQRService struct {
	sessionService *SessionService
	kvService      KVService
}

// AccessTokenCache access_token缓存结构
type AccessTokenCache struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// NewWechatQRService 创建微信二维码服务实例
func NewWechatQRService(sessionService *SessionService, kvService KVService) *WechatQRService {
	return &WechatQRService{
		sessionService: sessionService,
		kvService:      kvService,
	}
}

// WechatAccessTokenResponse 微信access_token响应结构
type WechatAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// WechatQRTicketResponse 微信临时二维码ticket响应结构
type WechatQRTicketResponse struct {
	Ticket   string `json:"ticket"`
	ExpireIn int    `json:"expire_seconds"`
	URL      string `json:"url"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

// QRCodeRequest 二维码请求参数
type QRCodeRequest struct {
	ActionName    string       `json:"action_name"`    // QR_STR_SCENE
	ExpireSeconds int          `json:"expire_seconds"` // 过期时间（秒）
	ActionInfo    QRActionInfo `json:"action_info"`    // 场景信息
}

// QRActionInfo 二维码场景信息
type QRActionInfo struct {
	Scene QRScene `json:"scene"`
}

// QRScene 二维码场景
type QRScene struct {
	SceneStr string `json:"scene_str"` // 场景值字符串
}

// GenerateQRCode 生成微信临时二维码
func (s *WechatQRService) GenerateQRCode() (*models.WechatLoginData, error) {
	// 1. 生成唯一的场景值
	qrScene, err := s.generateUniqueScene()
	if err != nil {
		return nil, fmt.Errorf("生成场景值失败: %v", err)
	}

	// 2. 生成会话ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("生成会话ID失败: %v", err)
	}

	// 3. 获取微信access_token
	accessToken, err := s.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取微信access_token失败: %v", err)
	}

	// 4. 调用微信API生成临时二维码
	ticket, err := s.createTempQRCode(accessToken, qrScene)
	if err != nil {
		return nil, fmt.Errorf("创建临时二维码失败: %v", err)
	}

	// 5. 生成二维码图片URL
	qrURL := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", ticket)

	// 6. 保存会话信息到Redis
	expiresIn := config.AppConfig.Wechat.QRExpiresIn
	if expiresIn <= 0 {
		expiresIn = constants.QRCodeDefaultExpireTime
	}

	loginSession := &models.LoginSession{
		SessionID: sessionID,
		QRScene:   qrScene,
		Status:    constants.SessionStatusPending,
		UserID:    0,
		OpenID:    "",
		CreatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	err = s.sessionService.SaveSession(loginSession)
	if err != nil {
		return nil, fmt.Errorf("保存会话信息失败: %v", err)
	}

	// 7. 返回结果
	return &models.WechatLoginData{
		QRUrl:     qrURL,
		SessionID: sessionID,
		QRScene:   qrScene,
		ExpiresIn: expiresIn,
	}, nil
}

// generateUniqueScene 生成唯一的场景值
func (s *WechatQRService) generateUniqueScene() (string, error) {
	// 使用时间戳 + 随机字符串确保唯一性
	timestamp := time.Now().Unix()

	// 生成8字节随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomBytes)
	scene := fmt.Sprintf("login_%d_%s", timestamp, randomStr)

	// 确保场景值长度符合微信要求（不超过64字节）
	if len(scene) > 64 {
		scene = scene[:64]
	}

	return scene, nil
}

// generateSessionID 生成会话ID
func (s *WechatQRService) generateSessionID() (string, error) {
	// 生成16字节随机数作为会话ID
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

// getAccessToken 获取微信access_token（带缓存）
func (s *WechatQRService) getAccessToken() (string, error) {
	// 1. 先从缓存中获取
	cacheKey := "wechat:access_token"
	cachedData, err := s.kvService.Get(cacheKey)
	if err == nil && cachedData != "" {
		var tokenCache AccessTokenCache
		if err := json.Unmarshal([]byte(cachedData), &tokenCache); err == nil {
			// 检查是否还有5分钟的有效期（提前刷新）
			if time.Now().Add(5 * time.Minute).Before(tokenCache.ExpiresAt) {
				return tokenCache.AccessToken, nil
			}
		}
	}

	// 2. 缓存不存在或已过期，重新获取
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		config.AppConfig.Wechat.AppID,
		config.AppConfig.Wechat.AppSecret,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求微信API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var tokenResp WechatAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("微信API错误: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// 3. 缓存新的access_token
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200 // 默认2小时
	}

	// 提前5分钟过期，确保安全
	cacheExpiresIn := expiresIn - 300
	if cacheExpiresIn <= 0 {
		cacheExpiresIn = 6900 // 2小时减去5分钟
	}

	tokenCache := AccessTokenCache{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(cacheExpiresIn) * time.Second),
	}

	cacheData, err := json.Marshal(tokenCache)
	if err == nil {
		// 缓存access_token，过期时间比微信官方短5分钟
		s.kvService.Set(cacheKey, string(cacheData), time.Duration(cacheExpiresIn)*time.Second)
	}

	return tokenResp.AccessToken, nil
}

// GetAccessToken 获取access_token（公共方法，用于调试）
func (s *WechatQRService) GetAccessToken() (string, error) {
	return s.getAccessToken()
}

// GetAccessTokenStatus 获取access_token状态信息（用于调试）
func (s *WechatQRService) GetAccessTokenStatus() (map[string]interface{}, error) {
	cacheKey := "wechat:access_token"
	cachedData, err := s.kvService.Get(cacheKey)

	status := map[string]interface{}{
		"cached":     false,
		"valid":      false,
		"expires_in": 0,
	}

	if err == nil && cachedData != "" {
		var tokenCache AccessTokenCache
		if err := json.Unmarshal([]byte(cachedData), &tokenCache); err == nil {
			status["cached"] = true
			status["access_token"] = tokenCache.AccessToken[:10] + "..." // 只显示前10位

			now := time.Now()
			if now.Before(tokenCache.ExpiresAt) {
				status["valid"] = true
				status["expires_in"] = int(tokenCache.ExpiresAt.Sub(now).Seconds())
				status["expires_at"] = tokenCache.ExpiresAt.Format("2006-01-02 15:04:05")
			} else {
				status["expired_at"] = tokenCache.ExpiresAt.Format("2006-01-02 15:04:05")
			}
		}
	}

	return status, nil
}

// createTempQRCode 创建临时二维码
func (s *WechatQRService) createTempQRCode(accessToken, scene string) (string, error) {
	// 构建请求参数
	expiresIn := config.AppConfig.Wechat.QRExpiresIn
	if expiresIn <= 0 {
		expiresIn = constants.QRCodeDefaultExpireTime
	}

	qrRequest := QRCodeRequest{
		ActionName:    "QR_STR_SCENE",
		ExpireSeconds: expiresIn,
		ActionInfo: QRActionInfo{
			Scene: QRScene{
				SceneStr: scene,
			},
		},
	}

	requestBody, err := json.Marshal(qrRequest)
	if err != nil {
		return "", err
	}

	// 调用微信API
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", accessToken)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var qrResp WechatQRTicketResponse
	if err := json.Unmarshal(body, &qrResp); err != nil {
		return "", err
	}

	if qrResp.ErrCode != 0 {
		return "", fmt.Errorf("创建二维码失败: %d - %s", qrResp.ErrCode, qrResp.ErrMsg)
	}

	return qrResp.Ticket, nil
}

// GetSession 根据会话ID获取会话信息（委托给SessionService）
func (s *WechatQRService) GetSession(sessionID string) (*models.LoginSession, error) {
	return s.sessionService.GetSession(sessionID)
}

// GetSessionByScene 根据场景值获取会话信息（委托给SessionService）
func (s *WechatQRService) GetSessionByScene(qrScene string) (*models.LoginSession, error) {
	return s.sessionService.GetSessionByScene(qrScene)
}

// UpdateSessionStatus 更新会话状态（委托给SessionService）
func (s *WechatQRService) UpdateSessionStatus(sessionID string, status string, userID uint, openID string) error {
	return s.sessionService.UpdateSessionStatus(sessionID, status, userID, openID)
}

// CleanExpiredSessions 清理过期会话（委托给SessionService）
func (s *WechatQRService) CleanExpiredSessions() error {
	return s.sessionService.CleanExpiredSessions()
}

// ValidateScene 验证场景值格式
func (s *WechatQRService) ValidateScene(scene string) bool {
	// 场景值应该以login_开头，包含时间戳和随机字符串
	return strings.HasPrefix(scene, "login_") && len(scene) <= 64
}
