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
│   └── server/              # 应用程序入口
├── configs/                 # 配置文件
│   ├── config.yml           # 默认配置（不含密钥）
│   ├── config.dev.yml       # 本地开发覆盖（gitignored）
│   ├── config.dev.yml.example  # 开发配置模板
│   └── config.test.yml      # 测试配置（gitignored）
├── docs/                    # Swagger 文档生成目录（已提交）
├── internal/
│   ├── app/                 # 应用初始化 & 依赖注入
│   ├── config/              # 配置结构体定义（koanf v2）
│   ├── constant/            # 常量（Redis key、context key）
│   ├── ctx/                 # 自定义请求上下文（Gin 封装）
│   ├── database/
│   │   └── migration/       # Goose 迁移文件（embed SQL）
│   ├── dto/                 # 数据传输对象
│   ├── enum/                # 枚举类型
│   ├── exception/           # 领域异常（含 i18n）
│   ├── handler/             # HTTP 处理器（控制器）
│   ├── i18n/                # 国际化
│   ├── middleware/           # HTTP 中间件
│   ├── model/               # 数据库模型（GORM）
│   ├── repo/                # 仓储层（数据访问）
│   ├── router/              # 路由定义
│   ├── seed/                # 数据库种子数据
│   └── service/             # 业务逻辑层
├── pkg/
│   ├── asyn_queue/          # Asynq 客户端/服务端
│   ├── auth/                # RBAC 中间件
│   ├── database/            # MySQL 连接 & 自动建库
│   ├── jwt/                 # JWT 令牌管理
│   ├── logger/              # Zap 日志 + Lumberjack 轮转
│   ├── redis/               # Redis 客户端
│   ├── snowflake/           # 雪花 ID 生成器
│   ├── translator/          # 验证器翻译（zh/en）
│   ├── utils/               # 通用工具
│   └── validator/           # 自定义验证规则
├── .air.toml                # Air 热重载配置
├── .env.example             # 环境变量模板
├── CLAUDE.md                # Claude Code 指引
├── Dockerfile               # 多阶段构建
├── docker-compose.yml       # 生产环境 Docker Compose
├── docker-compose.dev.yml   # 开发环境（仅 MySQL + Redis）
├── generate.sh              # 模块脚手架生成器
├── go.mod
├── go.sum
└── logs/                    # 日志文件
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
| GET | `/api/user/admin/table` | 获取用户列表（分页） | admin+ |
| GET | `/api/role/table` | 角色列表（分页） | admin+ |
| GET | `/api/role/{id}` | 查询角色详情 | admin+ |
| POST | `/api/role` | 创建角色 | admin+ |
| PUT | `/api/role/{id}` | 更新角色 | admin+ |
| DELETE | `/api/role/{id}` | 删除角色（软删除） | admin+ |
| GET | `/api/admin/tasks/archived` | 死信任务列表（Redis） | super_admin |
| POST | `/api/admin/tasks/archived/run` | 重试单个死信任务 | super_admin |
| POST | `/api/admin/tasks/archived/run-all` | 重试全部死信任务 | super_admin |
| DELETE | `/api/admin/tasks/archived` | 删除死信任务 | super_admin |
| GET | `/api/admin/dead-letters` | 死信列表（DB） | super_admin |
| POST | `/api/admin/dead-letters/retry` | 重试单条死信 | super_admin |
| POST | `/api/admin/dead-letters/retry-all` | 按类型批量重试 | super_admin |
| DELETE | `/api/admin/dead-letters` | 删除死信 | super_admin |

### Swagger UI

启动服务后访问 `http://localhost:8080/api/swagger/index.html` 即可使用交互式 API 文档。点击右上角 **Authorize** 按钮，输入 `Bearer {token}` 即可测试需要认证的接口。

修改 API 注解后重新生成文档：

```bash
swag init -g cmd/server/server.go -o docs
```

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

## 🗄️ 数据库迁移

