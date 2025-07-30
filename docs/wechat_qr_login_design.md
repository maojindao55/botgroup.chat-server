# 微信扫码登录功能设计文档

## 1. 功能概述

### 1.1 功能描述
在现有手机号验证码登录基础上，新增基于微信开放平台的扫码登录功能，用户可通过扫描二维码完成身份验证和登录，支持多种登录方式并存。

### 1.2 技术栈
- 后端框架：Gin (Go)
- 数据库：MySQL + Redis
- 微信开放平台 API
- WebSocket (用于实时状态推送)
- JWT Token 身份验证（与现有登录系统兼容）

### 1.3 功能特点
- **多登录方式支持**：手机号验证码 + 微信扫码登录
- **统一用户体系**：微信用户与手机号用户数据关联
- **安全性高**：基于微信官方认证 + JWT Token
- **用户体验好**：扫码即登录，无需记忆密码
- **实时反馈**：登录状态实时推送
- **会话管理**：支持多设备登录控制
- **向后兼容**：保留现有登录功能，平滑升级

## 2. 系统架构

### 2.1 整体架构图
```
用户浏览器 <-> 前端页面 <-> 后端API <-> 微信开放平台
                |           |
                |           v
                |        Redis缓存
                |           |
                |           v
                |        MySQL数据库
```

### 2.2 核心组件
1. **二维码生成服务** - 生成微信登录二维码
2. **状态轮询服务** - 检查扫码状态
3. **用户认证服务** - 处理用户登录逻辑
4. **会话管理服务** - 管理用户会话
5. **WebSocket服务** - 实时推送登录状态

## 3. 数据库设计

### 3.1 现有用户表保持不变 (users)
```sql
-- 现有users表结构保持不变，不添加任何微信相关字段
-- 保持原有的手机号验证码登录功能完整性
```

### 3.2 微信用户表 (wechat_users)
```sql
CREATE TABLE wechat_users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uid BIGINT COMMENT '关联用户ID，关联users表的id字段',
    openid VARCHAR(64) UNIQUE NOT NULL COMMENT '微信OpenID',
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
    last_login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后登录时间',
    bind_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_uid (uid),
    INDEX idx_openid (openid),
    INDEX idx_unionid (unionid),
    INDEX idx_bind_at (bind_at),
    FOREIGN KEY (uid) REFERENCES users(id) ON DELETE SET NULL
);
```

### 3.3 用户登录方式表 (user_login_types)
```sql
CREATE TABLE user_login_types (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    login_type ENUM('phone', 'wechat', 'both') DEFAULT 'phone' COMMENT '登录方式',
    wechat_user_id BIGINT COMMENT '关联的微信用户ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_login_type (login_type),
    INDEX idx_wechat_user_id (wechat_user_id),
    UNIQUE KEY uk_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (wechat_user_id) REFERENCES wechat_users(id) ON DELETE SET NULL
);
```

### 3.2 登录会话表 (login_sessions)
```sql
CREATE TABLE login_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id VARCHAR(64) UNIQUE NOT NULL COMMENT '会话ID',
    user_id BIGINT COMMENT '用户ID',
    qr_code VARCHAR(100) COMMENT '二维码标识',
    status ENUM('pending', 'scanned', 'confirmed', 'expired', 'failed') DEFAULT 'pending',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_qr_code (qr_code),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at)
);
```

## 4. API 接口设计

### 4.1 生成登录二维码
```
POST /api/auth/wechat/qr-code
```

**请求参数：**
```json
{
    "redirect_uri": "string", // 登录成功后的跳转地址
    "state": "string"         // 自定义状态参数
}
```

**响应数据：**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "qr_code": "string",      // 二维码标识
        "qr_url": "string",       // 二维码图片URL
        "session_id": "string",   // 会话ID
        "expires_in": 300         // 过期时间(秒)
    }
}
```

### 4.2 检查登录状态
```
GET /api/auth/wechat/status/{session_id}
```

**响应数据：**
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "status": "pending|scanned|confirmed|expired|failed",
        "user_info": {
            "user_id": "number",
            "nickname": "string",
            "avatar_url": "string"
        },
        "redirect_url": "string"
    }
}
```

