# 微信扫码登录功能使用说明

本文档介绍如何在现有项目中集成微信扫码登录功能，同时保留手机号验证码登录方式。

## 功能特性

✅ **微信扫码登录** - 新增功能  
✅ **手机号验证码登录** - 保留现有功能  
✅ **账号绑定管理** - 支持绑定/解绑微信账号  
✅ **统一JWT认证** - 与现有登录系统兼容  
✅ **多登录方式并存** - 用户可选择登录方式  
✅ **实时状态推送** - WebSocket支持  

## 快速开始

### 1. 环境准备

#### 1.1 微信开放平台配置
1. 注册微信开放平台账号：https://open.weixin.qq.com/
2. 创建网站应用
3. 配置授权回调域名
4. 获取AppID和AppSecret

#### 1.2 项目配置
在 `.env.api` 文件中添加微信配置：

```bash
# 微信开放平台配置
WECHAT_APP_ID=your_wechat_app_id
WECHAT_APP_SECRET=your_wechat_app_secret
WECHAT_REDIRECT_URI=https://your-domain.com/api/auth/wechat/callback

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

### 2. 数据库迁移

#### 2.1 扩展现有用户表
```sql
-- 为现有users表添加微信相关字段
ALTER TABLE users ADD COLUMN openid VARCHAR(64) COMMENT '微信OpenID';
ALTER TABLE users ADD COLUMN unionid VARCHAR(64) COMMENT '微信UnionID';
ALTER TABLE users ADD COLUMN wechat_nickname VARCHAR(100) COMMENT '微信昵称';
ALTER TABLE users ADD COLUMN wechat_avatar_url TEXT COMMENT '微信头像URL';
ALTER TABLE users ADD COLUMN wechat_gender TINYINT DEFAULT 0 COMMENT '微信性别';
ALTER TABLE users ADD COLUMN wechat_country VARCHAR(50) COMMENT '微信国家';
ALTER TABLE users ADD COLUMN wechat_province VARCHAR(50) COMMENT '微信省份';
ALTER TABLE users ADD COLUMN wechat_city VARCHAR(50) COMMENT '微信城市';
ALTER TABLE users ADD COLUMN wechat_language VARCHAR(20) COMMENT '微信语言';
ALTER TABLE users ADD COLUMN login_type ENUM('phone', 'wechat', 'both') DEFAULT 'phone' COMMENT '登录方式';
ALTER TABLE users ADD COLUMN wechat_session_key VARCHAR(100) COMMENT '微信会话密钥';
ALTER TABLE users ADD COLUMN wechat_access_token VARCHAR(500) COMMENT '微信访问令牌';
ALTER TABLE users ADD COLUMN wechat_refresh_token VARCHAR(500) COMMENT '微信刷新令牌';
ALTER TABLE users ADD COLUMN wechat_expires_at TIMESTAMP NULL COMMENT '微信令牌过期时间';

-- 添加索引
ALTER TABLE users ADD INDEX idx_openid (openid);
ALTER TABLE users ADD INDEX idx_unionid (unionid);
ALTER TABLE users ADD INDEX idx_login_type (login_type);
```

#### 2.2 创建微信绑定表
```sql
CREATE TABLE user_wechat_bindings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    openid VARCHAR(64) NOT NULL COMMENT '微信OpenID',
    unionid VARCHAR(64) COMMENT '微信UnionID',
    nickname VARCHAR(100) COMMENT '微信昵称',
    avatar_url TEXT COMMENT '微信头像URL',
    gender TINYINT DEFAULT 0 COMMENT '性别 0-未知 1-男 2-女',
    country VARCHAR(50) COMMENT '国家',
    province VARCHAR(50) COMMENT '省份',
    city VARCHAR(50) COMMENT '城市',
    language VARCHAR(20) COMMENT '语言',
    session_key VARCHAR(100) COMMENT '会话密钥',
    access_token VARCHAR(500) COMMENT '访问令牌',
    refresh_token VARCHAR(500) COMMENT '刷新令牌',
    expires_at TIMESTAMP NULL COMMENT '令牌过期时间',
    bind_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_openid (openid),
    INDEX idx_unionid (unionid),
    UNIQUE KEY uk_user_openid (user_id, openid)
);
```

## API 接口使用

### 1. 微信扫码登录

#### 1.1 生成登录二维码
```bash
curl -X POST http://localhost:8080/api/auth/wechat/qr-code \
  -H "Content-Type: application/json" \
  -d '{
    "redirect_uri": "https://your-app.com/login-success",
    "state": "custom_state"
  }'
