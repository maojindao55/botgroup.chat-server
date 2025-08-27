package api

import (
	"fmt"
	"net/http"
	"project/src/config"
	"project/src/models"
	"project/src/services"

	"github.com/gin-gonic/gin"
)

// WechatQRCodeHandler 生成微信二维码接口
func WechatQRCodeHandler(c *gin.Context) {
	var req models.WechatLoginRequest

	// 解析JSON请求体（可选参数）
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有JSON体，使用默认值
		req = models.WechatLoginRequest{}
	}

	// 创建Redis服务
	kvService := services.NewKVService(config.AppConfig.Redis)

	// 创建会话服务
	sessionService := services.NewSessionService(kvService)

	// 创建微信二维码服务
	qrService := services.NewWechatQRService(sessionService, kvService)

	// 生成二维码
	qrData, err := qrService.GenerateQRCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.WechatLoginResponse{
			Success: false,
			Message: "生成二维码失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, models.WechatLoginResponse{
		Success: true,
		Message: "二维码生成成功",
		Data:    qrData,
	})
}

// WechatCallbackHandler 微信回调处理接口
func WechatCallbackHandler(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		// 服务器验证
		handleWechatVerification(c)
	case http.MethodPost:
		// 事件处理
		handleWechatCallback(c)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "不支持的请求方法",
		})
	}
}

// handleWechatVerification 处理微信服务器验证
func handleWechatVerification(c *gin.Context) {
	// 获取验证参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
	callbackService := services.NewWechatCallbackService(sessionService, userService)

	// 验证签名
	if !callbackService.VerifySignature(signature, timestamp, nonce) {
		c.String(http.StatusUnauthorized, "签名验证失败")
		return
	}

	// 验证成功，返回echostr
	c.String(http.StatusOK, echostr)
}

// handleWechatCallback 处理微信事件回调
func handleWechatCallback(c *gin.Context) {
	// 获取验证参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
	callbackService := services.NewWechatCallbackService(sessionService, userService)

	// 验证签名
	if !callbackService.VerifySignature(signature, timestamp, nonce) {
		c.String(http.StatusUnauthorized, "签名验证失败")
		return
	}

	// 读取请求体
	body := c.Request.Body
	defer body.Close()

	// 处理消息
	replyXML, err := callbackService.HandleMessage(body)
	fmt.Println("replyXML", replyXML)
	if err != nil {
		c.String(http.StatusInternalServerError, "处理消息失败: "+err.Error())
		return
	}

	// 返回回复消息
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, replyXML)
}

// WechatLoginStatusHandler 查询微信登录状态接口
func WechatLoginStatusHandler(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少会话ID",
		})
		return
	}

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)

	// 获取会话信息
	session, err := sessionService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "会话不存在或已过期",
		})
		return
	}

	// 根据会话状态返回不同响应
	switch session.Status {
	case "pending":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  "pending",
			"message": "等待扫码",
		})
	case "success":
		// 获取用户信息
		userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
		user, err := userService.GetUserByID(session.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取用户信息失败",
			})
			return
		}

		// 生成新的JWT token
		token, expiresIn, err := userService.GenerateJWTToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "生成token失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  "success",
			"message": "登录成功",
			"data": gin.H{
				"user": gin.H{
					"id":       user.ID,
					"nickname": user.Nickname,
					"avatar":   user.AvatarURL,
				},
				"token":      token,
				"expires_in": expiresIn,
			},
		})
	case "expired":
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"status":  "expired",
			"message": "会话已过期，请重新扫码",
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"status":  "unknown",
			"message": "未知状态",
		})
	}
}

// WebSocketHandler WebSocket连接处理接口
func WebSocketHandler(c *gin.Context) {
	// 创建WebSocket服务
	wsService := services.NewWebSocketService()

	// 处理WebSocket升级
	wsService.HandleWebSocket(c.Writer, c.Request)
}

