package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"project/config"
	"project/models"
	"project/repository"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JWTHeader JWT 头部
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTPayload JWT 负载
type JWTPayload struct {
	UserID string `json:"userId"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// UserService 用户服务接口
type UserService interface {
	Login(phone, code string) (*models.UserData, error)
	ValidateToken(token string) (*models.User, error)
	SetSMSCode(phone, code string) error
}

// userService 用户服务实现
type userService struct {
	userRepo  repository.UserRepository
	kvService KVService
	jwtSecret string
}

// NewUserService 创建用户服务实例
func NewUserService(jwtSecret string, redisConfig config.RedisConfig) UserService {
	return &userService{
		userRepo:  repository.NewUserRepository(),
		kvService: NewKVService(redisConfig),
		jwtSecret: jwtSecret,
	}
}

// Login 用户登录
func (s *userService) Login(phone, code string) (*models.UserData, error) {
	// 验证手机号格式
	if !s.isValidPhone(phone) {
		return nil, fmt.Errorf("无效的手机号码")
	}

	// 验证验证码格式
	if !s.isValidCode(code) {
		return nil, fmt.Errorf("验证码格式错误")
	}

	// 从 KV 存储中获取验证码
	storedCode, err := s.kvService.Get(fmt.Sprintf("sms:%s", phone))
	if err != nil {
		return nil, fmt.Errorf("获取验证码失败: %v", err)
	}

	if storedCode == "" || storedCode != code {
		return nil, fmt.Errorf("验证码错误或已过期")
	}

	// 查询用户是否存在
	user, err := s.userRepo.GetUserByPhone(phone)
	if err != nil {
		// 用户不存在，创建新用户
		nickname := fmt.Sprintf("用户%s", phone[7:]) // 使用手机号后4位作为昵称
		user, err = s.userRepo.CreateUser(phone, nickname)
		if err != nil {
			return nil, fmt.Errorf("创建用户失败: %v", err)
		}
	} else {
		// 用户存在，更新登录时间
		err = s.userRepo.UpdateLastLoginTime(user.ID)
		if err != nil {
			return nil, fmt.Errorf("更新登录时间失败: %v", err)
		}

		// 重新获取用户信息（确保获取到更新后的时间）
		user, err = s.userRepo.GetUserByPhone(phone)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %v", err)
		}
	}

	// 生成 JWT token
	token, err := s.generateToken(strconv.Itoa(int(user.ID)))
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %v", err)
	}

	// 删除验证码
	err = s.kvService.Delete(fmt.Sprintf("sms:%s", phone))
	if err != nil {
		// 删除失败不影响登录流程，只记录错误
		fmt.Printf("删除验证码失败: %v\n", err)
	}

	// 返回用户数据
	userData := &models.UserData{
		Token: token,
		User:  user,
	}

	return userData, nil
}

// ValidateToken 验证JWT token
func (s *userService) ValidateToken(token string) (*models.User, error) {
	// 去除 Bearer 前缀
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	// 验证 token
	payload, err := s.validateToken(token)
	if err != nil {
		return nil, fmt.Errorf("无效的token: %v", err)
	}

	// 获取用户信息
	userID, err := strconv.ParseUint(payload.UserID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %v", err)
	}

	user, err := s.userRepo.GetUserByID(uint(userID))
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	return user, nil
}

// isValidPhone 验证手机号格式（中国大陆）
func (s *userService) isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// isValidCode 验证验证码格式（6位数字）
func (s *userService) isValidCode(code string) bool {
	codeRegex := regexp.MustCompile(`^\d{6}$`)
	return codeRegex.MatchString(code)
}

// SetSMSCode 设置短信验证码（用于测试或与短信服务集成）
func (s *userService) SetSMSCode(phone, code string) error {
	key := fmt.Sprintf("sms:%s", phone)
	// 设置5分钟过期时间
	return s.kvService.Set(key, code, 5*time.Minute)
}

// generateToken 生成 JWT token
func (s *userService) generateToken(userID string) (string, error) {
	// 创建头部
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	// 创建负载
	now := time.Now().Unix()
	payload := JWTPayload{
		UserID: userID,
		Exp:    now + (7 * 24 * 60 * 60), // 7天过期
		Iat:    now,
	}

	// Base64URL 编码
	headerEncoded, err := s.base64URLEncode(header)
	if err != nil {
		return "", err
	}

	payloadEncoded, err := s.base64URLEncode(payload)
	if err != nil {
		return "", err
	}

	// 生成签名
	signature, err := s.generateSignature(headerEncoded + "." + payloadEncoded)
	if err != nil {
		return "", err
	}

	// 组合最终的 token
	token := fmt.Sprintf("%s.%s.%s", headerEncoded, payloadEncoded, signature)
	return token, nil
}

// validateToken 验证 JWT token
func (s *userService) validateToken(token string) (*JWTPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signature := parts[2]

	// 验证签名
	expectedSignature, err := s.generateSignature(headerEncoded + "." + payloadEncoded)
	if err != nil {
		return nil, err
	}

	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// 解码负载
	payload, err := s.base64URLDecode(payloadEncoded)
	if err != nil {
		return nil, err
	}

	var jwtPayload JWTPayload
	if err := json.Unmarshal(payload, &jwtPayload); err != nil {
		return nil, err
	}

	// 检查过期时间
	if time.Now().Unix() > jwtPayload.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &jwtPayload, nil
}

// base64URLEncode Base64URL 编码
func (s *userService) base64URLEncode(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.EncodeToString(jsonData)
	// 移除填充字符
	encoded = strings.TrimRight(encoded, "=")
	return encoded, nil
}

// base64URLDecode Base64URL 解码
func (s *userService) base64URLDecode(data string) ([]byte, error) {
	// 添加必要的填充字符
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	return base64.URLEncoding.DecodeString(data)
}

// generateSignature 生成 HMAC-SHA256 签名
func (s *userService) generateSignature(data string) (string, error) {
	h := hmac.New(sha256.New, []byte(s.jwtSecret))
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	// 移除填充字符
	signature = strings.TrimRight(signature, "=")
	return signature, nil
}
