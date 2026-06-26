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
- Dependency injection is manual in `internal/app/app.go` — `App.Start()` wires everything in explicit order (config → db → redis → repo → seed → jwt → service → handler → router).
- There is no DI framework.

### Key Patterns

**Generic BaseRepo** (`internal/repo/base.go`): `BaseRepo[T]` provides full CRUD for any model type. Concrete repos (`UserRepo`, `UserRoleRepo`) embed it and add entity-specific methods. New models only need the concrete repo if they have custom queries beyond CRUD.

**QueryOption functions** (`internal/repo/repo_option.go`): composable `func(*gorm.DB) *gorm.DB` functions (`Where`, `Order`, `Preload`, `Paginate`, etc.) passed as variadic args to repo methods. Chainable and reusable.

**Custom Context** (`internal/ctx/ctx.go`): wraps both `context.Context` and `*gin.Context`. All handler and service methods receive `*ctx.Context`. Provides `ShouldBind` (with i18n validation error translation), `ToSuccess`/`ToError` (JSON responses with uniform `{code, message, data/details}` envelope), and user/device extraction helpers.

**Exception system** (`internal/exception/`): domain errors carry HTTP status + business code + i18n messages. Defined as package-level vars via `ExceptionScope.New()`. Error codes are scoped: Common=1000, User=20000, UserRole=21000. Service returns `(*Result, *exception.Exception)`; handler converts exceptions to JSON via `ctx.ToError()`.

**i18n** (`internal/i18n/`): `Text` struct holds `En`/`Zh` fields. Exceptions carry `I18nMsg`; locale is extracted from `locale` header or `Accept-Language` via middleware, stored in gin context keys.

### Middleware Stack (order matters)

1. `gin.Recovery()` — panic recovery
2. `ZapLogger` — request logging (redacts Authorization/Cookie headers)
3. `ZapRecovery` — panic recovery with stack traces in dev mode
4. `CORS` — dev mode only
5. `Translations` — locale extraction, sets translator + locale in context
6. `RateLimit(100)` — per-IP, 100 req/min via Redis GCRA algorithm

Route-group middleware: `jwt.JWT()` (token validation + auto-refresh when < 1/3 remaining), `auth.RoleCheckAny()` / `auth.RoleCheckAll()` for RBAC.

### Database

- MySQL via GORM; database is auto-created (`CREATE DATABASE IF NOT EXISTS`) on connect.
- AutoMigrate runs on startup for models listed in `app.go` (`UserRole`, `User`). Add new models there to auto-migrate.
- Seed data runs after migration: inserts 6 default user roles (idempotent — skips existing).
- No migration versioning; for production schema changes, consider introducing goose or golang-migrate.

### Config Loading

`internal/config/config.go`: loads `config.yml` (base), then `config.{mode}.yml` (merged), then env vars with `APP_` prefix. Key separator in env vars is `_` (e.g. `APP_DATABASE_PASSWORD` maps to `database.password`). Defaults come from `DefaultConfig` struct in `default.go`.

Available modes: `dev`, `prod`, `test` — set via `-mode` flag or `APP_MODE` env var.

### Adding a New Entity

1. Define model in `internal/model/`
2. (Optional) Create repo in `internal/repo/` if custom queries needed
3. (Optional) Create DTOs in `internal/dto/`
4. Create service in `internal/service/`
5. Create handler in `internal/handler/`
6. Add to `app.go` AutoMigrate call
7. Register routes in `internal/router/`
8. Wire into `Handler`/`Service`/`Repo` aggregation interfaces

The `generate.sh` script scaffolds model/repo/dto/service/handler boilerplate for a given entity name.

### Key Dependencies

| Purpose | Library |
|---------|---------|
| HTTP framework | Gin v1.11 |
| ORM | GORM + MySQL driver |
| Redis | go-redis/v9 |
| JWT | golang-jwt/v5 |
| Logging | Zap + Lumberjack (rotation) |
| Config | Viper (env prefix: `APP_`) |
| Task queue | Asynq |
| ID generation | bwmarrin/snowflake |
| Validation | go-playground/validator/v10 |
| Dev hot reload | Air (`.air.toml`) |

## Notes

- **Secrets**: Never commit secrets. `configs/config.{dev,test,prod}.yml` and `.env` are gitignored. Use `config.dev.yml.example` as template. Env vars with `APP_` prefix override config values in production.
- **Verification codes**: The `verifyCode()` function in `auth_service.go` currently only checks non-empty. A TODO marks where real SMS/email verification should be integrated.
- **Test mocks**: `internal/service/*_test.go` uses stub implementations that embed the real repo interfaces and override only the methods needed. Follow this pattern for new service tests.