// WechatLoginTestHandler 测试微信登录流程（仅开发环境使用）
func WechatLoginTestHandler(c *gin.Context) {
	// 检查是否为开发环境
	if config.AppConfig.Server.Port != "8080" { // 生产环境通常不使用8080端口
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "此接口仅在开发环境可用",
		})
		return
	}

	// 获取参数
	sessionID := c.Query("session_id")
	openID := c.Query("openid")

	if sessionID == "" || openID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少必要参数: session_id 和 openid",
		})
		return
	}

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)

	// 模拟微信登录
	userData, err := userService.LoginWithWechat(openID, "测试用户", "", "test_scene")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "模拟登录失败: " + err.Error(),
		})
		return
	}

	// 更新会话状态
	err = sessionService.UpdateSessionStatus(sessionID, "success", userData.User.ID, openID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新会话失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模拟登录成功",
		"data": gin.H{
			"user_id":    userData.User.ID,
			"session_id": sessionID,
		},
	})
}

// WechatTokenDebugHandler 微信Token调试接口（仅开发环境）
func WechatTokenDebugHandler(c *gin.Context) {
	// 检查是否为开发环境
	if config.AppConfig.Server.Port != "8080" { // 生产环境通常不使用8080端口
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "此接口仅在开发环境可用",
		})
		return
	}

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)
	qrService := services.NewWechatQRService(sessionService, kvService)

	// 获取access_token状态
	tokenStatus, err := qrService.GetAccessTokenStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取Token状态失败: " + err.Error(),
		})
		return
	}

	// 尝试获取新的access_token（如果缓存无效）
	if !tokenStatus["valid"].(bool) {
		_, err := qrService.GetAccessToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":      false,
				"message":      "获取新Token失败: " + err.Error(),
				"token_status": tokenStatus,
			})
			return
		}

		// 重新获取状态
		tokenStatus, _ = qrService.GetAccessTokenStatus()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token状态查询成功",
		"data":    tokenStatus,
	})
}

// WechatCallbackSimulateHandler 模拟微信回调接口（仅开发环境）
func WechatCallbackSimulateHandler(c *gin.Context) {
	// 检查是否为开发环境
	if config.AppConfig.Server.Port != "8080" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "此接口仅在开发环境可用",
		})
		return
	}

	var req struct {
		QRScene  string `json:"qr_scene" binding:"required"`
		OpenID   string `json:"openid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.OpenID == "" {
		req.OpenID = "test_openid_" + req.QRScene
	}
	if req.Nickname == "" {
		req.Nickname = "测试用户"
	}

	// 创建服务
	kvService := services.NewKVService(config.AppConfig.Redis)
	sessionService := services.NewSessionService(kvService)
	userService := services.NewUserService(config.AppConfig.JWTSecret, config.AppConfig.Redis)
	wsService := services.NewWebSocketService()

	// 查找会话
	session, err := sessionService.GetSessionByScene(req.QRScene)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "未找到对应的登录会话: " + err.Error(),
		})
		return
	}

	if session.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "会话状态不正确，当前状态: " + session.Status,
		})
		return
	}

	// 执行登录
	userData, err := userService.LoginWithWechat(req.OpenID, req.Nickname, req.Avatar, req.QRScene)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "登录失败: " + err.Error(),
		})
		return
	}

	// 更新会话状态
	err = sessionService.UpdateSessionStatus(session.SessionID, "success", userData.User.ID, req.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新会话状态失败: " + err.Error(),
		})
		return
	}

	// 通过WebSocket通知前端
	userInfo := &models.UserInfo{
		UserID:    userData.User.ID,
		Nickname:  userData.User.Nickname,
		AvatarURL: userData.User.AvatarURL,
		LoginType: "wechat",
	}

	err = wsService.NotifyLoginSuccess(session.SessionID, userInfo, userData.Token, 604800)
	if err != nil {
		// WebSocket通知失败不影响主流程
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "模拟登录成功，但WebSocket通知失败",
			"data": gin.H{
				"user_id":    userData.User.ID,
				"session_id": session.SessionID,
				"token":      userData.Token,
				"ws_error":   err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模拟微信回调成功",
		"data": gin.H{
			"user_id":    userData.User.ID,
			"session_id": session.SessionID,
			"token":      userData.Token,
			"openid":     req.OpenID,
		},
	})
}