```

**响应示例：**
```json
{
  "success": true,
  "message": "二维码生成成功",
  "data": {
    "qr_code": "wechat_qr_abc123",
    "qr_url": "https://open.weixin.qq.com/connect/qrconnect?appid=...",
    "session_id": "session_xyz789",
    "expires_in": 300
  }
}
```

#### 1.2 检查登录状态
```bash
curl -X GET http://localhost:8080/api/auth/wechat/status/session_xyz789
```

**响应示例：**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "status": "pending",
    "user_info": null,
    "redirect_url": null
  }
}
```

**状态说明：**
- `pending`: 等待扫码
- `scanned`: 已扫码，等待确认
- `confirmed`: 已确认，登录成功
- `expired`: 已过期
- `failed`: 登录失败

### 2. 账号绑定管理

#### 2.1 绑定微信账号
```bash
curl -X POST http://localhost:8080/api/user/bind-wechat \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "qr_code": "wechat_qr_abc123",
    "session_id": "session_xyz789"
  }'
```

#### 2.2 解绑微信账号
```bash
curl -X POST http://localhost:8080/api/user/unbind-wechat \
  -H "Authorization: Bearer your_jwt_token"
```

### 3. 获取用户信息（扩展）

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer your_jwt_token"
```

**响应示例：**
```json
{
  "success": true,
  "message": "success",
  "data": {
    "id": 1,
    "phone": "13800138000",
    "nickname": "测试用户",
    "avatar_url": "",
    "status": 1,
    "login_type": "both",
    "wechat_info": {
      "openid": "wx_openid_123",
      "nickname": "微信昵称",
      "avatar_url": "https://thirdwx.qlogo.cn/...",
      "gender": 1,
      "country": "中国",
      "province": "广东",
      "city": "深圳"
    },
    "created_at": "2025-03-26T08:39:15Z",
    "updated_at": "2025-03-26T08:39:15Z",
    "last_login_at": "2025-03-26T08:39:15Z"
  }
}
```

## 前端集成示例

### 1. 微信扫码登录组件

```javascript
class WechatLogin {
  constructor() {
    this.sessionId = null;
    this.qrCode = null;
    this.pollingInterval = null;
  }

