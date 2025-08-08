package services

import (
	"encoding/json"
	"fmt"
	"project/src/constants"
	"project/src/models"
	"time"
)

// SessionService 会话管理服务
type SessionService struct {
	kvService KVService
}

// NewSessionService 创建会话管理服务实例
func NewSessionService(kvService KVService) *SessionService {
	return &SessionService{
		kvService: kvService,
	}
}

// SaveSession 保存会话信息到Redis
func (s *SessionService) SaveSession(session *models.LoginSession) error {
	key := constants.WechatLoginSessionPrefix + session.SessionID

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话信息失败: %v", err)
	}

	// 设置过期时间
	ttl := time.Duration(session.ExpiresAt-time.Now().Unix()) * time.Second
	if ttl <= 0 {
		return fmt.Errorf("会话已过期")
	}

	return s.kvService.Set(key, string(data), ttl)
}

// GetSession 根据会话ID获取会话信息
func (s *SessionService) GetSession(sessionID string) (*models.LoginSession, error) {
	key := constants.WechatLoginSessionPrefix + sessionID

	data, err := s.kvService.Get(key)
	if err != nil {
		return nil, fmt.Errorf("获取会话信息失败: %v", err)
	}

	var session models.LoginSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("反序列化会话信息失败: %v", err)
	}

	// 检查是否过期
	if time.Now().Unix() > session.ExpiresAt {
		// 更新状态为过期
		session.Status = constants.SessionStatusExpired
		s.SaveSession(&session)
		return &session, nil
	}

	return &session, nil
}

// GetSessionByScene 根据场景值获取会话信息
func (s *SessionService) GetSessionByScene(qrScene string) (*models.LoginSession, error) {
	// 通过模式匹配查找所有会话
	pattern := constants.WechatLoginSessionPrefix + "*"
	keys, err := s.kvService.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("查找会话失败: %v", err)
	}

	for _, key := range keys {
		data, err := s.kvService.Get(key)
		if err != nil {
			continue // 跳过错误的key
		}

		var session models.LoginSession
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			continue // 跳过无效的session
		}

		if session.QRScene == qrScene {
			// 检查是否过期
			if time.Now().Unix() > session.ExpiresAt {
				session.Status = constants.SessionStatusExpired
				s.SaveSession(&session)
			}
			return &session, nil
		}
	}

	return nil, fmt.Errorf("未找到场景值对应的会话")
}

// UpdateSessionStatus 更新会话状态
func (s *SessionService) UpdateSessionStatus(sessionID string, status string, userID uint, openID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("获取会话失败: %v", err)
	}

	// 更新会话信息
	session.Status = status
	session.UserID = userID
	session.OpenID = openID

	return s.SaveSession(session)
}

// UpdateSessionByScene 根据场景值更新会话状态
func (s *SessionService) UpdateSessionByScene(qrScene string, status string, userID uint, openID string) error {
	session, err := s.GetSessionByScene(qrScene)
	if err != nil {
		return fmt.Errorf("获取会话失败: %v", err)
	}

	// 更新会话信息
	session.Status = status
	session.UserID = userID
	session.OpenID = openID

	return s.SaveSession(session)
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(sessionID string) error {
	key := constants.WechatLoginSessionPrefix + sessionID
	return s.kvService.Delete(key)
}

// ExtendSession 延长会话有效期
func (s *SessionService) ExtendSession(sessionID string, extendSeconds int64) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("获取会话失败: %v", err)
	}

	// 延长过期时间
	session.ExpiresAt = time.Now().Unix() + extendSeconds

	return s.SaveSession(session)
}

// CleanExpiredSessions 清理过期会话
func (s *SessionService) CleanExpiredSessions() error {
	pattern := constants.WechatLoginSessionPrefix + "*"
	keys, err := s.kvService.Keys(pattern)
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %v", err)
	}

	now := time.Now().Unix()
	cleanedCount := 0

	for _, key := range keys {
		data, err := s.kvService.Get(key)
		if err != nil {
			continue // 跳过错误的key
		}

		var session models.LoginSession
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			// 数据格式错误，直接删除
			s.kvService.Delete(key)
			cleanedCount++
			continue
		}

		// 删除过期会话
		if now > session.ExpiresAt {
			s.kvService.Delete(key)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		fmt.Printf("清理了 %d 个过期会话\n", cleanedCount)
	}

	return nil
}

// GetAllActiveSessions 获取所有活跃会话
func (s *SessionService) GetAllActiveSessions() ([]*models.LoginSession, error) {
	pattern := constants.WechatLoginSessionPrefix + "*"
	keys, err := s.kvService.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %v", err)
	}

	var sessions []*models.LoginSession
	now := time.Now().Unix()

	for _, key := range keys {
		data, err := s.kvService.Get(key)
		if err != nil {
			continue // 跳过错误的key
		}

		var session models.LoginSession
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			continue // 跳过无效的session
		}

		// 只返回未过期的会话
		if now <= session.ExpiresAt {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

// GetSessionCount 获取会话数量统计
func (s *SessionService) GetSessionCount() (map[string]int, error) {
	sessions, err := s.GetAllActiveSessions()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{
		constants.SessionStatusPending: 0,
		constants.SessionStatusSuccess: 0,
		constants.SessionStatusExpired: 0,
		"total":                        0,
	}

	for _, session := range sessions {
		counts[session.Status]++
		counts["total"]++
	}

	return counts, nil
}

// IsValidSessionStatus 验证会话状态是否有效
func (s *SessionService) IsValidSessionStatus(status string) bool {
	switch status {
	case constants.SessionStatusPending,
		constants.SessionStatusSuccess,
		constants.SessionStatusExpired:
		return true
	default:
		return false
	}
}
