# 微信登录环境变量配置示例

## 环境变量配置

请在 `.env.api` 文件中添加以下微信相关配置：

```bash
# ========== 微信公众号配置 ==========
# 从微信公众平台 (mp.weixin.qq.com) 获取
WECHAT_APP_ID=wx1234567890abcdef          # 公众号AppID
WECHAT_APP_SECRET=1234567890abcdef12345678 # 公众号AppSecret
WECHAT_TOKEN=your_wechat_token_here        # 自定义Token，用于验证微信消息
WECHAT_CALLBACK_URL=https://your-domain.com/api/auth/wechat/callback  # 微信回调URL

# 可选配置（有默认值）
WECHAT_QR_EXPIRES_IN=600      # 二维码过期时间（秒），默认600秒
WECHAT_SESSION_EXPIRES_IN=600 # 会话过期时间（秒），默认600秒

# ========== WebSocket 配置 ==========
# 可选配置（有默认值）
WS_READ_BUFFER_SIZE=1024   # 读取缓冲区大小，默认1024
WS_WRITE_BUFFER_SIZE=1024  # 写入缓冲区大小，默认1024
WS_CHECK_ORIGIN=true       # 是否检查源，默认true
```

## 配置说明

### 微信公众号配置

1. **WECHAT_APP_ID**: 从微信公众平台获取的AppID
2. **WECHAT_APP_SECRET**: 从微信公众平台获取的AppSecret
3. **WECHAT_TOKEN**: 自定义的Token，用于验证微信服务器的消息
4. **WECHAT_CALLBACK_URL**: 微信事件回调的完整URL

### WebSocket配置

1. **WS_READ_BUFFER_SIZE**: WebSocket读取缓冲区大小
2. **WS_WRITE_BUFFER_SIZE**: WebSocket写入缓冲区大小
3. **WS_CHECK_ORIGIN**: 是否检查WebSocket连接的源

## 微信公众平台配置

1. 登录微信公众平台 (mp.weixin.qq.com)
2. 在"开发 > 基本配置"中获取AppID和AppSecret
3. 设置服务器配置：
   - URL: `https://your-domain.com/api/auth/wechat/callback`
   - Token: 与`WECHAT_TOKEN`环境变量一致
   - EncodingAESKey: 可选，暂不使用
4. 配置IP白名单（如果需要）

## 注意事项

- 回调URL必须使用HTTPS
- 域名需要在微信公众平台配置白名单
- Token用于验证消息来源，请设置复杂字符串
- 生产环境请使用强密码和安全的配置