  // 生成二维码
  async generateQRCode() {
    try {
      const response = await fetch('/api/auth/wechat/qr-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          redirect_uri: window.location.origin + '/login-success',
          state: 'wechat_login'
        })
      });

      const result = await response.json();
      if (result.success) {
        this.sessionId = result.data.session_id;
        this.qrCode = result.data.qr_code;
        
        // 显示二维码
        this.displayQRCode(result.data.qr_url);
        
        // 开始轮询状态
        this.startPolling();
        
        return result.data;
      } else {
        throw new Error(result.message);
      }
    } catch (error) {
      console.error('生成二维码失败:', error);
      throw error;
    }
  }

  // 显示二维码
  displayQRCode(qrUrl) {
    const qrContainer = document.getElementById('wechat-qr-container');
    qrContainer.innerHTML = `
      <div class="qr-code-wrapper">
        <img src="${qrUrl}" alt="微信登录二维码" />
        <p>请使用微信扫描二维码登录</p>
        <div class="qr-status" id="qr-status">等待扫码...</div>
      </div>
    `;
  }

  // 开始轮询状态
  startPolling() {
    this.pollingInterval = setInterval(async () => {
      try {
        const response = await fetch(`/api/auth/wechat/status/${this.sessionId}`);
        const result = await response.json();
        
        if (result.success) {
          this.updateStatus(result.data.status, result.data.user_info);
          
          if (result.data.status === 'confirmed') {
            this.handleLoginSuccess(result.data);
          } else if (result.data.status === 'expired' || result.data.status === 'failed') {
            this.handleLoginFailure(result.data.status);
          }
        }
      } catch (error) {
        console.error('检查状态失败:', error);
      }
    }, 2000); // 每2秒检查一次
  }

  // 更新状态显示
  updateStatus(status, userInfo) {
    const statusElement = document.getElementById('qr-status');
    const statusMap = {
      'pending': '等待扫码...',
      'scanned': '已扫码，请在手机上确认',
      'confirmed': '登录成功！',
      'expired': '二维码已过期',
      'failed': '登录失败'
    };
    
    statusElement.textContent = statusMap[status] || status;
    
    if (userInfo) {
      statusElement.innerHTML += `<br><small>欢迎，${userInfo.nickname}</small>`;
    }
  }

  // 处理登录成功
  handleLoginSuccess(data) {
    clearInterval(this.pollingInterval);
    
    // 保存token和用户信息
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    
    // 显示成功消息
    this.updateStatus('confirmed', data.user_info);
    
    // 跳转到成功页面
    setTimeout(() => {
      window.location.href = data.redirect_url || '/dashboard';
    }, 1500);
  }

  // 处理登录失败
  handleLoginFailure(status) {
    clearInterval(this.pollingInterval);
    this.updateStatus(status);
    
    // 显示重新生成按钮
    const qrContainer = document.getElementById('wechat-qr-container');
    qrContainer.innerHTML += `
      <button onclick="wechatLogin.regenerateQR()" class="btn btn-primary">
        重新生成二维码
      </button>
    `;
  }

  // 重新生成二维码
  async regenerateQR() {
    try {
      await this.generateQRCode();
    } catch (error) {
      console.error('重新生成二维码失败:', error);
    }
  }

  // 停止轮询
  stopPolling() {
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
      this.pollingInterval = null;
    }
  }
}

// 使用示例
const wechatLogin = new WechatLogin();

// 页面加载时生成二维码
document.addEventListener('DOMContentLoaded', () => {
  wechatLogin.generateQRCode();
});

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
  wechatLogin.stopPolling();
});
```

### 2. 账号绑定组件

```javascript
class WechatBinding {
  constructor() {
    this.token = localStorage.getItem('token');
  }