使用 [goose](https://github.com/pressly/goose) 管理数据库迁移。所有迁移文件通过 `embed` 打包进二进制，启动时自动执行——生产环境无需额外 CLI 工具。

### 迁移文件

位于 `internal/database/migration/migrations/`，每个迁移包含 `.up.sql` 和 `.down.sql` 一对文件：

```
migrations/
├── 00001_create_user_roles.up.sql
├── 00001_create_user_roles.down.sql
├── 00002_create_users.up.sql
├── 00002_create_users.down.sql
├── 00003_create_user_role_refs.up.sql
└── 00003_create_user_role_refs.down.sql
```

### 添加新迁移

```bash
# 生成空迁移文件
goose -dir internal/database/migration/migrations create add_new_table sql
```

在生成的文件中编写 DDL，重启服务即可——`migrate.go` 中的 `embed` 指令会自动发现新文件，无需修改 Go 代码。

### 回滚

```go
migration.Down(sqlDB)  // 回滚最近一次迁移
```

goose 在 `goose_db_version` 表中记录已执行的迁移版本，确保不重复执行，并支持完整回滚。

## 📱 验证码与通知

验证码存储在 Redis 中，5 分钟有效期 + 60 秒重复请求冷却。发送通过任务队列异步完成。

### 流程

```
POST /api/auth/send-sms-code  →  生成 6 位验证码  →  Redis SETEX  →  taskq.EnqueueUnique  →  阿里云短信
POST /api/auth/login/mobile   →  Redis GET 比对    →  匹配成功 DEL  →  JWT token
```

### 阿里云集成

`pkg/notify/` 提供可插拔的 `SmsSender` 和 `EmailSender` 接口。阿里云实现（`AlibabaSmsSender`、`AlibabaEmailSender`）在 `app.go` 中注入。替换服务商只需实现接口并修改注入代码。

本地开发使用 `LogSender`（仅打印日志），无需阿里云凭证。

### 邮件模板

模板文件通过 `embed.FS` 打包进二进制。欢迎邮件模板位于 `pkg/notify/template/templates/welcome_email.html`。新增模板只需在该目录添加 `.html` 文件即可自动识别。

渲染模板：

```go
import notifytmpl "go-server-starter/pkg/notify/template"

html, _ := notifytmpl.GetEngine().Render("welcome_email.html", data)
```

## 📨 任务队列（异步任务）

基于 [Asynq](https://github.com/hibiken/asynq)，以 Redis 为存储后端。Worker 与 HTTP 服务同时启动，通过 `pkg/taskq` 入队和消费任务。

### 添加新任务

1. 在 `pkg/taskq/tasks.go` 中定义任务类型常量和 payload struct
2. 编写构造函数（`NewXxxTask`）
3. 编写处理函数（`HandleXxx`）
4. 在 `app.go` 中注册：`taskqServer.HandleFunc(taskq.TaskXxx, taskq.HandleXxx)`
5. 在业务代码中入队：

```go
task, _ := taskq.NewEmailWelcomeTask(taskq.EmailWelcomePayload{
    UserUniCode: user.UniCode,
    Email:       user.Email,
})
uniqueKey := taskq.WelcomeEmailUniqueKey(user.UniCode)
s.taskq.EnqueueUnique(ctx, task, uniqueKey, 24*time.Hour,
    taskq.RetryByType(taskq.TaskEmailWelcome)...)
```

### 重试、幂等与告警

- **按任务定制重试**：`RetryByType(taskType)` 返回该任务专属的 `MaxRetry` 和 `Timeout`。
- **幂等**：`EnqueueUnique` 基于 `asynq.Unique`（Redis SETNX），同一 key 在 TTL 窗口内不会重复入队。多次调用安全。
- **重试耗尽**：`ErrorHandler` 记录错误日志，并调用 `Alerter` 接口。
- **Alerter 接口**：可插拔——实现 `Alerter` 接口即可接入 Slack、Webhook 或写入死信表。通过 `taskq.NewServer(..., alerter)` 传入，默认不告警。

### 监控与死信管理

**控制台**：安装 [asynqmon](https://github.com/hibiken/asynqmon)，指向 AsynQ 的 Redis DB：

```bash
go install github.com/hibiken/asynq/tools/asynqmon@latest
asynqmon --redis-addr=localhost:6379 --redis-password=root --redis-db=1
# 浏览器打开 http://localhost:8081
```

**手动重试**已归档（重试耗尽）的任务，通过 `taskq.Client`：

```go
tasks, _ := client.ListArchivedTasks("default")
client.RunArchivedTask("default", taskID)   // 重试单个
client.RunAllArchivedTasks("default")       // 全部重试
client.DeleteArchivedTask("default", taskID) // 永久删除
```

### 优雅关闭

Worker 会等待正在执行的任务完成后才退出：

```go
a.taskqServer.Shutdown()  // 排空后停止
a.taskqClient.Close()     // 关闭 Redis 连接
```

重试耗尽后任务被 Asynq 归档，可通过 `asynqmon` 查看。

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
