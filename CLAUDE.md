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
go test ./internal/service/... -run TestNewAuthService -v
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
Handler (HTTP binding) → Service (business logic, transactions) → Repository (data access)
```

- `internal/` is private application code; `pkg/` is reusable libraries.
- All layers communicate through interfaces defined alongside implementations.
- Dependency injection is manual in `internal/app/app.go` — `App.Start()` wires everything in explicit order. No DI framework.

### Service layer — `context.Context`, not `*gin.Context`

Service methods accept `context.Context` (stdlib) and explicit parameters. The custom `*ctx.Context` is only used in handlers and middleware. Services are testable with `context.Background()` and reusable outside HTTP.

```go
// Service — framework-agnostic
func (s *UserServiceImpl) GetByID(ctx context.Context, id uint64) (*model.User, *exception.Exception)

// Handler — extracts values from Gin, then calls service
func (h *UserHandlerImpl) GetMyInfo(c *gin.Context) {
    appCtx := ctx.FromGinCtx(c)
    user, _ := h.service.User().GetByID(appCtx.Ctx, id)
    appCtx.ToSuccess(user)
}
```

### Tenant ID in context.Context

`internal/ctx/tenant.go` provides `WithTenant(ctx, tid)` / `GetTenantID(ctx)` for storing tenant_id in `context.Context`. JWT middleware injects it from claims. Services call `cctx.GetTenantID(ctx)` to apply `tenantFilter(ctx)` in queries.

### Key Patterns

**Generic BaseRepo** (`internal/repo/base.go`): `BaseRepo[T]` provides full CRUD for any model type. Concrete repos embed it and add entity-specific methods.

**QueryOption functions** (`internal/repo/repo_option.go`): composable `func(*gorm.DB) *gorm.DB` passed as variadic args. Column-name functions (`WhereAutoLike*`) validate against a safe regex and backtick-quote — use `quoteColumnName()` for new ones.

**Custom Context** (`internal/ctx/ctx.go`): wraps `context.Context` + `*gin.Context`. Used only in handlers/middleware. Provides `ShouldBind` (i18n validation errors), `ToSuccess`/`ToError` (JSON `{code, message, data/details}` envelope), `GetUserUniCode`/`GetTenantID`/`GetDeviceType` extraction.

**Exception system** (`internal/exception/`): domain errors carry HTTP status + business code + i18n messages. Scoped codes: Common=1000, User=20000, UserRole=21000. Service returns `(*Result, *exception.Exception)`; handler converts to JSON via `ctx.ToError()`.

**i18n** (`internal/i18n/`): `Text{En, Zh}` structs. `Tf()` convenience method for key-value pairs: `i18n.EchoHello.Tf(ctx.GetLocale(), "name", name)`.

**Swagger**: swaggo annotations on handler methods. Regenerate: `swag init -g cmd/server/server.go -o docs`. UI at `/api/swagger/index.html`.

### Snowflake IDs

All models embed `Model` which has `BeforeCreate` hook — auto-generates Snowflake ID when caller doesn't set one. `model.GenerateID` is set in `app.go` from the Snowflake node. No explicit `snowflake.GenerateID()` calls needed in business code.

### Middleware Stack (order matters)

1. `gin.Recovery()` — panic recovery
2. `ZapLogger` — request logging (redacts Authorization/Cookie; level-guarded with `zap.Check`)
3. `ZapRecovery` — panic recovery with stack traces in dev mode
4. `CORS` — dev mode only
5. `Translations` — locale extraction (`Accept-Language` cached via `sync.Map`)
6. `RateLimit(100)` — per-IP, 100 req/min via Redis GCRA; falls back to local sliding window when Redis is down

Route-group middleware: `jwt.JWT()` (token validation + auto-refresh when < 1/3 remaining), `auth.RoleCheckAny()` / `auth.RoleCheckAll()` for static roles, `auth.RoleCheckAnyS()` / `auth.RoleCheckAllS()` for dynamic (string) roles.

### RBAC

JWT carries `uniCode` + `tenantID` in claims. `pkg/auth/` middleware fetches roles via `UserRoleService.GetCachedRolesCodeByUniCode` (Redis cache-aside with 5min TTL + singleflight dedup). Role mutations invalidate cache. Predefined roles: `super_admin`, `admin`, `guest`, `user`, `user_vip`, `user_svip`. Dynamic roles use string variants `RoleCheckAnyS("editor")`.

### Token Auto-Refresh

When remaining TTL < 1/3 of total, server issues new token in `new-token` response header. Client captures and replaces old token. Overlap window ensures no request is rejected.

### Database

- MySQL via GORM; database auto-created (`CREATE DATABASE IF NOT EXISTS`).
- **Migrations**: [goose](https://github.com/pressly/goose) with embedded SQL in `internal/database/migration/migrations/`. Run on startup via `migration.Run(sqlDB)`. Single-file format with `-- +goose Up` / `-- +goose Down` markers. `goose_db_version` tracks applied versions.
- **Seed**: inserts 6 default roles + default tenant "default" (both idempotent).
- **Indexes/constraints**: defined exclusively in migration SQL. GORM struct tags serve as documentation only.

### Config Loading

**koanf v2**. Priority: `DefaultConfig` → `config.yml` → `config.{mode}.yml` → env vars.

- Env vars use `APP_` prefix. All-lowercase keys map automatically (e.g. `APP_DATABASE_PASSWORD`). CamelCase keys require explicit `bindEnv()` entries in `config.go`.
- Struct tags use `koanf:"..."`. Defaults from `DefaultConfig` struct.

### Task Queue (Asynq)

`pkg/taskq/` wraps Asynq with Redis-backed workers (`--appendonly yes --appendfsync everysec` for persistence).

- **`EnqueueUnique`**: idempotent enqueue via Redis SETNX (TaskID + Unique TTL).
- **`RetryByType`**: per-task-type `MaxRetry` + `Timeout`.
- **`Alerter`**: pluggable interface called when retries exhaust. Currently `DeadLetterService` implements it — writes to MySQL `dead_letters` table.
- **`Server.SetAlerter()`**: thread-safe late binding via `atomic.Pointer`.

### Dead Letters

Two-tier: Asynq auto-archives to Redis (visible in `asynqmon`). Exhausted retries also persist to MySQL `dead_letters` table via `DeadLetterService.Alert()`. Admin API at `/api/admin/dead-letters` supports list, retry, batch retry, and delete.

### Verification Codes & Notifications

`pkg/verify_code/`: Redis-backed codes (5min TTL, 60s resend cooldown). Handler validates code via `VerifyCodeService.Validate()` before calling login.

`pkg/notify/`: `SmsSender`/`EmailSender` interfaces. Alibaba Cloud implementations in `alisms.go`/`alimail.go`. Dev default is `LogSender` (no-op). Email templates embedded via `embed.FS` in `pkg/notify/template/`.

### Cron Scheduler

`pkg/cronjob/`: `robfig/cron` wrapper registered in `register.go`. `app.go` only calls `Register()`/`Start()`/`Stop()`. Add jobs to the `Register()` function.

### Scaffolding

`./generate.sh product` creates a full CRUD module (model/repo/dto/service/handler/router/migration) and auto-registers into `repo.go`/`service.go`/`handler.go`/`router.go`, then runs `gofmt -w` on them. `./generate.sh -d product` deletes. See README for the 12-file breakdown.

### Key Dependencies

| Purpose | Library |
|---------|---------|
| HTTP framework | Gin |
| ORM | GORM + MySQL |
| Redis | go-redis/v9 |
| JWT | golang-jwt/v5 |
| Logging | Zap + Lumberjack |
| Config | koanf v2 (env prefix: `APP_`) |
| Task queue | Asynq (Redis AOF) |
| Migrations | goose (embedded SQL) |
| ID generation | bwmarrin/snowflake (BeforeCreate hook) |
| Validation | go-playground/validator/v10 |
| SMS/Email | Alibaba Cloud SDK (pluggable) |
| Cron | robfig/cron |
| API docs | swaggo/swag + gin-swagger |
| Dev | Air (`.air.toml`) |

## Notes

- **Secrets**: `configs/config.{dev,test,prod}.yml` and `.env` are gitignored. Use `config.dev.yml.example` as template. Env vars with `APP_` prefix override config file values.
- **Zap**: Multi-core output (info.log / error.log + console) with Lumberjack rotation. ZapLogger uses `logger.Check()` to skip allocation when log level is above Info.
- **RateLimit**: Falls back to local sliding window when Redis is unreachable — no cascading 500s.
- **Config tags**: Use `koanf:"..."`. CamelCase keys need `bindEnv` entry.
- **QueryOption safety**: `WhereAutoLike*` validate column names via `quoteColumnName()`.
- **Service layer**: Services accept `context.Context`. The custom `*ctx.Context` stays in handlers/middleware.
- **Test mocks**: Stub implementations embed real repo interfaces, override only needed methods.
- **Redis persistence**: docker-compose enables AOF (`--appendonly yes --appendfsync everysec`).
- **gofmt on save**: `.zed/settings.json` configures `format_on_save`.
- **Snowflake IDs**: `model.BeforeCreate` hook auto-fills; migration SQL has `BIGINT UNSIGNED PRIMARY KEY` without `AUTO_INCREMENT`.
- **Login = register**: New users auto-register with default tenant and "user" role. No explicit registration endpoint.
