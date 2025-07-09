# 开发环境配置指南

本文档介绍如何在Docker容器中使用Air工具实现Go应用的热重载，提升开发效率。

## 🚀 功能特性

✅ **Air热重载** - 文件变化时自动重新编译和重启  
✅ **Docker容器化** - 完整的开发环境容器化  
✅ **源码挂载** - 本地代码实时同步到容器  
✅ **依赖自动管理** - go.mod变化时自动更新依赖  
✅ **实时日志** - 查看编译和运行日志  
✅ **快速启动** - 一键启动完整开发环境  

## 📁 开发环境文件结构

```
botgroup.chat-server/
├── docker-compose.dev.yaml      # 开发环境Docker配置
├── Dockerfile.golang.dev        # 开发版Go Dockerfile
├── .air.toml                    # Air热重载配置
├── devrun.sh                    # 开发启动脚本
└── tmp/                         # Air编译临时目录（自动生成）
```

## 🛠 环境要求

- **Docker** >= 20.0
- **Docker Compose** >= 2.0
- **Git** （用于克隆代码）

## 🏁 快速开始

### 1. 配置API密钥

```bash
# 复制环境变量模板文件
cp .env.api.example .env.api

# 编辑配置文件，添加你的API密钥
vim .env.api
```

### 2. 启动开发环境

```bash
# 使用开发脚本启动（推荐）
./devrun.sh

# 或手动启动
docker-compose -f docker-compose.dev.yaml up -d
```

### 3. 验证热重载

修改任意Go源文件（如 `src/main.go`），Air会自动：
1. 检测文件变化
2. 重新编译应用
3. 重启服务
4. 显示编译日志

## 📝 Air配置说明

### .air.toml 配置文件

```toml
[build]
  # 编译命令
  cmd = "go build -o ./tmp/main ./main.go"
  
  # 输出文件路径
  bin = "./tmp/main"
  
  # 监控的文件扩展名
  include_ext = ["go", "tpl", "tmpl", "html", "yaml", "yml"]
  
  # 排除的目录
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules", "static", "doc", "mysql", "redis", "nginx"]
  
  # 延迟时间（毫秒）
  delay = 1000
```

### 热重载监控范围

- ✅ **监控文件**: `.go`, `.yaml`, `.yml`, `.html`, `.tpl`, `.tmpl`
- ❌ **忽略目录**: `tmp/`, `vendor/`, `static/`, `doc/`, `mysql/`, `redis/`, `nginx/`
- ❌ **忽略文件**: `*_test.go`

## 🐳 Docker开发配置

### 开发版Dockerfile特性

```dockerfile
# 安装Air工具
RUN go install github.com/cosmtrek/air@latest

# 使用Air启动
CMD ["air", "-c", ".air.toml"]
```

### 容器卷挂载

```yaml
volumes:
  # 源代码实时同步
  - ./src:/app
  - ./go.mod:/app/go.mod
  - ./go.sum:/app/go.sum
  - ./.air.toml:/app/.air.toml
  # 编译临时目录
  - air-tmp:/app/tmp
```

## 🔧 开发命令

### 基本操作

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yaml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yaml ps

# 停止开发环境
docker-compose -f docker-compose.dev.yaml down

# 重新构建并启动
docker-compose -f docker-compose.dev.yaml up -d --build
```

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f docker-compose.dev.yaml logs -f

# 查看Go应用日志（热重载日志）
docker-compose -f docker-compose.dev.yaml logs -f golang-app-dev

# 查看最近100行日志
docker-compose -f docker-compose.dev.yaml logs --tail=100 golang-app-dev
```

### 容器操作

```bash
# 进入Go应用容器
docker-compose -f docker-compose.dev.yaml exec golang-app-dev sh

# 手动重启Go服务
docker-compose -f docker-compose.dev.yaml restart golang-app-dev

# 查看容器内文件
docker-compose -f docker-compose.dev.yaml exec golang-app-dev ls -la /app
```

### 依赖管理

```bash
# 更新Go依赖（容器内）
docker-compose -f docker-compose.dev.yaml exec golang-app-dev go mod tidy

# 添加新依赖（容器内）
docker-compose -f docker-compose.dev.yaml exec golang-app-dev go get github.com/example/package
```

## 🎯 开发流程

### 1. 典型开发循环

```bash
# 1. 启动开发环境
./devrun.sh

# 2. 编辑代码
vim src/api/chat.go

# 3. Air自动重新编译（无需手动操作）
# 4. 测试API
curl http://localhost:8082/api/chat

# 5. 查看日志
docker-compose -f docker-compose.dev.yaml logs -f golang-app-dev
```

### 2. 添加新功能

```bash
# 1. 创建新文件
touch src/api/new_feature.go

# 2. 编辑文件（Air会自动监控）
vim src/api/new_feature.go

# 3. 更新路由（如果需要）
vim src/main.go

# 4. Air自动重启服务
# 5. 测试新功能
```

### 3. 调试技巧

```bash
# 查看编译错误
docker-compose -f docker-compose.dev.yaml logs golang-app-dev | grep "build failed"

# 查看Air状态
docker-compose -f docker-compose.dev.yaml exec golang-app-dev ps aux | grep air

# 手动触发重新编译（修改任意监控文件）
touch src/main.go
```

