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

#### Token Auto-Refresh

When a token's remaining lifetime drops below 1/3 of its total TTL, the server issues a new token in the `new-token` response header. The client should capture it and replace the old token:

```javascript
// axios
axios.interceptors.response.use(res => {
  const newToken = res.headers['new-token']
  if (newToken) localStorage.setItem('token', newToken)
  return res
})

// fetch
const res = await fetch('/api/user/my-info', { headers })
const newToken = res.headers.get('new-token')
if (newToken) localStorage.setItem('token', newToken)
```

The old token remains valid until expiry — no requests are rejected during the overlap window.

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
| GET | `/api/healthz` | Liveness + readiness probe | No |
| POST | `/api/auth/send-sms-code` | Send SMS verification code | No |
| POST | `/api/auth/send-email-code` | Send email verification code | No |
| POST | `/api/auth/login/mobile` | Login via mobile + code | No |
| POST | `/api/auth/login/email` | Login via email + code | No |
| GET | `/api/user/my-info` | Get current user info | Yes |
| PUT | `/api/user/my-info` | Update current user info | Yes |
| GET | `/api/user/admin/table` | List users (paginated) | admin+ |
| GET | `/api/user/admin/{id}` | Get user by ID | admin+ |
| POST | `/api/user/admin` | Create user | admin+ |
| PUT | `/api/user/admin/{id}` | Update user | admin+ |
| DELETE | `/api/user/admin/{id}` | Delete user (soft) | admin+ |
| GET | `/api/role/{id}` | Get role by ID | admin+ |
| GET | `/api/role/table` | List roles (paginated) | admin+ |
| POST | `/api/role` | Create role | admin+ |
| PUT | `/api/role/{id}` | Update role | admin+ |
| DELETE | `/api/role/{id}` | Delete role (soft) | admin+ |
| GET | `/api/admin/tenants/code` | Generate tenant code | super_admin |
| GET | `/api/admin/tenants/{id}` | Get tenant by ID | super_admin |
| GET | `/api/admin/tenants` | List tenants (paginated) | super_admin |
| POST | `/api/admin/tenants` | Create tenant | super_admin |
| PUT | `/api/admin/tenants/{id}` | Update tenant | super_admin |
| DELETE | `/api/admin/tenants/{id}` | Delete tenant (soft) | super_admin |
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

### Migration format

Each migration is a single `.sql` file with `-- +goose Up` / `-- +goose Down` markers:

```
internal/database/migration/migrations/
├── 00001_create_user_roles.sql
├── 00002_create_users.sql
├── 00003_create_user_role_refs.sql
└── 00004_create_dead_letters.sql
```

### Adding a new table

```bash
goose -dir internal/database/migration/migrations create add_new_table sql
```

Write DDL in the generated file, restart — migrations run on startup.

### Adding indexes or constraints

Create a new migration with `ALTER TABLE`:

```sql
-- +goose Up
ALTER TABLE users ADD INDEX idx_status (status);
-- +goose Down
ALTER TABLE users DROP INDEX idx_status;
```

GORM struct tags (`gorm:"index"`, `gorm:"uniqueIndex"`) serve as documentation after goose takes over DDL. Indexes and unique constraints are defined exclusively in migration SQL.

### Rollback

```go
migration.Down(sqlDB)  // roll back the last migration
```

`goose_db_version` table tracks applied versions — never run twice, full rollback support.

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

**Dashboard**: install [asynqmon](https://github.com/hibiken/asynqmon) to inspect Redis queues in real time:

```bash
go install github.com/hibiken/asynq/tools/asynqmon@latest
asynqmon --redis-addr=localhost:6379 --redis-password=root --redis-db=1
# open http://localhost:8081
```

**Dead-letter persistence**: retry-exhausted tasks are automatically written to MySQL (`dead_letters` table) via the `Alerter`. List, retry, or delete via admin API:

```
GET  /api/admin/dead-letters?taskType=email:welcome
POST /api/admin/dead-letters/retry        {"id": 5}
POST /api/admin/dead-letters/retry-all?taskType=email:welcome
```

### Graceful shutdown

The task worker waits for in-flight tasks to complete before the process exits:

```go
a.taskqServer.Shutdown()  // drain then stop
a.taskqClient.Close()     // close Redis connection
```

Tasks are archived by Asynq after retries are exhausted and can be inspected via `asynqmon`.

## ⏰ Cron Scheduler

[robfig/cron](https://github.com/robfig/cron) schedules recurring jobs in-process. All jobs are registered in `pkg/cronjob/register.go` via `cronjob.Register(repo, logger)`. `app.go` only calls `Start()`/`Stop()`.

### Adding a job

Add a `Job` entry in `pkg/cronjob/register.go`:

```go
jobs := []Job{
    {Name: "my-job", Spec: "@every 1h", Fn: myJob(log)},
    // ...
}
```

Then implement `myJob` in the same file. No changes to `app.go`.

Supports 6-field cron (`sec min hour dom month dow`) and interval format (`@every 1h30m`).

### Built-in jobs

| Job | Schedule | Action |
|-----|----------|--------|
| `heartbeat` | `@every 1h` | Log a heartbeat |
| `purge-dead-letters` | `0 0 3 * * *` (3am daily) | Hard-delete retried dead letters older than 30 days |

### Graceful shutdown

```go
a.cronSched.Stop()  // waits for running jobs to finish
```

## 🛠️ Development

### Scaffold a Module

Generate a full CRUD module (model + repo + dto + service + handler + route + migration) in one command:

```bash
./generate.sh product
```

Output:

```
internal/model/product.go                      # data struct
internal/repo/product_repo.go                  # BaseRepo wrapper
internal/dto/product_dto.go                    # request/response types
internal/service/product_service.go            # business logic
internal/handler/product_handler.go            # HTTP binding
internal/router/product_router.go              # routes + permissions
internal/database/migration/migrations/xxx.sql # goose migration
```

The script also prints the exact lines you need to paste into `repo.go`, `service.go`, `handler.go`, and `router.go` to register the new module (3 files, ~4 lines each).

### What gets touched

Adding a full CRUD module touches 12 files:

| # | File | Purpose | Auto? |
|---|------|---------|:--:|
| 1 | `internal/model/xxx.go` | Data struct | ✅ generated |
| 2 | `internal/database/migration/migrations/xxx.sql` | DDL | ✅ generated |
| 3 | `internal/repo/xxx_repo.go` | BaseRepo wrapper (10 lines) | ✅ generated |
| 4 | `internal/repo/repo.go` | Register repo to aggregate | ✏️ 1 line each |
| 5 | `internal/dto/xxx_dto.go` | Request/response types | ✅ generated |
| 6 | `internal/service/xxx_service.go` | Business logic (~100 lines) | ✅ generated |
| 7 | `internal/service/service.go` | Register service to aggregate | ✏️ 1 line each |
| 8 | `internal/handler/xxx_handler.go` | HTTP binding (~80 lines) | ✅ generated |
| 9 | `internal/handler/handler.go` | Register handler to aggregate | ✏️ 1 line each |
| 10 | `internal/router/xxx_router.go` | Routes + permissions (10 lines) | ✅ generated |
| 11 | `internal/router/router.go` | Register routes | ✏️ 1 line |
| 12 | (none) | goose auto-runs on restart | ✅ automatic |

Files marked ✅ are fully scaffolded by `generate.sh`. Files marked ✏️ need one insertion each — the script prints the exact code to paste.

To delete a module: `./generate.sh -d product`.

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
