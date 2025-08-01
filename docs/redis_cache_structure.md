# Redis 缓存结构设计

## 微信登录会话缓存

### 数据结构

#### Key 格式
```
wechat_login:session:{session_id}
```

#### Value 格式 (JSON)
```json
{
    "session_id": "string",         // 会话ID (UUID)
    "qr_scene": "string",           // 二维码场景值
    "status": "string",             // pending|success|expired
    "user_id": "number",            // 登录成功后的用户ID
    "openid": "string",             // 微信openid
    "created_at": "timestamp",      // 创建时间戳
    "expires_at": "timestamp"       // 过期时间戳
}
```

#### 过期时间
- **默认过期时间**: 600秒 (10分钟)
- **自动清理**: Redis TTL 机制自动清理过期数据

### 使用示例

#### 1. 创建登录会话
```bash
# 生成二维码时创建会话
SET wechat_login:session:550e8400-e29b-41d4-a716-446655440000 '{"session_id":"550e8400-e29b-41d4-a716-446655440000","qr_scene":"login_2024032601","status":"pending","user_id":0,"openid":"","created_at":1711123200,"expires_at":1711123800}' EX 600
```

#### 2. 更新会话状态
```bash
# 用户扫码关注后更新
SET wechat_login:session:550e8400-e29b-41d4-a716-446655440000 '{"session_id":"550e8400-e29b-41d4-a716-446655440000","qr_scene":"login_2024032601","status":"success","user_id":123,"openid":"oGx123456789","created_at":1711123200,"expires_at":1711123800}' EX 600
```

#### 3. 查询会话状态
```bash
# WebSocket 和 API 查询会话
GET wechat_login:session:550e8400-e29b-41d4-a716-446655440000
```

#### 4. 清理过期会话
```bash
# Redis 自动清理，也可手动清理
DEL wechat_login:session:550e8400-e29b-41d4-a716-446655440000
```

### 状态流转

```
pending  →  success  →  (自动过期清理)
    ↓
expired
```

1. **pending**: 二维码生成后的初始状态，等待用户扫码
2. **success**: 用户扫码关注成功，登录完成
3. **expired**: 超过过期时间，会话失效

### 性能考虑

1. **索引**: 使用 session_id 作为 Key，O(1) 查询复杂度
2. **内存优化**: 设置合理的过期时间，避免内存泄露
3. **并发处理**: Redis 原子操作保证并发安全

### 监控指标

1. **会话创建数**: 每日二维码生成次数
2. **会话成功率**: 成功登录/总会话数
3. **会话过期率**: 过期会话/总会话数
4. **平均响应时间**: 从扫码到登录成功的时间

### 清理策略

1. **自动清理**: Redis TTL 机制，过期自动删除
2. **定期清理**: 可选的定时任务清理异常数据
3. **手动清理**: 运维工具支持手动清理指定会话