### 4.3 确认登录
```
POST /api/auth/wechat/confirm
```

**请求参数：**
```json
{
    "session_id": "string",
    "action": "confirm|cancel"
}
```

### 4.4 获取用户信息
```
GET /api/user/profile
```

**请求头：**
```
Authorization: Bearer {token}
```

**响应数据：**
```json
{
    "success": true,
    "message": "success",
    "data": {
        "id": "number",
        "phone": "string",
        "nickname": "string",
        "avatar_url": "string",
        "status": "number",
        "login_type": "phone|wechat|both",
        "wechat_info": {
            "wechat_user_id": "number",
            "openid": "string",
            "nickname": "string",
            "avatar_url": "string",
            "gender": "number",
            "country": "string",
            "province": "string",
            "city": "string",
            "bind_at": "string"
        },
        "created_at": "string",
        "updated_at": "string",
        "last_login_at": "string"
    }
}
```

### 4.5 绑定微信账号
```
POST /api/user/bind-wechat
```

**请求头：**
```
Authorization: Bearer {token}
```

**请求参数：**
```json
{
    "qr_code": "string",
    "session_id": "string"
}
```

**响应数据：**
```json
{
    "success": true,
    "message": "微信账号绑定成功",
    "data": {
        "user_id": "number",
        "login_type": "both",
        "wechat_info": {
            "openid": "string",
            "nickname": "string",
            "avatar_url": "string"
        }
    }
}
```

### 4.6 解绑微信账号
```
POST /api/user/unbind-wechat
```

**请求头：**
```
Authorization: Bearer {token}
```

**响应数据：**
```json
{
    "success": true,
    "message": "微信账号解绑成功",
    "data": {
        "user_id": "number",
        "login_type": "phone"
    }
}
```

## 5. 核心代码结构

### 5.1 模型层 (Models)
- `User` - 保持现有用户模型不变
- `WechatUser` - 微信用户模型，通过uid字段关联User
- `UserLoginType` - 用户登录方式模型
- `LoginSession` - 登录会话模型

### 5.2 服务层 (Services)
- `UserService` - 扩展现有用户服务，支持微信登录
- `WechatAuthService` - 微信认证服务
- `QRCodeService` - 二维码生成服务
- `SessionService` - 会话管理服务
- `KVService` - 扩展现有KV服务，支持微信会话存储

### 5.3 控制器层 (Controllers)
- `LoginController` - 扩展现有登录控制器
- `WechatAuthController` - 微信认证控制器
- `UserController` - 用户信息管理控制器

### 5.4 中间件 (Middleware)
- `AuthMiddleware` - 扩展现有认证中间件
- `RateLimitMiddleware` - 限流中间件

### 5.5 数据仓库 (Repository)
- `UserRepository` - 保持现有用户仓库不变
- `WechatUserRepository` - 微信用户仓库
- `UserLoginTypeRepository` - 用户登录方式仓库

## 6. 实现流程

### 6.1 微信扫码登录流程
1. 前端请求生成微信登录二维码
2. 后端生成唯一session_id和qr_code
3. 调用微信API获取二维码
4. 将信息存储到Redis和数据库
5. 返回二维码信息给前端
6. 用户扫描二维码
7. 微信服务器回调我们的接口
8. 更新登录状态为"scanned"
9. 通过WebSocket推送状态给前端
10. 用户确认登录
11. 获取微信用户信息
12. 检查wechat_users表中是否存在该openid
13. 如果存在且已绑定用户，直接登录
14. 如果存在但未绑定用户，创建绑定关系
15. 如果不存在，创建新的wechat_user记录
16. 更新user_login_types表的登录方式
17. 生成JWT Token（与现有登录系统兼容）
18. 返回登录成功信息

