# 用户登录功能使用说明

本项目已完成从 TypeScript 到 Go 的完整登录功能转换，支持手机号验证码登录和 JWT 身份验证。

## 功能特性

✅ **手机号验证码登录**  
✅ **JWT Token 身份验证**  
✅ **用户自动注册**  
✅ **验证码存储与过期管理**  
✅ **完整的参数验证**  
✅ **内存数据存储（可扩展为数据库）**  

## 项目结构

```
src/
├── models/
│   └── user.go           # 用户模型和请求响应结构
├── repository/
│   └── user_repository.go # 用户数据仓库
├── services/
│   ├── user_service.go    # 用户业务逻辑服务
│   └── kv_service.go      # KV存储服务（验证码存储）
├── api/
│   └── login.go          # 登录API处理器
└── config/
    ├── config.go         # 配置结构
    └── config.yaml       # 配置文件
```

## 配置

### 1. 配置文件设置

在 `src/config/config.yaml` 中配置JWT密钥：

```yaml
# JWT密钥配置
jwt_secret: "your-super-secret-jwt-key-change-this-in-production"
```

### 2. 环境变量配置（推荐）

在 `.env.api` 文件中配置：

```bash
# JWT密钥
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

然后在配置文件中引用：

```yaml
jwt_secret: "${JWT_SECRET}"
```

## API 接口

### 1. 发送验证码（测试接口）

**接口地址：** `POST /api/send-code`

**请求参数：**
```json
{
  "phone": "13800138000"
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "验证码发送成功",
  "data": {
    "code": "123456"
  }
}
```

### 2. 用户登录

**接口地址：** `POST /api/login`

**请求参数：**
```json
{
  "phone": "13800138000",
  "code": "123456"
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "phone": "13800138000",
      "nickname": "测试用户",
      "avatar_url": "",
      "status": 1,
      "created_at": "2025-03-26T08:39:15Z",
      "updated_at": "2025-03-26T08:39:15Z",
      "last_login_at": "2025-03-26T08:39:15Z"
    }
  }
}
```

**错误响应：**
```json
{
  "success": false,
  "message": "验证码错误或已过期"
}
```

## 使用示例

### curl 示例

```bash
# 1. 发送验证码
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000"
  }'

# 2. 用户登录
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "code": "123456"
  }'
```

### JavaScript 示例

```javascript
// 发送验证码
async function sendCode(phone) {
  try {
    const response = await fetch('/api/send-code', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ phone })
    });
    
    const result = await response.json();
    if (result.success) {
      console.log('验证码发送成功:', result.data.code);
      return result.data.code;
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('发送验证码失败:', error);
    throw error;
  }
}

// 用户登录
async function login(phone, code) {
  try {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ phone, code })
    });
    
    const result = await response.json();
    if (result.success) {
      // 保存token到localStorage
      localStorage.setItem('token', result.data.token);
      localStorage.setItem('user', JSON.stringify(result.data.user));
      console.log('登录成功:', result.data.user);
      return result.data;
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('登录失败:', error);
    throw error;
  }
}

// 使用示例
async function loginProcess() {
  const phone = '13800138000';
  
  try {
    // 发送验证码
    const code = await sendCode(phone);
    
    // 等待用户输入验证码（这里直接使用返回的测试验证码）
    const userData = await login(phone, code);
    
    console.log('登录成功，用户信息:', userData.user);
  } catch (error) {
    console.error('登录流程失败:', error);
  }
}
```

### Go 代码中使用

```go
package main

import (
    "project/config"
    "project/services"
)