  // 绑定微信账号
  async bindWechat() {
    try {
      // 先生成二维码
      const qrResponse = await fetch('/api/auth/wechat/qr-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          redirect_uri: window.location.origin + '/bind-success',
          state: 'wechat_binding'
        })
      });

      const qrResult = await qrResponse.json();
      if (!qrResult.success) {
        throw new Error(qrResult.message);
      }

      // 显示绑定二维码
      this.displayBindingQR(qrResult.data.qr_url, qrResult.data.session_id);
      
      // 开始轮询绑定状态
      this.pollBindingStatus(qrResult.data.session_id);
      
    } catch (error) {
      console.error('绑定微信账号失败:', error);
      throw error;
    }
  }

  // 显示绑定二维码
  displayBindingQR(qrUrl, sessionId) {
    const container = document.getElementById('binding-container');
    container.innerHTML = `
      <div class="binding-qr-wrapper">
        <h3>绑定微信账号</h3>
        <img src="${qrUrl}" alt="微信绑定二维码" />
        <p>请使用微信扫描二维码完成绑定</p>
        <div class="binding-status" id="binding-status">等待扫码...</div>
        <button onclick="wechatBinding.cancelBinding()" class="btn btn-secondary">
          取消绑定
        </button>
      </div>
    `;
  }

  // 轮询绑定状态
  async pollBindingStatus(sessionId) {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`/api/auth/wechat/status/${sessionId}`);
        const result = await response.json();
        
        if (result.success && result.data.status === 'confirmed') {
          clearInterval(interval);
          await this.completeBinding(sessionId);
        } else if (result.success && (result.data.status === 'expired' || result.data.status === 'failed')) {
          clearInterval(interval);
          this.handleBindingFailure(result.data.status);
        }
      } catch (error) {
        console.error('检查绑定状态失败:', error);
      }
    }, 2000);
  }

  // 完成绑定
  async completeBinding(sessionId) {
    try {
      const response = await fetch('/api/user/bind-wechat', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          session_id: sessionId
        })
      });

      const result = await response.json();
      if (result.success) {
        this.handleBindingSuccess(result.data);
      } else {
        throw new Error(result.message);
      }
    } catch (error) {
      console.error('完成绑定失败:', error);
      this.handleBindingFailure('failed');
    }
  }

  // 处理绑定成功
  handleBindingSuccess(data) {
    const container = document.getElementById('binding-container');
    container.innerHTML = `
      <div class="binding-success">
        <h3>绑定成功！</h3>
        <p>微信账号已成功绑定</p>
        <p>昵称: ${data.wechat_info.nickname}</p>
        <button onclick="location.reload()" class="btn btn-primary">
          刷新页面
        </button>
      </div>
    `;
  }

  // 处理绑定失败
  handleBindingFailure(status) {
    const container = document.getElementById('binding-container');
    container.innerHTML = `
      <div class="binding-failed">
        <h3>绑定失败</h3>
        <p>状态: ${status}</p>
        <button onclick="wechatBinding.bindWechat()" class="btn btn-primary">
          重新绑定
        </button>
      </div>
    `;
  }

  // 取消绑定
  cancelBinding() {
    const container = document.getElementById('binding-container');
    container.innerHTML = `
      <div class="binding-cancelled">
        <h3>已取消绑定</h3>
        <button onclick="wechatBinding.bindWechat()" class="btn btn-primary">
          重新绑定
        </button>
      </div>
    `;
  }

  // 解绑微信账号
  async unbindWechat() {
    try {
      const response = await fetch('/api/user/unbind-wechat', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json',
        }
      });

      const result = await response.json();
      if (result.success) {
        alert('微信账号解绑成功');
        location.reload();
      } else {
        throw new Error(result.message);
      }
    } catch (error) {
      console.error('解绑微信账号失败:', error);
      alert('解绑失败: ' + error.message);
    }
  }
}