### 6.2 账号绑定流程
1. 已登录用户请求绑定微信账号
2. 生成微信登录二维码
3. 用户扫描并确认
4. 获取微信用户信息
5. 检查wechat_users表中该openid是否已被其他用户绑定
6. 如果未绑定，创建wechat_user记录并设置uid为当前用户ID
7. 如果已绑定其他用户，返回错误
8. 更新user_login_types表的login_type为"both"
9. 返回绑定成功信息

### 6.3 状态检查流程
1. 前端定期轮询登录状态
2. 后端检查Redis中的状态
3. 如果状态变化，返回最新状态
4. 支持WebSocket实时推送

### 6.4 多登录方式兼容流程
1. 用户可通过手机号验证码登录（现有功能）
2. 用户可通过微信扫码登录（新增功能）
3. 用户可绑定两种登录方式
4. 统一使用JWT Token进行身份验证
5. 支持登录方式切换和账号解绑

## 7. 安全考虑

### 7.1 数据安全
- 敏感信息加密存储
- 使用HTTPS传输
- 定期清理过期数据

### 7.2 接口安全
- 接口限流防刷
- 参数验证和过滤
- CSRF防护

### 7.3 会话安全
- 会话ID随机生成
- 设置合理的过期时间
- 支持会话撤销

## 8. 配置要求

### 8.1 扩展现有配置文件
在 `src/config/config.yaml` 中添加微信配置：

```yaml
# 现有配置
jwt_secret: "${JWT_SECRET}"

# 新增微信配置
wechat:
  app_id: "${WECHAT_APP_ID}"
  app_secret: "${WECHAT_APP_SECRET}"
  redirect_uri: "${WECHAT_REDIRECT_URI}"
  scope: "snsapi_login"
  qr_expires_in: 300  # 二维码过期时间(秒)
  session_expires_in: 600  # 会话过期时间(秒)

# Redis配置（扩展现有配置）
redis:
  host: "${REDIS_HOST}"
  port: "${REDIS_PORT}"
  db: 0
  password: "${REDIS_PASSWORD}"
  key_prefix: "wechat_login:"
```

### 8.2 环境变量配置
在 `.env.api` 文件中添加：

