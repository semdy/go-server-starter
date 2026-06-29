# Go Server Starter

[中文文档](./README.zh_CN.md)

A production-ready Go web server boilerplate/starter kit with clean architecture, built-in authentication, RBAC, and comprehensive tooling.

## ✨ Features

- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **Database**: MySQL with [GORM](https://gorm.io/) ORM, auto-migration support
- **Cache**: [Redis](https://github.com/redis/go-redis) integration
- **Authentication**: JWT-based auth with multi-device token expiration support
- **Authorization**: Role-Based Access Control (RBAC)
- **Validation**: Request validation via [go-playground/validator](https://github.com/go-playground/validator)
- **Internationalization**: i18n support (English & Chinese)
- **Logging**: Structured logging with [Zap](https://github.com/uber-go/zap) + log rotation with [Lumberjack](https://github.com/natefinch/lumberjack)
- **Configuration**: Environment-based config management with [Viper](https://github.com/spf13/viper)
- **Async Tasks**: Background job processing with [Asynq](https://github.com/hibiken/asynq)
- **ID Generation**: Distributed ID generation with [Snowflake](https://github.com/bwmarrin/snowflake)
- **Graceful Shutdown**: Clean server shutdown handling
- **Clean Architecture**: Layered structure (Handler → Service → Repository)

## 📁 Project Structure

```
go-server-starter/
├── cmd/
│   └── server/              # Application entry point
├── configs/                 # Configuration files
│   ├── config.yml           # Default config (no secrets)
│   ├── config.dev.yml       # Local dev overrides (gitignored)
│   ├── config.dev.yml.example  # Dev config template
│   └── config.test.yml      # Test config (gitignored)
├── docs/                    # Swagger generated docs (committed)
├── internal/
│   ├── app/                 # Application initialization & DI
│   ├── config/              # Config structs & loading (koanf v2)
│   ├── constant/            # Constants (Redis keys, context keys)
│   ├── ctx/                 # Custom request context (Gin wrapper)
│   ├── database/
│   │   └── migration/       # Goose migrations (embedded SQL)
│   ├── dto/                 # Data Transfer Objects
│   ├── enum/                # Enumerations
│   ├── exception/           # Domain exceptions with i18n
│   ├── handler/             # HTTP handlers (controllers)
│   ├── i18n/                # Internationalization
│   ├── middleware/           # HTTP middlewares
│   ├── model/               # Database models (GORM)
│   ├── repo/                # Repository layer (data access)
│   ├── router/              # Route definitions
│   ├── seed/                # Database seeders
│   └── service/             # Business logic layer
├── pkg/
│   ├── asyn_queue/          # Asynq client/server
│   ├── auth/                # RBAC middleware
│   ├── database/            # MySQL connection & auto-create DB
│   ├── jwt/                 # JWT token management
│   ├── logger/              # Zap logger + Lumberjack rotation
│   ├── redis/               # Redis client
│   ├── snowflake/           # Snowflake ID generator
│   ├── translator/          # Validator translator (zh/en)
│   ├── utils/               # Common utilities
│   └── validator/           # Custom validation rules
├── .air.toml                # Air hot reload config
├── .env.example             # Environment variables template
├── CLAUDE.md                # Claude Code guidance
├── Dockerfile               # Multi-stage build
├── docker-compose.yml       # Production Docker Compose
├── docker-compose.dev.yml   # Dev Docker Compose (MySQL + Redis only)
├── generate.sh              # Entity scaffold generator
├── go.mod
├── go.sum
└── logs/                    # Log files
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### Installation

1. Clone the repository:

```bash
git clone https://github.com/your-username/go-server-starter.git
cd go-server-starter
```

2. Install dependencies:

```bash
go mod download
```

3. Configure the application:

Copy and modify the config file for your environment:

```bash
cp configs/config.yml configs/config.dev.yml
```

Edit `configs/config.dev.yml` with your database and Redis settings.

4. Run the server:

```bash
# Development mode
go run cmd/server/server.go -mode=dev

# Production mode
go run cmd/server/server.go -mode=prod

# Test mode
go run cmd/server/server.go -mode=test
```

The server will start on `http://localhost:8080` by default.

## ⚙️ Configuration

Configuration is managed via YAML files. Key settings include:

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

## 🔐 Authentication & Authorization

### JWT Authentication

The project uses JWT for stateless authentication with device-specific token expiration:

- **Web**: 24 hours
- **Mobile/Desktop**: 15 days
- **Chrome Extension**: 30 days
- **API**: 48 hours

### Role-Based Access Control

Built-in roles:
- `super_admin` - Super administrator
- `admin` - Administrator
- `user` - Regular user
- `user_vip` - VIP user
- `user_svip` - SVIP user
- `guest` - Guest user

Protect routes with role checks:

```go
// Any of the specified roles
router.GET("/admin", auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin), handler)

// All specified roles required
router.GET("/super", auth.RoleCheckAll(enum.RoleCodeSuperAdmin), handler)
```

## 🌐 API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/hello` | Health check | No |
| POST | `/api/auth/send-sms-code` | Send SMS verification code | No |
| POST | `/api/auth/send-email-code` | Send email verification code | No |
| POST | `/api/auth/login/mobile` | Login via mobile + code | No |
| POST | `/api/auth/login/email` | Login via email + code | No |
| GET | `/api/user/info` | Get current user info | Yes |
| PUT | `/api/user/info` | Update user info | Yes |
| GET | `/api/user/admin/table` | Get users list (paginated) | admin+ |
| GET | `/api/role/table` | List roles (paginated) | admin+ |
| GET | `/api/role/{id}` | Get role by ID | admin+ |
| POST | `/api/role` | Create role | admin+ |
| PUT | `/api/role/{id}` | Update role | admin+ |
| DELETE | `/api/role/{id}` | Delete role (soft) | admin+ |
| GET | `/api/admin/tasks/archived` | List dead-letter tasks (Redis) | super_admin |
| POST | `/api/admin/tasks/archived/run` | Retry one dead-letter task | super_admin |
| POST | `/api/admin/tasks/archived/run-all` | Retry all dead-letter tasks | super_admin |
| DELETE | `/api/admin/tasks/archived` | Delete one dead-letter task | super_admin |
| GET | `/api/admin/dead-letters` | List dead letters (DB) | super_admin |
| POST | `/api/admin/dead-letters/retry` | Retry one dead letter | super_admin |
| POST | `/api/admin/dead-letters/retry-all` | Retry all by type | super_admin |
| DELETE | `/api/admin/dead-letters` | Delete dead letter | super_admin |

### Swagger UI

Start the server and open `http://localhost:8080/api/swagger/index.html` for interactive API documentation. Use the **Authorize** button to set `Bearer {token}` for authenticated endpoints.

To regenerate docs after changing API annotations:

```bash
swag init -g cmd/server/server.go -o docs
```

## 🌍 Internationalization

The API supports multiple languages via the `Accept-Language` header:

```bash
# English
curl -H "Accept-Language: en" http://localhost:8080/api/hello

# Chinese
curl -H "Accept-Language: zh" http://localhost:8080/api/hello
```

## 📝 Logging

Logs are structured with Zap and automatically rotated:

- **Log levels**: debug, info, warn, error, fatal
- **Output**: Console (dev) + File (info.log, error.log)
- **Rotation**: Configurable max size, age, and backup count

## 🗄️ Database Migration

Database migrations are managed by [goose](https://github.com/pressly/goose). All migration files are embedded in the binary and run automatically on startup — no external CLI required in production.

### Migration files

Located at `internal/database/migration/migrations/`, each migration is a pair of `.up.sql` and `.down.sql` files:

```
migrations/
├── 00001_create_user_roles.up.sql
├── 00001_create_user_roles.down.sql
├── 00002_create_users.up.sql
├── 00002_create_users.down.sql
├── 00003_create_user_role_refs.up.sql
└── 00003_create_user_role_refs.down.sql
```

### Adding a new migration

```bash
# Generate empty migration files
goose -dir internal/database/migration/migrations create add_new_table sql
```

Write your DDL in the generated file, then restart the server — migrations run on every startup. The `embed` directive in `migrate.go` picks up new files automatically with zero Go code changes.

### Rollback

```go
migration.Down(sqlDB)  // roll back the last migration
```

`goose` tracks applied versions in a `goose_db_version` table, ensuring migrations are never run twice and supporting full rollback.

## 📱 Verification Codes & Notifications

Verification codes are stored in Redis with a 5-min TTL and a 60-second resend cooldown. Sending is handled asynchronously via the task queue.

### Flow

```
POST /api/auth/send-sms-code  →  generate 6-digit code  →  Redis SETEX  →  taskq.EnqueueUnique  →  Alibaba Cloud SMS
POST /api/auth/login/mobile   →  Redis GET compare     →  DEL on match →  JWT token
```

### Alibaba Cloud integration

`pkg/notify/` provides pluggable `SmsSender` and `EmailSender` interfaces. The Alibaba Cloud implementations (`AlibabaSmsSender`, `AlibabaEmailSender`) are wired in `app.go`. To swap providers, implement the interface and change the wiring.

For local development, `LogSender` prints to console — no Alibaba Cloud credentials needed.

### Email templates

Templates are HTML files embedded in the binary via `embed.FS`. The welcome email is at `pkg/notify/template/templates/welcome_email.html`. Add new templates as `.html` files in that directory — they are picked up automatically.

To render a template programmatically:

```go
import notifytmpl "go-server-starter/pkg/notify/template"

html, _ := notifytmpl.GetEngine().Render("welcome_email.html", data)
```

## 📨 Task Queue (Async Jobs)

[Asynq](https://github.com/hibiken/asynq) provides Redis-backed task processing. Workers start alongside the HTTP server, and tasks are enqueued via `pkg/taskq`.

### Adding a new task

1. Define a task type constant and payload struct in `pkg/taskq/tasks.go`
2. Create a constructor function (`NewXxxTask`)
3. Write a handler function (`HandleXxx`)
4. Register it in `app.go`: `taskqServer.HandleFunc(taskq.TaskXxx, taskq.HandleXxx)`
5. Enqueue from your service:

```go
task, _ := taskq.NewEmailWelcomeTask(taskq.EmailWelcomePayload{
    UserUniCode: user.UniCode,
    Email:       user.Email,
})
uniqueKey := taskq.WelcomeEmailUniqueKey(user.UniCode)
s.taskq.EnqueueUnique(ctx, task, uniqueKey, 24*time.Hour,
    taskq.RetryByType(taskq.TaskEmailWelcome)...)
```

### Retry, idempotency & alerting

- **Per-task retry**: `RetryByType(taskType)` returns `MaxRetry` + `Timeout` tuned per task.
- **Idempotency**: `EnqueueUnique` uses `asynq.Unique` (Redis SETNX) to deduplicate tasks with the same key within a TTL window. Safe to call multiple times.
- **Exhausted retries**: `ErrorHandler` logs the failure and calls the `Alerter` interface.
- **Alerter**: pluggable — implement the `Alerter` interface to send Slack, webhook, or write to a dead-letter table. Pass your implementation to `taskq.NewServer(..., alerter)`. Default is no-op.

### Monitoring & dead-letter

**Dashboard**: install [asynqmon](https://github.com/hibiken/asynqmon) and point it at the AsynQ Redis DB:

```bash
go install github.com/hibiken/asynq/tools/asynqmon@latest
asynqmon --redis-addr=localhost:6379 --redis-password=root --redis-db=1
# open http://localhost:8081
```

**Manual retry** of archived (retry-exhausted) tasks via `taskq.Client`:

```go
tasks, _ := client.ListArchivedTasks("default")
client.RunArchivedTask("default", taskID)   // retry one
client.RunAllArchivedTasks("default")       // retry all
client.DeleteArchivedTask("default", taskID) // delete permanently
```

### Graceful shutdown

The task worker waits for in-flight tasks to complete before the process exits:

```go
a.taskqServer.Shutdown()  // drain then stop
a.taskqClient.Close()     // close Redis connection
```

Tasks are archived by Asynq after retries are exhausted and can be inspected via `asynqmon`.

## 🛠️ Development

### Generate Code

```bash
./generate.sh
```

### Hot Reload

For development with hot reload, use [Air](https://github.com/cosmtrek/air):

```bash
air
```

### With Docker

  1. startup docker compose
  ``bash
  docker compose -f docker-compose.dev.yml up -d
  ```

  2. startup Go app (hot reload, auto restart on code change)
  ``bash
  air
  ```

  3. stop docker compose
  ``bash
  docker compose -f docker-compose.dev.yml down
  ```

## 📄 License

This project is open-sourced software licensed under the [MIT license](LICENSE).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
