# 微信扫码登录 API 文档

## API 概述

微信扫码登录系统提供了一套完整的API接口，支持生成二维码、处理微信回调、查询登录状态和WebSocket实时通知。

## 接口列表

### 1. 生成微信登录二维码

**接口地址：** `POST /api/auth/wechat/qr-code`

**请求参数：**
```json
{
  "redirect_uri": "https://your-app.com/login-success" // 可选，登录成功后的跳转地址
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "二维码生成成功",
  "data": {
    "qr_url": "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=xxx",
    "session_id": "abc123def456789",
    "qr_scene": "login_1640000000_abcd1234",
    "expires_in": 600
  }
}
```

**限流规则：** 每个IP每分钟最多10次请求

---

### 2. 查询登录状态

**接口地址：** `GET /api/auth/wechat/status/{session_id}`

**路径参数：**
- `session_id`: 会话ID（从生成二维码接口获取）

**响应示例：**

等待扫码：
```json
{
  "success": true,
  "status": "pending",
  "message": "等待扫码"
}
```

登录成功：
```json
{
  "success": true,
  "status": "success",
  "message": "登录成功",
  "data": {
    "user": {
      "id": 123,
      "nickname": "微信用户",
      "avatar": "https://example.com/avatar.jpg"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 604800
  }
}
```

会话过期：
```json
{
  "success": false,
  "status": "expired",
  "message": "会话已过期，请重新扫码"
}
```

**限流规则：** 每个IP每分钟最多60次请求

---

### 3. WebSocket 实时通知

**接口地址：** `WebSocket /ws/auth/{session_id}`

**连接示例：**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/auth/abc123def456789');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};
```

**消息格式：**

登录成功通知：
```json
{
  "type": "login_result",
  "data": {
    "status": "success",
    "message": "登录成功",
    "user_info": {
      "user_id": 123,
      "nickname": "微信用户",
      "avatar_url": "https://example.com/avatar.jpg",
      "login_type": "wechat"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 604800
  }
}
```

登录失败通知：
```json
{
  "type": "login_result",
  "data": {
    "status": "failed",
    "message": "登录失败：用户取消授权",
    "user_info": null,
    "token": "",
    "expires_in": 0
  }
}
```

心跳消息：
```json
{
  "type": "ping",
  "data": "heartbeat"
}
```

---

### 4. 微信回调接口（微信服务器专用）

**接口地址：** 
- `GET /api/auth/wechat/callback` - 服务器验证
- `POST /api/auth/wechat/callback` - 事件回调

这些接口由微信服务器调用，开发者无需直接使用。

**限流规则：** 每个IP每分钟最多100次请求

---

### 5. 测试接口（仅开发环境）

**接口地址：** `GET /api/auth/wechat/test`

**查询参数：**
- `session_id`: 会话ID
- `openid`: 模拟的微信OpenID

**响应示例：**
```json
{
  "success": true,
  "message": "模拟登录成功",
  "data": {
    "user_id": 123,
    "session_id": "abc123def456789"
  }
}
```

**注意：** 此接口仅在开发环境（端口8080）可用。

---

## 使用流程

### 标准登录流程

1. **前端调用生成二维码接口**
   ```javascript
   const response = await fetch('/api/auth/wechat/qr-code', {
     method: 'POST',
     headers: { 'Content-Type': 'application/json' },
     body: JSON.stringify({})
   });
   const { data } = await response.json();
   ```

2. **显示二维码并建立WebSocket连接**
   ```javascript
   // 显示二维码
   document.getElementById('qr-image').src = data.qr_url;
   
   // 建立WebSocket连接
   const ws = new WebSocket(`ws://localhost:8080/ws/auth/${data.session_id}`);
   ws.onmessage = handleLoginResult;
   ```

3. **用户扫码关注公众号**
   - 用户使用微信扫描二维码
   - 关注公众号触发登录流程

4. **接收登录结果**
   ```javascript
   function handleLoginResult(event) {
     const message = JSON.parse(event.data);
     if (message.type === 'login_result') {
       if (message.data.status === 'success') {
         // 登录成功，保存token
         localStorage.setItem('token', message.data.token);
         window.location.href = '/dashboard';
       }
     }
   }
   ```

### 轮询方式（WebSocket的替代方案）

如果不使用WebSocket，可以通过轮询状态接口：

```javascript
function pollLoginStatus(sessionId) {
  const interval = setInterval(async () => {
    const response = await fetch(`/api/auth/wechat/status/${sessionId}`);
    const result = await response.json();
    
    if (result.status === 'success') {
      clearInterval(interval);
      // 登录成功
      localStorage.setItem('token', result.data.token);
      window.location.href = '/dashboard';
    } else if (result.status === 'expired') {
      clearInterval(interval);
      // 二维码过期，需要重新生成
      alert('二维码已过期，请重新获取');
    }
  }, 2000); // 每2秒查询一次
}
```

---

## 错误码说明

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数无效 | 检查请求参数格式 |
| 401 | 签名验证失败 | 检查微信配置 |
| 404 | 会话不存在 | 检查session_id是否正确 |
| 429 | 请求频率过高 | 降低请求频率 |
| 500 | 服务器内部错误 | 联系技术支持 |

---

## 安全注意事项

1. **HTTPS要求：** 生产环境必须使用HTTPS
2. **Token安全：** JWT Token应安全存储，避免泄露
3. **限流保护：** 各接口都有限流保护，防止滥用
4. **签名验证：** 微信回调接口有签名验证机制
5. **会话过期：** 登录会话有时效性，过期需重新登录

---

## 配置要求

### 微信公众号配置

1. **服务器配置：**
   - URL: `https://your-domain.com/api/auth/wechat/callback`
   - Token: 与环境变量 `WECHAT_TOKEN` 一致
   - 消息加解密方式: 明文模式

2. **环境变量：**
   ```bash
   WECHAT_APP_ID=wx1234567890abcdef
   WECHAT_APP_SECRET=abcdef1234567890abcdef1234567890
   WECHAT_TOKEN=your_secure_token
   WECHAT_CALLBACK_URL=https://your-domain.com/api/auth/wechat/callback
   ```

### 系统要求

- Go 1.19+
- Redis 6.0+
- MySQL 8.0+
- 支持WebSocket的Web服务器