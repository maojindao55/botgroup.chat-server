# Docker 容器配置说明

本项目现在包含以下容器服务：

## 服务列表

- **nginx**: 反向代理服务器 (端口 8082)
- **golang-app**: Go 应用程序 (端口 8080)
- **rag-app**: RAG 应用程序 (端口 8070)
- **redis**: Redis 缓存数据库 (端口 6379)
- **mysql**: MySQL 关系型数据库 (端口 3306)

## 环境变量配置

请创建 `.env.api` 文件来配置数据库连接信息：

```bash
# MySQL 配置
MYSQL_ROOT_PASSWORD=root123
MYSQL_DATABASE=botgroup_chat
MYSQL_USER=botgroup
MYSQL_PASSWORD=botgroup123

# Redis 配置
REDIS_PASSWORD=redis123

# 应用配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=botgroup
DB_PASSWORD=botgroup123
DB_NAME=botgroup_chat

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis123

# JWT密钥配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

## 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

## 数据库连接信息

### MySQL
- **主机**: localhost 或 mysql (容器内部)
- **端口**: 3306
- **数据库**: botgroup_chat
- **用户**: botgroup
- **密码**: botgroup123
- **Root 密码**: root123

#### 当前数据结构
- **users 表**: 存储用户信息
  - `id`: 用户ID (自增主键)
  - `phone`: 手机号 (11位，唯一)
  - `nickname`: 昵称
  - `avatar_url`: 头像URL
  - `status`: 状态 (默认为1)
  - `created_at`: 创建时间
  - `updated_at`: 更新时间
  - `last_login_at`: 最后登录时间

#### 预置测试数据
- 用户1: 13800138000 (测试用户)

### Redis
- **主机**: localhost 或 redis (容器内部)
- **端口**: 6379
- **密码**: redis123

## 数据持久化

- **MySQL 数据**: 存储在 `./mysql/data/` 目录中
- **Redis 数据**: 存储在 `./redis/data/` 目录中
- **验证码存储**: 使用Redis存储，支持TTL过期（fallback到内存存储）

## 数据库初始化

MySQL 容器会自动执行 `mysql/init/` 目录下的 SQL 脚本。当前包含：

- `01-init-database.sql`: 创建数据库和基础表结构
  - 创建 `botgroup_chat` 数据库
  - `users`: 用户表，包含测试数据

## 健康检查

所有服务都配置了健康检查：
- **MySQL**: 通过 `mysqladmin ping` 检查
- **Redis**: 通过 `redis-cli ping` 检查

## 服务依赖

- `golang-app` 和 `rag-app` 都依赖于 `redis` 和 `mysql`
- `nginx` 依赖于 `golang-app` 和 `rag-app`

## 管理命令

```bash
# 停止所有服务
docker-compose down

# 重新构建并启动
docker-compose up --build -d

# 清理本地数据（注意：会丢失数据）
rm -rf mysql/data redis/data

# 连接到 MySQL
docker-compose exec mysql mysql -u botgroup -p botgroup_chat

# 查看所有数据库
docker-compose exec mysql mysql -u root -p -e "SHOW DATABASES;"

# 查看用户表数据
docker-compose exec mysql mysql -u botgroup -p botgroup_chat -e "SELECT * FROM users;"

# 连接到 Redis
docker-compose exec redis redis-cli -a redis123
```

## 开发建议

1. 在应用程序中使用容器名称作为主机名连接数据库
2. 使用环境变量配置数据库连接参数
3. 在生产环境中修改默认密码
4. 定期备份 `mysql/data` 和 `redis/data` 目录中的数据
5. 确保本地 `mysql/data` 和 `redis/data` 目录有正确的权限
6. 可以通过本地文件系统直接访问和备份数据库文件 