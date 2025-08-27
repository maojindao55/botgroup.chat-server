package services

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"project/src/config"
	"project/src/constants"
	"sort"
	"strings"
	"time"
)

// WechatCallbackService 微信回调服务
type WechatCallbackService struct {
	sessionService *SessionService
	userService    UserService // 假设已有用户服务
}

// NewWechatCallbackService 创建微信回调服务实例
func NewWechatCallbackService(sessionService *SessionService, userService UserService) *WechatCallbackService {
	return &WechatCallbackService{
		sessionService: sessionService,
		userService:    userService,
	}
}

// WechatMessage 微信消息结构（XML格式）
type WechatMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`   // 开发者微信号
	FromUserName string   `xml:"FromUserName"` // 发送方帐号（OpenID）
	CreateTime   int64    `xml:"CreateTime"`   // 消息创建时间（整型）
	MsgType      string   `xml:"MsgType"`      // 消息类型
	Event        string   `xml:"Event"`        // 事件类型（当MsgType为event时）
	EventKey     string   `xml:"EventKey"`     // 事件KEY值（扫码关注时为qr_scene_str）
	Ticket       string   `xml:"Ticket"`       // 二维码的ticket，可用来换取二维码图片
}

// WechatReplyMessage 微信回复消息结构
type WechatReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

// VerifySignature 验证微信签名
func (s *WechatCallbackService) VerifySignature(signature, timestamp, nonce string) bool {
	token := config.AppConfig.Wechat.Token

	// 1. 将token、timestamp、nonce三个参数进行字典序排序
	strs := []string{token, timestamp, nonce}
	sort.Strings(strs)

	// 2. 将三个参数字符串拼接成一个字符串进行sha1加密
	str := strings.Join(strs, "")
	h := sha1.New()
	h.Write([]byte(str))
	encrypted := fmt.Sprintf("%x", h.Sum(nil))

	// 3. 开发者获得加密后的字符串可与signature对比，标识该请求来源于微信
	return encrypted == signature
}

// ParseMessage 解析微信XML消息
func (s *WechatCallbackService) ParseMessage(body io.Reader) (*WechatMessage, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("读取消息体失败: %v", err)
	}

	var msg WechatMessage
	if err := xml.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("解析XML消息失败: %v", err)
	}

	return &msg, nil
}

// HandleSubscribeEvent 处理关注事件和扫码事件
func (s *WechatCallbackService) HandleSubscribeEvent(msg *WechatMessage) (*WechatReplyMessage, error) {
	// 检查是否有场景值
	if msg.EventKey == "" {
		// 普通关注事件，没有场景值，返回欢迎消息
		return s.createWelcomeReply(msg), nil
	}

	// 直接使用场景值（微信官方不会添加前缀）
	qrScene := msg.EventKey

	// 验证场景值格式
	if !s.validateSceneFormat(qrScene) {
		return s.createErrorReply(msg, "无效的场景值"), nil
	}

	// 根据场景值查找对应的登录会话
	session, err := s.sessionService.GetSessionByScene(qrScene)
	if err != nil {
		return s.createErrorReply(msg, "未找到对应的登录会话"), nil
	}

	// 检查会话状态
	if session.Status != constants.SessionStatusPending {
		return s.createErrorReply(msg, "会话已失效或已完成"), nil
	}

	// 处理用户登录逻辑
	userID, err := s.handleUserLogin(msg.FromUserName, qrScene)
	if err != nil {
		return s.createErrorReply(msg, "登录处理失败"), fmt.Errorf("处理用户登录失败: %v", err)
	}

	// 更新会话状态为成功
	err = s.sessionService.UpdateSessionByScene(qrScene, constants.SessionStatusSuccess, userID, msg.FromUserName)
	if err != nil {
		return s.createErrorReply(msg, "更新会话状态失败"), fmt.Errorf("更新会话状态失败: %v", err)
	}

	// 返回登录成功回复
	return s.createLoginSuccessReply(msg), nil
}

// handleUserLogin 处理用户登录逻辑
func (s *WechatCallbackService) handleUserLogin(openID, qrScene string) (uint, error) {
	// 统一使用 UserService.LoginWithWechat 处理所有情况
	// 该方法会自动处理新用户创建和老用户更新
	nickname := "微信用户" // 可以从微信API获取真实昵称
	avatarURL := ""    // 可以从微信API获取真实头像

	userData, err := s.userService.LoginWithWechat(openID, nickname, avatarURL, qrScene)
	if err != nil {
		return 0, fmt.Errorf("处理微信用户登录失败: %v", err)
	}

	return userData.User.ID, nil
}

// validateSceneFormat 验证场景值格式
func (s *WechatCallbackService) validateSceneFormat(scene string) bool {
	// 场景值应该以login_开头，包含时间戳和随机字符串
	return strings.HasPrefix(scene, "login_") && len(scene) <= 64
}

// createWelcomeReply 创建欢迎回复消息
func (s *WechatCallbackService) createWelcomeReply(msg *WechatMessage) *WechatReplyMessage {
	return &WechatReplyMessage{
		ToUserName:   msg.FromUserName,
		FromUserName: msg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      "欢迎关注！感谢您的支持！",
	}
}

// createLoginSuccessReply 创建登录成功回复消息
func (s *WechatCallbackService) createLoginSuccessReply(msg *WechatMessage) *WechatReplyMessage {
	return &WechatReplyMessage{
		ToUserName:   msg.FromUserName,
		FromUserName: msg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      "🎉 登录成功！您可以返回网页继续操作了。",
	}
}

// createErrorReply 创建错误回复消息
func (s *WechatCallbackService) createErrorReply(msg *WechatMessage, errorMsg string) *WechatReplyMessage {
	return &WechatReplyMessage{
		ToUserName:   msg.FromUserName,
		FromUserName: msg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      fmt.Sprintf("❌ %s，请重新尝试扫码登录。", errorMsg),
	}
}

// FormatReplyXML 格式化回复XML
func (s *WechatCallbackService) FormatReplyXML(reply *WechatReplyMessage) (string, error) {
	data, err := xml.Marshal(reply)
	if err != nil {
		return "", fmt.Errorf("格式化回复XML失败: %v", err)
	}

	// 添加XML声明
	xmlStr := xml.Header + string(data)
	return xmlStr, nil
}

// HandleMessage 处理微信消息的主入口
func (s *WechatCallbackService) HandleMessage(body io.Reader) (string, error) {
	// 解析消息
	msg, err := s.ParseMessage(body)
	if err != nil {
		return "", fmt.Errorf("解析消息失败: %v", err)
	}

	var reply *WechatReplyMessage

	// 打印 msg 的完整信息 - JSON 格式
	if msgJSON, err := json.MarshalIndent(msg, "", "  "); err == nil {
		fmt.Println("=== 微信消息 JSON 格式 ===")
		fmt.Println(string(msgJSON))
	} else {
		fmt.Printf("JSON 序列化失败: %v\n", err)
	}

	// 打印 msg 的详细字段信息
	fmt.Println("=== 微信消息详细字段 ===")
	fmt.Printf("ToUserName: %s\n", msg.ToUserName)
	fmt.Printf("FromUserName: %s\n", msg.FromUserName)
	fmt.Printf("CreateTime: %d\n", msg.CreateTime)
	fmt.Printf("MsgType: %s\n", msg.MsgType)
	fmt.Printf("Event: %s\n", msg.Event)
	fmt.Printf("EventKey: %s\n", msg.EventKey)
	fmt.Printf("Ticket: %s\n", msg.Ticket)
	fmt.Println("=========================")
	// 根据消息类型处理
	switch msg.MsgType {
	case "event":
		switch msg.Event {
		case "subscribe", "SCAN":
			// 处理关注事件和扫码事件
			reply, err = s.HandleSubscribeEvent(msg)
			if err != nil {
				return "", fmt.Errorf("处理事件失败: %v", err)
			}
		default:
			// 其他事件暂不处理
			return "success", nil
		}
	case "text":
		// 处理文本消息（可以实现简单的自动回复）
		reply = s.createWelcomeReply(msg)
	default:
		// 其他消息类型暂不处理
		return "success", nil
	}

	// 格式化回复XML
	if reply != nil {
		return s.FormatReplyXML(reply)
	}

	return "success", nil
}