## 📊 性能监控

### 编译时间监控

Air会显示每次编译的时间：

```
building...
built in 1.234s
```

### 内存使用监控

```bash
# 查看容器资源使用
docker stats

# 查看特定容器
docker stats $(docker-compose -f docker-compose.dev.yaml ps -q golang-app-dev)
```

## 🚨 故障排除

### 1. Air无法启动

```bash
# 检查Air是否安装
docker-compose -f docker-compose.dev.yaml exec golang-app-dev which air

# 检查配置文件
docker-compose -f docker-compose.dev.yaml exec golang-app-dev cat .air.toml

# 重新构建容器
docker-compose -f docker-compose.dev.yaml build --no-cache golang-app-dev
```

### 2. 热重载不工作

```bash
# 检查文件挂载
docker-compose -f docker-compose.dev.yaml exec golang-app-dev ls -la /app

# 检查Air进程
docker-compose -f docker-compose.dev.yaml exec golang-app-dev ps aux | grep air

# 查看Air日志
docker-compose -f docker-compose.dev.yaml logs golang-app-dev | grep "watching"
```

### 3. 编译失败

```bash
# 查看详细错误
docker-compose -f docker-compose.dev.yaml logs golang-app-dev

# 检查Go模块
docker-compose -f docker-compose.dev.yaml exec golang-app-dev go mod verify

# 清理并重新下载依赖
docker-compose -f docker-compose.dev.yaml exec golang-app-dev rm -rf /go/pkg/mod
docker-compose -f docker-compose.dev.yaml exec golang-app-dev go mod download
```

### 4. 端口冲突

```bash
# 检查端口占用
lsof -i :8080
lsof -i :8082

# 修改端口（docker-compose.dev.yaml）
ports:
  - "8083:80"  # 改为8083
```

### 5. 权限问题

```bash
# 检查文件权限
ls -la src/

# 修复权限（如果需要）
chmod -R 755 src/
```

## ⚡ 性能优化

### 1. 减少不必要的重编译

在 `.air.toml` 中配置：

```toml
[build]
  # 只在真正变化时重新编译
  exclude_unchanged = true
  
  # 减少延迟
  delay = 500
```

### 2. 优化挂载

```yaml
# 使用缓存挂载提升性能
volumes:
  - ./src:/app:cached
  - go-cache:/go/pkg/mod
```

### 3. 并行编译

```bash
# 设置编译并行度
export GOMAXPROCS=4
```

## 📚 相关工具

### Air替代方案

- **realize** - 另一个Go热重载工具
- **fresh** - 轻量级热重载工具
- **gin** - 简单的热重载工具

### 调试工具

```bash
# 安装Delve调试器（容器内）
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试模式
dlv debug ./main.go --listen=:2345 --headless=true --api-version=2
```

### 代码质量工具

```bash
# 安装golangci-lint
docker-compose -f docker-compose.dev.yaml exec golang-app-dev sh -c "
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin v1.54.2
"

# 运行代码检查
docker-compose -f docker-compose.dev.yaml exec golang-app-dev golangci-lint run
```

## 🌐 访问地址

开发环境启动后，可以通过以下地址访问：

- **前端应用**: http://localhost:8082
- **Go API**: http://localhost:8080 (容器内部)
- **Redis管理**: 使用Redis客户端连接 `localhost:6379`
- **MySQL管理**: 使用MySQL客户端连接 `localhost:3306`

## 🔄 生产环境对比

| 特性 | 开发环境 | 生产环境 |
|------|----------|----------|
| 启动方式 | `docker-compose.dev.yaml` | `docker-compose.yaml` |
| 热重载 | ✅ Air支持 | ❌ 预编译二进制 |
| 源码挂载 | ✅ 实时同步 | ❌ 镜像内置 |
| 日志级别 | DEBUG | INFO/ERROR |
| 性能 | 开发优化 | 生产优化 |
| 安全性 | 开发友好 | 生产安全 |

## 📝 最佳实践

### 1. 代码组织

```
src/
├── api/          # API路由处理器
├── config/       # 配置管理
├── middleware/   # 中间件
├── models/       # 数据模型
├── repository/   # 数据访问层
├── services/     # 业务逻辑层
└── utils/        # 工具函数
```

### 2. 开发流程

1. **启动环境**: `./devrun.sh`
2. **编写代码**: 修改src/目录下文件
3. **自动测试**: Air自动重新编译
4. **验证功能**: 测试API接口
5. **查看日志**: 监控应用日志
6. **提交代码**: Git提交变更

### 3. 调试技巧

- 使用`fmt.Println()`进行简单调试
- 查看Air编译日志定位语法错误
- 使用Postman/curl测试API接口
- 监控Redis/MySQL数据变化

## 🚀 总结

通过Air热重载功能，开发效率可以显著提升：

- **零配置热重载** - 修改代码自动生效
- **快速反馈循环** - 秒级编译和重启
- **完整开发环境** - 数据库、缓存一体化
- **生产环境一致性** - Docker保证环境一致

开始享受高效的Go开发体验吧！🎉 