```bash
# 现有配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# 新增微信配置
WECHAT_APP_ID=your_wechat_app_id
WECHAT_APP_SECRET=your_wechat_app_secret
WECHAT_REDIRECT_URI=https://your-domain.com/api/auth/wechat/callback

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

### 8.3 配置结构体扩展
在 `src/config/config.go` 中添加：

```go
type WechatConfig struct {
    AppID          string `mapstructure:"app_id"`
    AppSecret      string `mapstructure:"app_secret"`
    RedirectURI    string `mapstructure:"redirect_uri"`
    Scope          string `mapstructure:"scope"`
    QRExpiresIn    int    `mapstructure:"qr_expires_in"`
    SessionExpiresIn int  `mapstructure:"session_expires_in"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    DB       int    `mapstructure:"db"`
    Password string `mapstructure:"password"`
    KeyPrefix string `mapstructure:"key_prefix"`
}

type AppConfig struct {
    JWTSecret string       `mapstructure:"jwt_secret"`
    Wechat    WechatConfig `mapstructure:"wechat"`
    Redis     RedisConfig  `mapstructure:"redis"`
}
```

## 9. 部署说明

### 9.1 环境要求
- Go 1.24+
- MySQL 8.0+
- Redis 6.0+
- 微信开放平台账号

### 9.2 部署步骤
1. **配置微信开放平台应用**
   - 注册微信开放平台账号
   - 创建网站应用
   - 配置授权回调域名
   - 获取AppID和AppSecret

2. **更新配置文件**
   - 修改 `src/config/config.yaml`
   - 设置 `.env.api` 环境变量
   - 配置Redis连接信息

3. **数据库迁移**
   ```sql
   -- 创建微信用户表
   CREATE TABLE wechat_users (
       id BIGINT PRIMARY KEY AUTO_INCREMENT,
       uid BIGINT COMMENT '关联用户ID',
       openid VARCHAR(64) UNIQUE NOT NULL COMMENT '微信OpenID',
       -- ... (其他字段)
   );
   
   -- 创建用户登录方式表
   CREATE TABLE user_login_types (
       id BIGINT PRIMARY KEY AUTO_INCREMENT,
       user_id BIGINT NOT NULL COMMENT '用户ID',
       login_type ENUM('phone', 'wechat', 'both') DEFAULT 'phone' COMMENT '登录方式',
       wechat_user_id BIGINT COMMENT '关联的微信用户ID',
       -- ... (其他字段)
   );
   
   -- 为现有用户初始化登录方式
   INSERT INTO user_login_types (user_id, login_type) 
   SELECT id, 'phone' FROM users;
   ```

4. **启动应用服务**
   ```bash
   go run src/main.go
   ```

### 9.3 监控指标
- 二维码生成成功率
- 微信登录成功率
- 手机号登录成功率（现有功能）
- 账号绑定成功率
- 接口响应时间
- 错误率统计

### 9.4 兼容性说明
- ✅ 保留现有手机号验证码登录功能
- ✅ 新增微信扫码登录功能
- ✅ 支持两种登录方式并存
- ✅ 统一JWT Token认证机制
- ✅ 向后兼容现有API接口

## 10. 测试计划

### 10.1 单元测试
- 模型层测试（扩展现有User模型）
- 服务层测试（WechatAuthService、QRCodeService）
- 工具函数测试（微信API调用、Token生成）

### 10.2 集成测试
- API接口测试（微信登录、账号绑定）
- 数据库操作测试（wechat_users表、user_login_types表）
- Redis缓存测试（会话存储、状态管理）
- 现有登录功能兼容性测试
- 表关联关系测试

### 10.3 端到端测试
- 微信扫码登录完整流程测试
- 账号绑定和解绑流程测试
- 多登录方式切换测试
- 异常情况处理测试（网络异常、微信API异常）
- 性能压力测试

### 10.4 兼容性测试
- 现有手机号登录功能验证
- JWT Token兼容性验证
- 用户数据迁移验证
- API接口向后兼容验证

## 11. 后续优化

### 11.1 功能扩展
- 支持微信小程序登录
- 支持微信公众号登录
- 支持QQ、支付宝等第三方登录
- 多平台账号绑定和统一管理
- 登录方式偏好设置

### 11.2 性能优化
- Redis缓存优化（用户信息、会话状态）
- 数据库查询优化（索引优化、分页查询）
- 并发处理优化（微信API调用限流）
- 二维码生成性能优化

### 11.3 用户体验
- 登录状态持久化
- 自动登录功能
- 登录历史记录
- 登录安全提醒
- 异常登录检测

### 11.4 安全性增强
- 微信账号绑定验证
- 登录设备管理
- 异常登录告警
- 账号安全等级评估

## 12. 与现有系统集成

### 12.1 现有功能保留
- ✅ 手机号验证码登录（`POST /api/login`）
- ✅ 发送验证码（`POST /api/send-code`）
- ✅ JWT Token认证机制
- ✅ 用户信息获取（`GET /api/user/profile`）

### 12.2 新增功能
- 🔄 微信扫码登录（`POST /api/auth/wechat/qr-code`）
- 🔄 微信登录状态检查（`GET /api/auth/wechat/status/{session_id}`）
- 🔄 微信账号绑定（`POST /api/user/bind-wechat`）
- 🔄 微信账号解绑（`POST /api/user/unbind-wechat`）

### 12.3 数据迁移策略
1. **零影响迁移**：现有users表完全不变，保持原有功能
2. **独立表设计**：微信用户数据存储在独立的wechat_users表中
3. **关联关系管理**：通过uid字段和user_login_types表管理关联关系
4. **可选绑定**：用户可选择是否绑定微信账号
5. **登录方式升级**：支持从单一登录方式升级为多登录方式

### 12.4 API兼容性
- 保持现有API接口不变
- 新增微信相关接口
- 统一响应格式
- 向后兼容现有客户端 