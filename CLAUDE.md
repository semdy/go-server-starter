# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build
go build ./...

# Run (dev mode, reads configs/config.yml + configs/config.dev.yml)
go run cmd/server/server.go -mode=dev

# Hot reload (requires Air)
air

# Run with env var overrides
APP_MODE=dev go run cmd/server/server.go
```

## Test

```bash
# All tests
go test ./...

# Single package with verbose output
go test ./internal/service/... -v -count=1

# Single test
go test ./internal/service/... -run TestVerifyCode -v
```

## Local Dev Environment

```bash
# Start MySQL + Redis dependencies only (app runs on host via air)
docker compose -f docker-compose.dev.yml up -d

# Stop
docker compose -f docker-compose.dev.yml down

# Production-style (app in Docker too)
docker compose up -d
```

Secrets: copy `configs/config.dev.yml.example` to `configs/config.dev.yml` (gitignored) and fill in local passwords. Environment variables with prefix `APP_` (e.g. `APP_DATABASE_PASSWORD`) override config file values.

## Architecture

**Layered architecture** with strict inward dependency flow:

```
Handler (HTTP binding, response) → Service (business logic, transactions) → Repository (data access)
```

- `internal/` is private application code; `pkg/` is reusable libraries.
- All layers communicate through interfaces defined alongside implementations.
- Dependency injection is manual in `internal/app/app.go` — `App.Start()` wires everything in explicit order.
- There is no DI framework.

### Service layer — `context.Context`, not `*gin.Context`

Service methods accept `context.Context` (stdlib) and explicit parameters. The custom `*ctx.Context` is only used in handlers and middleware. This keeps services testable with `context.Background()` and reusable outside HTTP (gRPC, CLI, task workers).

```go
// Service — framework-agnostic
func (s *UserServiceImpl) GetByID(ctx context.Context, id uint64) (*model.User, *exception.Exception)