// 使用示例
const wechatBinding = new WechatBinding();
```

## 登录方式切换

### 1. 登录页面示例

```html
<!DOCTYPE html>
<html>
<head>
    <title>用户登录</title>
    <style>
        .login-container {
            max-width: 400px;
            margin: 50px auto;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 8px;
        }
        .login-tabs {
            display: flex;
            margin-bottom: 20px;
        }
        .login-tab {
            flex: 1;
            padding: 10px;
            text-align: center;
            cursor: pointer;
            border-bottom: 2px solid transparent;
        }
        .login-tab.active {
            border-bottom-color: #007bff;
            color: #007bff;
        }
        .login-content {
            display: none;
        }
        .login-content.active {
            display: block;
        }
        .qr-code-wrapper {
            text-align: center;
            padding: 20px;
        }
        .qr-code-wrapper img {
            max-width: 200px;
            margin-bottom: 10px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-tabs">
            <div class="login-tab active" onclick="switchTab('phone')">手机号登录</div>
            <div class="login-tab" onclick="switchTab('wechat')">微信登录</div>
        </div>

        <!-- 手机号登录 -->
        <div id="phone-login" class="login-content active">
            <form id="phone-form">
                <div>
                    <label>手机号:</label>
                    <input type="tel" id="phone" required>
                </div>
                <div>
                    <label>验证码:</label>
                    <input type="text" id="code" required>
                    <button type="button" onclick="sendCode()">发送验证码</button>
                </div>
                <button type="submit">登录</button>
            </form>
        </div>

        <!-- 微信登录 -->
        <div id="wechat-login" class="login-content">
            <div id="wechat-qr-container">
                <div class="qr-code-wrapper">
                    <p>正在生成二维码...</p>
                </div>
            </div>
        </div>
    </div>

    <script>
        const wechatLogin = new WechatLogin();

        function switchTab(type) {
            // 切换标签样式
            document.querySelectorAll('.login-tab').forEach(tab => tab.classList.remove('active'));
            event.target.classList.add('active');

            // 切换内容
            document.querySelectorAll('.login-content').forEach(content => content.classList.remove('active'));
            document.getElementById(type + '-login').classList.add('active');

            // 如果是微信登录，生成二维码
            if (type === 'wechat') {
                wechatLogin.generateQRCode();
            } else {
                wechatLogin.stopPolling();
            }
        }

        // 手机号登录相关函数
        async function sendCode() {
            const phone = document.getElementById('phone').value;
            try {
                const response = await fetch('/api/send-code', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ phone })
                });
                const result = await response.json();
                if (result.success) {
                    alert('验证码发送成功: ' + result.data.code);
                } else {
                    alert('发送失败: ' + result.message);
                }
            } catch (error) {
                alert('发送失败: ' + error.message);
            }
        }

        document.getElementById('phone-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const phone = document.getElementById('phone').value;
            const code = document.getElementById('code').value;

            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ phone, code })
                });
                const result = await response.json();
                if (result.success) {
                    localStorage.setItem('token', result.data.token);
                    localStorage.setItem('user', JSON.stringify(result.data.user));
                    alert('登录成功！');
                    window.location.href = '/dashboard';
                } else {
                    alert('登录失败: ' + result.message);
                }
            } catch (error) {
                alert('登录失败: ' + error.message);
            }
        });
    </script>
</body>
</html>
```

## 测试指南

### 1. 功能测试

```bash
# 1. 测试微信二维码生成
curl -X POST http://localhost:8080/api/auth/wechat/qr-code \
  -H "Content-Type: application/json" \
  -d '{"redirect_uri": "http://localhost:3000/success"}'

# 2. 测试状态检查
curl -X GET http://localhost:8080/api/auth/wechat/status/session_id_here

# 3. 测试现有手机号登录（确保兼容性）
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'

curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "code": "123456"}'
```

### 2. 兼容性测试

1. **现有功能验证**
   - 手机号验证码登录正常
   - JWT Token验证正常
   - 用户信息获取正常

2. **新功能验证**
   - 微信二维码生成正常
   - 微信登录流程正常
   - 账号绑定功能正常

3. **数据一致性验证**
   - 用户数据正确扩展
   - 登录方式字段正确更新
   - 微信信息正确存储

## 常见问题

**Q: 微信登录和手机号登录会产生重复用户吗？**
A: 不会。系统会检查微信OpenID是否已存在，如果存在则关联到现有用户，否则创建新用户。

**Q: 用户可以同时使用两种登录方式吗？**
A: 可以。用户可以选择绑定微信账号，绑定后可以使用任意一种方式登录。

**Q: 解绑微信账号后，用户还能登录吗？**
A: 可以。解绑后用户仍可使用手机号验证码登录，login_type会更新为"phone"。

**Q: 微信登录的Token和手机号登录的Token有区别吗？**
A: 没有区别。两种登录方式都使用相同的JWT Token格式和验证机制。

**Q: 如何迁移现有用户数据？**
A: 现有用户数据无需迁移，系统会自动扩展用户表，新字段默认为空值。

## 完成状态

✅ **微信扫码登录功能设计**  
✅ **与现有登录系统集成**  
✅ **数据库结构设计**  
✅ **API接口设计**  
✅ **前端集成示例**  
✅ **配置说明**  
✅ **测试指南**  
✅ **常见问题解答**  

现在您拥有了一个完整的、与现有系统兼容的微信扫码登录功能！🚀 