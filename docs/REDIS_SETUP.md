# Redis 配置使用说明

项目已将 KV 存储从内存实现迁移到 Redis，支持持久化存储和集群部署。

## 功能特性

✅ **Redis 持久化存储** - 验证码数据持久化  
✅ **TTL 过期管理** - 自动过期机制  
✅ **连接失败回退** - 自动回退到内存存储  
✅ **容器化部署** - Docker 容器支持  
✅ **密码认证** - 支持 Redis 密码保护  

## 配置说明

### 1. 配置文件设置

在 `src/config/config.yaml` 中配置 Redis 连接：

```yaml
# Redis配置
redis:
  host: "redis"  # Docker容器中使用服务名，本地开发可改为localhost
  port: "6379"
  password: "${REDIS_PASSWORD:-redis123}"
  db: 0
```

### 2. 环境变量配置

在 `.env.api` 文件中设置：

```bash
# Redis 配置
REDIS_PASSWORD=redis123
REDIS_HOST=redis
REDIS_PORT=6379
```

## 使用方式

### 1. Docker 容器运行（推荐）

```bash
# 启动所有服务（包含Redis）
docker-compose up -d

# 查看Redis服务状态
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis
```

### 2. 本地 Redis 服务

如果在本地运行Redis服务：

```yaml
# 修改配置文件中的host
redis:
  host: "localhost"  # 本地Redis服务
  port: "6379"
  password: "your-redis-password"
  db: 0
```

## Redis 连接测试

### 1. 通过容器连接

```bash
# 连接到Redis容器
docker-compose exec redis redis-cli -a redis123

# 测试连接
127.0.0.1:6379> ping
PONG

# 查看所有键
127.0.0.1:6379> keys *

# 查看验证码相关的键
127.0.0.1:6379> keys sms:*
```

### 2. 查看验证码存储

```bash
# 设置一个测试验证码（通过API）
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'

# 在Redis中查看
docker-compose exec redis redis-cli -a redis123
127.0.0.1:6379> get "sms:13800138000"
"123456"

# 查看TTL（剩余过期时间）
127.0.0.1:6379> ttl "sms:13800138000"
(integer) 298  # 约5分钟
```

## 数据结构说明

### 验证码存储格式

- **Key 格式**: `sms:{phone}`
- **Value**: 验证码字符串
- **TTL**: 5分钟（300秒）

示例：
```
Key: "sms:13800138000"
Value: "123456"
TTL: 300秒
```

## 容错机制

### 1. 自动回退

当Redis连接失败时，系统会自动回退到内存存储：

```go
// Redis连接失败时的日志
fmt.Println("Redis连接失败，使用内存KV存储")

// Redis连接成功时的日志
fmt.Println("Redis连接成功，使用Redis KV存储")
```

### 2. 连接重试

应用启动时会尝试连接Redis：
- ✅ **连接成功**: 使用Redis存储
- ❌ **连接失败**: 自动使用内存存储

## 性能优化

### 1. 连接池配置

Redis客户端默认使用连接池，可以在需要时自定义：

```go
// 自定义Redis客户端配置
rdb := redis.NewClient(&redis.Options{
    Addr:         "redis:6379",
    Password:     "redis123",
    DB:           0,
    PoolSize:     10,           // 连接池大小
    MinIdleConns: 5,            // 最小空闲连接
    MaxRetries:   3,            // 最大重试次数
})
```

### 2. 监控和日志

```bash
# 查看Redis内存使用
docker-compose exec redis redis-cli -a redis123 info memory

# 查看连接数
docker-compose exec redis redis-cli -a redis123 info clients

# 查看命令统计
docker-compose exec redis redis-cli -a redis123 info commandstats
```

## 数据备份

### 1. Redis 数据备份

```bash
# 手动触发保存
docker-compose exec redis redis-cli -a redis123 bgsave

# 查看备份文件
docker-compose exec redis ls -la /data/

# 复制备份文件到本地
docker cp $(docker-compose ps -q redis):/data/dump.rdb ./redis_backup.rdb
```

### 2. 数据恢复

```bash
# 停止Redis服务
docker-compose stop redis

# 复制备份文件到数据目录
cp redis_backup.rdb ./redis/data/dump.rdb

# 重启Redis服务
docker-compose start redis
```