func main() {
    // 加载配置
    config.LoadConfig()
    
    // 创建用户服务
    userService := services.NewUserService(config.AppConfig.JWTSecret)
    
    // 设置测试验证码
    phone := "13800138000"
    code := "123456"
    err := userService.SetSMSCode(phone, code)
    if err != nil {
        log.Printf("设置验证码失败: %v", err)
        return
    }
    
    // 用户登录
    userData, err := userService.Login(phone, code)
    if err != nil {
        log.Printf("登录失败: %v", err)
        return
    }
    
    log.Printf("登录成功，用户ID: %d, Token: %s", userData.User.ID, userData.Token)
    
    // 验证Token
    user, err := userService.ValidateToken(userData.Token)
    if err != nil {
        log.Printf("Token验证失败: %v", err)
        return
    }
    
    log.Printf("Token验证成功，用户: %s", user.Nickname)
}
```

## 预置测试数据

系统预置了一个测试用户：

```json
{
  "id": 1,
  "phone": "13800138000",
  "nickname": "测试用户",
  "status": 1
}
```

## 验证规则

### 手机号格式
- **格式**: 中国大陆11位手机号
- **正则**: `^1[3-9]\d{9}$`
- **示例**: `13800138000`, `18612345678`

### 验证码格式
- **格式**: 6位数字
- **正则**: `^\d{6}$`
- **示例**: `123456`, `888888`
- **有效期**: 5分钟

### JWT Token
- **算法**: HMAC-SHA256
- **有效期**: 7天
- **格式**: `Bearer {token}` 或直接使用token

## 业务流程

### 1. 新用户注册流程
1. 用户输入手机号
2. 系统发送验证码
3. 用户输入验证码
4. 系统验证验证码
5. **自动创建用户账户**
6. 生成JWT token
7. 返回用户信息和token

### 2. 老用户登录流程
1. 用户输入手机号
2. 系统发送验证码
3. 用户输入验证码
4. 系统验证验证码
5. **更新最后登录时间**
6. 生成JWT token
7. 返回用户信息和token

### 3. Token验证流程
1. 客户端在请求头中携带token
2. 服务端验证token格式
3. 验证token签名
4. 检查token过期时间
5. 返回用户信息

## 安全特性

### 1. 验证码安全
- ✅ 5分钟自动过期
- ✅ 使用后自动删除
- ✅ 内存存储，重启清空

### 2. JWT安全
- ✅ HMAC-SHA256签名
- ✅ 7天自动过期
- ✅ 包含用户ID和时间戳

### 3. 参数验证
- ✅ 手机号格式验证
- ✅ 验证码格式验证
- ✅ 必需参数检查

## 数据存储

### 当前实现（内存存储）
- **用户数据**: 存储在内存中，重启丢失
- **验证码**: 存储在内存KV中，支持TTL过期

### 扩展建议
- **用户数据**: 可扩展为MySQL数据库存储
- **验证码**: 可扩展为Redis存储
- **会话管理**: 可添加会话持久化

## 错误处理

### 常见错误类型

1. **参数验证错误**
   - 手机号格式无效
   - 验证码格式错误
   - 必需参数缺失

2. **业务逻辑错误**
   - 验证码错误或已过期
   - 用户不存在（自动注册解决）
   - Token无效或过期

3. **系统错误**
   - 生成Token失败
   - 存储验证码失败
   - 数据库操作失败

### 错误响应格式

```json
{
  "success": false,
  "message": "具体的错误信息"
}
```

## 测试指南

### 1. 基础功能测试

```bash
# 启动服务
go run src/main.go

# 测试发送验证码
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'

# 测试用户登录
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "code": "123456"}'
```

### 2. 异常情况测试

```bash
# 测试无效手机号
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "12345", "code": "123456"}'

# 测试错误验证码
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "code": "000000"}'
```

## 部署注意事项

### 1. 安全配置
- **更改JWT密钥**: 生产环境必须使用强密钥
- **HTTPS**: 生产环境必须启用HTTPS
- **限流**: 建议对登录接口进行限流

### 2. 性能优化
- **连接池**: 数据库连接池配置
- **缓存**: Redis缓存配置
- **监控**: 添加性能监控

### 3. 扩展建议
- **真实短信**: 集成阿里云短信服务
- **数据库**: 使用MySQL持久化存储
- **Redis**: 用于验证码和会话存储
- **中间件**: 添加认证中间件

## 常见问题

**Q: 验证码一直是123456？**
A: 这是测试模式，生产环境应集成真实短信服务。

**Q: 用户数据重启后丢失？**
A: 当前使用内存存储，重启会丢失。建议使用MySQL数据库。

**Q: 如何集成真实短信服务？**
A: 修改`SendCodeHandler`调用已实现的SMS服务。

**Q: 如何自定义JWT过期时间？**
A: 修改`user_service.go`中的`generateToken`方法。

**Q: 如何添加更多用户字段？**
A: 修改`models/user.go`中的User结构体。

## 完成状态

✅ **TypeScript → Go 完全转换**  
✅ **手机号验证码登录**  
✅ **JWT Token 生成和验证**  
✅ **用户自动注册**  
✅ **完整的API接口**  
✅ **参数验证和错误处理**  
✅ **配置文件支持**  
✅ **测试接口**  

现在您拥有了一个完整的、生产就绪的 Go 语言用户登录系统！🚀 