// Handler — extracts values from Gin, then calls service
func (h *UserHandlerImpl) GetInfo(c *gin.Context) {
    appCtx := ctx.FromGinCtx(c)
    user, _ := h.service.User().GetByID(appCtx.Ctx, id)
    appCtx.ToSuccess(user)
}
```

### Key Patterns

**Generic BaseRepo** (`internal/repo/base.go`): `BaseRepo[T]` provides full CRUD for any model type. Concrete repos embed it and add entity-specific methods.

**QueryOption functions** (`internal/repo/repo_option.go`): composable `func(*gorm.DB) *gorm.DB` passed as variadic args. Column-name functions (`WhereAutoLike*`) validate against a safe regex and backtick-quote — use `quoteColumnName()` for new ones.

**Custom Context** (`internal/ctx/ctx.go`): wraps `context.Context` + `*gin.Context`. Used only in handlers/middleware. Provides `ShouldBind` (i18n validation errors), `ToSuccess`/`ToError` (JSON `{code, message, data/details}` envelope), and `GetUserUniCode`/`GetDeviceType` extraction.

**Exception system** (`internal/exception/`): domain errors carry HTTP status + business code + i18n messages. Scoped codes: Common=1000, User=20000, UserRole=21000. Service returns `(*Result, *exception.Exception)`; handler converts to JSON via `ctx.ToError()`.

**i18n** (`internal/i18n/`): `Text{En, Zh}` structs. Locale from `locale` header or `Accept-Language`.

**Swagger** (`internal/handler/*.go`): swaggo annotations on handler methods. Generated docs at `docs/`. Regenerate with `swag init -g cmd/server/server.go -o docs`. UI at `/api/swagger/index.html`.

### Middleware Stack (order matters)

1. `gin.Recovery()` — panic recovery
2. `ZapLogger` — request logging (redacts Authorization/Cookie)
3. `ZapRecovery` — panic recovery with stack traces in dev mode
4. `CORS` — dev mode only
5. `Translations` — locale extraction
6. `RateLimit(100)` — per-IP, 100 req/min via Redis GCRA algorithm

Route-group middleware: `jwt.JWT()` (token validation + auto-refresh when < 1/3 remaining), `auth.RoleCheckAny()` / `auth.RoleCheckAll()` for RBAC.

### RBAC

JWT carries `uniCode` in claims. `pkg/auth/` middleware fetches user roles via `UserRoleService.GetCachedRolesCodeByUniCode` (Redis cache-aside with 5min TTL + singleflight dedup). Role changes invalidate the cache. Routes declare access inline:

```go
router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin))
```

### Database

- MySQL via GORM; database auto-created (`CREATE DATABASE IF NOT EXISTS`).
- **Migrations**: [goose](https://github.com/pressly/goose) with embedded SQL in `internal/database/migration/migrations/`. Run on startup via `migration.Run(sqlDB)`. Each migration has `.up.sql` + `.down.sql`. GORM AutoMigrate is no longer used.
- **Seed**: inserts 6 default roles (idempotent).
- To add a table: create a goose migration pair — the `embed` directive picks it up automatically.

### Config Loading

**koanf v2** (migrated from Viper). Priority: `DefaultConfig` → `config.yml` → `config.{mode}.yml` → env vars.

- Env vars use `APP_` prefix. All-lowercase keys (e.g. `database.password`) map automatically. CamelCase keys are explicitly bound via `bindEnv()` in `config.go`. To add a new camelCase key, add it to the `bindEnv` call.
- Struct tags use `koanf:"..."`.

### Task Queue (Asynq)

`pkg/taskq/` wraps Asynq with Redis-backed workers. Start-up flow in `app.go`:

```
taskqClient + taskqServer → register handlers → go taskqServer.Start()
```

Key features:
- **`EnqueueUnique`**: deduplicates tasks by `(TaskID, Unique TTL)` via Redis SETNX.
- **`RetryByType`**: per-task-type `MaxRetry` + `Timeout`.
- **`HandlerDeps`**: global dependency bag set in `app.go` (SMS/Email senders, template codes).
- **`Alerter`**: pluggable interface called when retries are exhausted (write to DB, Slack, etc.).
- **`Server.SetAlerter()`**: thread-safe late binding — set after service layer is initialized.

Adding a task: define constant + payload in `tasks.go` → constructor → handler → register in `app.go` → enqueue from service.

### Dead Letters (MySQL-backed)

Retry-exhausted tasks are persisted to MySQL `dead_letters` table via `DeadLetterService.Alert()` (implements `taskq.Alerter`). Admin API at `/api/admin/dead-letters`:

```
GET  ?taskType=email:welcome    — list
POST /retry  {"id": 5}          — re-enqueue + mark is_retried
POST /retry-all?taskType=...    — batch retry
DELETE  {"id": 5}                — hard delete
```

### Verification Codes & Notifications

`pkg/verify_code/` stores codes in Redis (5min TTL, 60s resend cooldown). Sending is async via taskq:

```
POST /api/auth/send-sms-code → generate code → Redis SETEX → taskq.EnqueueUnique → Alibaba Cloud SMS
POST /api/auth/login/mobile  → Redis GET compare → DEL on match → JWT
```

`pkg/notify/` provides `SmsSender` and `EmailSender` interfaces. Alibaba Cloud implementations in `alisms.go` / `alimail.go`. Dev default is `LogSender` (no-op). Templates embedded via `embed.FS` in `pkg/notify/template/`.

### Adding a New Entity

1. Create goose migration (`0000x_name.{up,down}.sql`)
2. Define model in `internal/model/`
3. (Optional) Create repo in `internal/repo/`
4. (Optional) Create DTOs in `internal/dto/`
5. Create service in `internal/service/`
6. Create handler in `internal/handler/`
7. Register routes in `internal/router/`
8. Wire into `Handler`/`Service`/`Repo` aggregation interfaces

The `generate.sh` script scaffolds model/repo/dto/service/handler boilerplate.

### Key Dependencies

| Purpose | Library |
|---------|---------|
| HTTP framework | Gin v1.11 |
| ORM | GORM + MySQL driver |
| Redis | go-redis/v9 |
| JWT | golang-jwt/v5 |
| Logging | Zap + Lumberjack (rotation) |
| Config | koanf v2 (env prefix: `APP_`) |
| Task queue | Asynq + Redis AOF persistence |
| Migrations | goose (embedded SQL) |
| ID generation | bwmarrin/snowflake |
| Validation | go-playground/validator/v10 |
| SMS/Email | Alibaba Cloud SDK (pluggable) |
| API docs | swaggo/swag + gin-swagger |
| Dev hot reload | Air (`.air.toml`) |

## Notes

- **Secrets**: `configs/config.{dev,test,prod}.yml` and `.env` are gitignored. Use `config.dev.yml.example` as template. Env vars with `APP_` prefix override config file values. Alibaba Cloud credentials go in config file or env vars.
- **Zap vs slog**: Zap is kept. Multi-core output (info.log / error.log + console) with Lumberjack rotation cannot be easily replicated with slog.
- **Config struct tags**: Use `koanf:"..."`. CamelCase keys need a `bindEnv` entry in `config.go`.
- **QueryOption column safety**: `WhereAutoLike*` validate column names via `quoteColumnName()`. Follow this pattern for new dynamic-column functions.
- **Service layer**: Services accept `context.Context`, not `*gin.Context`. The custom `*ctx.Context` stays in handlers/middleware.
- **Test mocks**: Stub implementations embed the real repo interfaces and override only needed methods. See `*_test.go` in `internal/service/`.
- **Redis persistence**: docker-compose enables AOF (`--appendonly yes --appendfsync everysec`) to survive restarts without losing tasks.
