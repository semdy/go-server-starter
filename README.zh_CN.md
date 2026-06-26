# Go Server Starter

[English](./README.md)

一个生产就绪的 Go Web 服务器脚手架/启动模板，采用清晰的分层架构，内置认证、RBAC 权限控制和全面的工具集。

## ✨ 特性

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- **数据库**: MySQL + [GORM](https://gorm.io/) ORM，支持自动迁移
- **缓存**: [Redis](https://github.com/redis/go-redis) 集成
- **认证**: 基于 JWT 的认证，支持多设备 Token 过期时间配置
- **授权**: 基于角色的访问控制 (RBAC)
- **验证**: 使用 [go-playground/validator](https://github.com/go-playground/validator) 进行请求参数验证
- **国际化**: i18n 支持（中文和英文）
- **日志**: 使用 [Zap](https://github.com/uber-go/zap) 结构化日志 + [Lumberjack](https://github.com/natefinch/lumberjack) 日志轮转
- **配置管理**: 使用 [Viper](https://github.com/spf13/viper) 进行多环境配置管理
- **异步任务**: 使用 [Asynq](https://github.com/hibiken/asynq) 进行后台任务处理
- **ID 生成**: 使用 [Snowflake](https://github.com/bwmarrin/snowflake) 雪花算法生成分布式 ID
- **优雅关闭**: 支持服务器优雅关闭
- **清晰架构**: 分层结构（Handler → Service → Repository）

## 📁 项目结构

```
go-server-starter/
├── cmd/
│   └── server/          # 应用程序入口
├── configs/             # 配置文件
│   ├── config.yml       # 默认配置
│   ├── config.dev.yml   # 开发环境配置
│   └── config.test.yml  # 测试环境配置
├── internal/
│   ├── app/             # 应用初始化
│   ├── config/          # 配置结构体定义
│   ├── constant/        # 常量
│   ├── ctx/             # 自定义上下文
│   ├── dto/             # 数据传输对象
│   ├── enum/            # 枚举类型
│   ├── exception/       # 异常处理
│   ├── handler/         # HTTP 处理器（控制器）
│   ├── i18n/            # 国际化
│   ├── middleware/      # HTTP 中间件
│   ├── model/           # 数据库模型
│   ├── repo/            # 仓储层（数据访问）
│   ├── router/          # 路由定义
│   ├── seed/            # 数据库种子数据
│   └── service/         # 业务逻辑层
├── pkg/
│   ├── asyn_queue/      # Asynq 客户端/服务端
│   ├── auth/            # 授权工具
│   ├── database/        # 数据库连接
│   ├── jwt/             # JWT 工具
│   ├── logger/          # 日志配置
│   ├── redis/           # Redis 客户端
│   ├── snowflake/       # 雪花 ID 生成器
│   ├── translator/      # 翻译工具
│   ├── utils/           # 通用工具
│   └── validator/       # 验证规则
└── logs/                # 日志文件
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装

1. 克隆仓库：

```bash
git clone https://github.com/your-username/go-server-starter.git
cd go-server-starter
```

2. 安装依赖：

```bash
go mod download
```

3. 配置应用：

复制并修改配置文件：

```bash
cp configs/config.yml configs/config.dev.yml
```

编辑 `configs/config.dev.yml`，配置你的数据库和 Redis 设置。

4. 运行服务器：

```bash
# 开发模式
go run cmd/server/server.go -mode=dev

# 生产模式
go run cmd/server/server.go -mode=prod

# 测试模式
go run cmd/server/server.go -mode=test
```

服务器默认启动在 `http://localhost:8080`。

## ⚙️ 配置说明

配置通过 YAML 文件管理，主要设置包括：

```yaml
server:
  port: 8080
  readTimeout: 10s
  writeTimeout: 10s
  apiPrefix: "/api"

jwt:
  issuer: go-server-starter
  tokenSecret: your-secret-key
  tokenExpires:
    web: 24h
    mobile: 360h
    desktop: 360h

database:
  host: localhost
  port: 3306
  username: root
  password: your-password
  databaseName: your-db

redis:
  host: localhost
  port: 6379
  password: your-password
  db: 0
```

## 🔐 认证与授权

### JWT 认证

项目使用 JWT 进行无状态认证，支持针对不同设备类型设置不同的 Token 过期时间：

- **Web 端**: 24 小时
- **移动端/桌面端**: 15 天
- **Chrome 扩展**: 30 天
- **API**: 48 小时

### 基于角色的访问控制

内置角色：
- `super_admin` - 超级管理员
- `admin` - 管理员
- `user` - 普通用户
- `user_vip` - VIP 用户
- `user_svip` - SVIP 用户
- `guest` - 访客

使用角色检查保护路由：

```go
// 满足任一指定角色即可访问
router.GET("/admin", auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin), handler)

// 需要满足所有指定角色
router.GET("/super", auth.RoleCheckAll(enum.RoleCodeSuperAdmin), handler)
```

## 🌐 API 接口

| 方法 | 接口 | 描述 | 需要认证 |
|------|------|------|----------|
| GET | `/api/hello` | 健康检查 | 否 |
| POST | `/api/auth/login/mobile` | 手机号 + 验证码登录 | 否 |
| POST | `/api/auth/login/email` | 邮箱 + 验证码登录 | 否 |
| GET | `/api/user/info` | 获取当前用户信息 | 是 |
| PUT | `/api/user/info` | 更新用户信息 | 是 |
| GET | `/api/user/table` | 获取用户列表（分页） | 是 |

## 🌍 国际化

API 支持通过 `Accept-Language` 请求头切换语言：

```bash
# 英文
curl -H "Accept-Language: en" http://localhost:8080/api/hello

# 中文
curl -H "Accept-Language: zh" http://localhost:8080/api/hello
```

## 📝 日志系统

日志使用 Zap 进行结构化记录，并自动轮转：

- **日志级别**: debug, info, warn, error, fatal
- **输出方式**: 控制台（开发环境）+ 文件（info.log, error.log）
- **日志轮转**: 可配置最大大小、保存天数和备份数量

## 🛠️ 开发指南

### 代码生成

```bash
./generate.sh
```

### 热重载

开发时可使用 [Air](https://github.com/cosmtrek/air) 实现热重载：

```bash
air
```

### 日常开发流程

  1. 启动依赖服务
  ``bash
  docker compose -f docker-compose.dev.yml up -d
  ```

  2. 启动 Go 应用（热重载，修改代码自动重启）
  ``bash
  air
  ```

  3. 开发完成后停止依赖服务
  ``bash
  docker compose -f docker-compose.dev.yml down
  ```

## 📄 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。

## 🤝 贡献

欢迎贡献代码！请随时提交 Pull Request。
