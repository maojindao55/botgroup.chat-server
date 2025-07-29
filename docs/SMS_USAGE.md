# SMS 短信服务使用说明

本项目已集成阿里云短信服务，支持发送验证码和自定义模板短信。

## 配置

### 1. 配置文件设置

在 `src/config/config.yaml` 中添加短信服务配置：

```yaml
# 阿里云短信服务配置
sms:
  access_key_id: "YOUR_ALIYUN_ACCESS_KEY_ID"
  access_key_secret: "YOUR_ALIYUN_ACCESS_KEY_SECRET"
  sign_name: "您的签名"
  template_code: "SMS_123456789"
```

### 2. 环境变量配置（推荐）

为了安全性，建议在 `.env.api` 文件中配置敏感信息：

```bash
# 阿里云短信配置
ALIYUN_SMS_ACCESS_KEY_ID=your_access_key_id
ALIYUN_SMS_ACCESS_KEY_SECRET=your_access_key_secret
ALIYUN_SMS_SIGN_NAME=您的签名
ALIYUN_SMS_TEMPLATE_CODE=SMS_123456789
```

然后在配置文件中使用环境变量：

```yaml
sms:
  access_key_id: "${ALIYUN_SMS_ACCESS_KEY_ID}"
  access_key_secret: "${ALIYUN_SMS_ACCESS_KEY_SECRET}"
  sign_name: "${ALIYUN_SMS_SIGN_NAME}"
  template_code: "${ALIYUN_SMS_TEMPLATE_CODE}"
```

## API 接口

### 1. 发送验证码短信

**接口地址：** `POST /api/sms/send`

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
  "message": "短信发送成功"
}
```

**错误响应：**
```json
{
  "success": false,
  "message": "短信发送失败: 错误详情"
}
```

### 2. 发送自定义模板短信

**接口地址：** `POST /api/sms/send-template`

**请求参数：**
```json
{
  "phone": "13800138000",
  "template_code": "SMS_123456789",
  "template_param": {
    "code": "123456",
    "product": "测试产品"
  }
}
```

**响应示例：**
```json
{
  "success": true,
  "message": "短信发送成功"
}
```

## 使用示例

### curl 示例

```bash
# 发送验证码短信
curl -X POST http://localhost:8080/api/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "code": "123456"
  }'

# 发送自定义模板短信
curl -X POST http://localhost:8080/api/sms/send-template \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "template_code": "SMS_123456789",
    "template_param": {
      "code": "123456",
      "name": "张三"
    }
  }'
```

### JavaScript 示例

```javascript
// 发送验证码短信
async function sendSMS(phone, code) {
  try {
    const response = await fetch('/api/sms/send', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        phone: phone,
        code: code
      })
    });
    
    const result = await response.json();
    if (result.success) {
      console.log('短信发送成功');
    } else {
      console.error('短信发送失败:', result.message);
    }
  } catch (error) {
    console.error('请求失败:', error);
  }
}

// 使用示例
sendSMS('13800138000', '123456');
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
    
    // 创建SMS服务
    smsService := services.NewSMSService(config.AppConfig.SMS)
    
    // 发送验证码短信
    err := smsService.SendSMS("13800138000", "123456")
    if err != nil {
        log.Printf("短信发送失败: %v", err)
        return
    }
    
    log.Println("短信发送成功")
}
```

## 参数说明

### 手机号格式
- 支持中国大陆手机号格式：`1[3-9]xxxxxxxxx`
- 11位数字，以1开头，第二位为3-9

### 验证码格式
- 支持4-8位数字
- 示例：`1234`、`123456`、`12345678`

## 错误处理

常见错误类型：

1. **参数验证错误**
   - 手机号格式无效
   - 验证码格式无效
   - 必需参数缺失

2. **阿里云API错误**
   - AccessKey 配置错误
   - 短信模板不存在
   - 短信签名不正确
   - 发送频率限制

3. **网络错误**
   - 网络连接失败
   - 超时错误

## 注意事项

1. **安全性**
   - 不要在客户端代码中暴露 AccessKey
   - 使用环境变量存储敏感配置
   - 实施发送频率限制

2. **费用控制**
   - 阿里云短信按条收费
   - 建议实施发送频率限制
   - 监控短信发送量

3. **合规性**
   - 确保短信内容符合法律法规
   - 获得用户明确同意
   - 提供退订机制

## 阿里云短信服务配置

### 1. 开通服务
1. 登录阿里云控制台
2. 开通短信服务
3. 创建 AccessKey

### 2. 配置签名和模板
1. 在短信服务控制台添加短信签名
2. 创建短信模板
3. 等待审核通过

### 3. 获取配置信息
- **AccessKeyId**: 阿里云访问密钥ID
- **AccessKeySecret**: 阿里云访问密钥Secret
- **SignName**: 已审核通过的短信签名
- **TemplateCode**: 已审核通过的短信模板代码

## 常见问题

**Q: 短信发送失败，提示签名不存在？**
A: 检查签名是否已通过审核，确保配置的签名名称完全一致。

**Q: 模板参数错误？**
A: 检查模板参数是否与阿里云控制台中的模板变量名称一致。

**Q: 发送频率限制？**
A: 阿里云对短信发送有频率限制，建议在应用层实施限流。

**Q: 如何测试短信功能？**
A: 可以使用阿里云提供的测试环境或配置测试模板进行验证。 