## 安全配置

### 1. 密码保护

```yaml
# 强密码配置
redis:
  password: "${REDIS_PASSWORD:-your-super-strong-redis-password}"
```

### 2. 网络隔离

Redis仅在Docker网络内部可访问，不对外暴露端口（生产环境推荐）：

```yaml
# docker-compose.yaml 中不暴露端口
redis:
  # ports:
  #   - "6379:6379"  # 注释掉端口映射
```

### 3. 访问控制

```bash
# 进入Redis容器设置ACL
docker-compose exec redis redis-cli -a redis123

# 创建受限用户
127.0.0.1:6379> ACL SETUSER app_user on >app_password ~sms:* +get +set +del +ttl
```

## 故障排除

### 1. 连接问题

```bash
# 检查Redis服务状态
docker-compose ps redis

# 查看Redis日志
docker-compose logs redis

# 测试网络连接
docker-compose exec golang-app ping redis
```

### 2. 认证问题

```bash
# 检查密码配置
docker-compose exec redis redis-cli -a wrong_password
(error) NOAUTH Authentication required.

# 正确的认证
docker-compose exec redis redis-cli -a redis123
127.0.0.1:6379> ping
PONG
```

### 3. 内存问题

```bash
# 查看内存使用
docker-compose exec redis redis-cli -a redis123 info memory

# 清理过期键
docker-compose exec redis redis-cli -a redis123 flushall
```

## 开发调试

### 1. 本地开发配置

```yaml
# 本地开发时的配置
redis:
  host: "localhost"
  port: "6379"
  password: ""  # 本地Redis可以不设密码
  db: 0
```

### 2. 调试工具

```bash
# 实时监控Redis命令
docker-compose exec redis redis-cli -a redis123 monitor

# 查看慢查询
docker-compose exec redis redis-cli -a redis123 slowlog get 10
```

## 集成测试

### 1. 验证码存储测试

```bash
# 1. 发送验证码
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'

# 2. 在Redis中验证
docker-compose exec redis redis-cli -a redis123 get "sms:13800138000"

# 3. 用户登录
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "code": "123456"}'

# 4. 验证验证码已删除
docker-compose exec redis redis-cli -a redis123 get "sms:13800138000"
(nil)
```

### 2. TTL 测试

```bash
# 设置验证码
curl -X POST http://localhost:8080/api/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'

# 检查TTL
docker-compose exec redis redis-cli -a redis123 ttl "sms:13800138000"
# 应该显示约300秒（5分钟）

# 等待5分钟后再检查
docker-compose exec redis redis-cli -a redis123 get "sms:13800138000"
# 应该返回 (nil)，表示已过期
```

## 监控和告警

### 1. 健康检查

Redis容器已配置健康检查：

```yaml
healthcheck:
  test: ["CMD", "redis-cli", "ping"]
  interval: 30s
  timeout: 10s
  retries: 5
```

### 2. 监控指标

```bash
# 连接数监控
docker-compose exec redis redis-cli -a redis123 info clients | grep connected_clients

# 内存使用监控
docker-compose exec redis redis-cli -a redis123 info memory | grep used_memory_human

# 命令执行监控
docker-compose exec redis redis-cli -a redis123 info commandstats
```

## 生产环境建议

### 1. 高可用配置

- **Redis Sentinel**: 主从复制和自动故障转移
- **Redis Cluster**: 分布式集群部署
- **备份策略**: 定期数据备份

### 2. 性能调优

- **内存优化**: 设置合适的maxmemory策略
- **持久化配置**: AOF + RDB混合持久化
- **网络优化**: 减少网络延迟

### 3. 安全加固

- **网络隔离**: VPC内部访问
- **认证授权**: 强密码 + ACL
- **数据加密**: TLS传输加密

## 完成状态

✅ **KV存储Redis化** - 从内存迁移到Redis  
✅ **自动回退机制** - Redis故障时使用内存存储  
✅ **容器化部署** - Docker Compose支持  
✅ **TTL过期管理** - 验证码自动过期  
✅ **配置灵活性** - 支持环境变量配置  
✅ **监控和调试** - 完整的故障排除指南  

现在您的验证码存储已经使用Redis，具备了生产环境的可靠性和扩展性！🚀 