package constants

// 微信登录相关常量
const (
	// Redis Key 前缀
	WechatLoginSessionPrefix = "wechat_login:session:"

	// 会话状态
	SessionStatusPending = "pending" // 等待扫码
	SessionStatusSuccess = "success" // 登录成功
	SessionStatusExpired = "expired" // 已过期

	// 默认过期时间（秒）
	SessionDefaultExpireTime = 600 // 10分钟
	QRCodeDefaultExpireTime  = 600 // 10分钟

	// WebSocket消息类型
	WSMessageTypeLoginResult = "login_result"

	// 登录类型
	LoginTypeWechat = "wechat"
	LoginTypePhone  = "phone"
	LoginTypeBoth   = "both"
)
