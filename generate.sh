#!/bin/bash

# Usage:
#   ./generate.sh <module_name>          - Create module files + auto-register
#   ./generate.sh -d <module_name>       - Delete module files
# Example:
#   ./generate.sh product
#   ./generate.sh -d product

DELETE_MODE=false
MODULE_NAME=""

if [ "$1" = "-d" ]; then
    DELETE_MODE=true
    MODULE_NAME=$2
else
    MODULE_NAME=$1
fi

if [ -z "$MODULE_NAME" ]; then
    echo "Usage:"
    echo "  ./generate.sh <module_name>          - Create module files + auto-register"
    echo "  ./generate.sh -d <module_name>       - Delete module files"
    exit 1
fi

# Convert: product-category → product_category → ProductCategory
MODULE_NAME_SNAKE=$(echo "$MODULE_NAME" | sed 's/-/_/g' | sed 's/\([A-Z]\)/_\1/g' | sed 's/^_//' | tr '[:upper:]' '[:lower:]')
MODULE_NAME_LOWER=$MODULE_NAME_SNAKE
MODULE_NAME_UPPER=$(echo "$MODULE_NAME_SNAKE" | awk -F_ '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1' OFS="")

# File paths
MODEL_FILE="internal/model/${MODULE_NAME_LOWER}.go"
REPO_FILE="internal/repo/${MODULE_NAME_LOWER}_repo.go"
DTO_FILE="internal/dto/${MODULE_NAME_LOWER}_dto.go"
SERVICE_FILE="internal/service/${MODULE_NAME_LOWER}_service.go"
HANDLER_FILE="internal/handler/${MODULE_NAME_LOWER}_handler.go"
EXCEPTION_FILE="internal/exception/${MODULE_NAME_LOWER}_exception.go"
ROUTER_FILE="internal/router/${MODULE_NAME_LOWER}_router.go"
MIG_NUM=$(printf "%05d" $(($(ls internal/database/migration/migrations/*.sql 2>/dev/null | wc -l | tr -d ' ') + 1)))
MIGRATION_UP="internal/database/migration/migrations/${MIG_NUM}_create_${MODULE_NAME_LOWER}.sql"

# ============================================
# DELETE MODE
# ============================================
if [ "$DELETE_MODE" = true ]; then
    echo "[delete] Deleting module: $MODULE_NAME"
    for f in "$MODEL_FILE" "$REPO_FILE" "$DTO_FILE" "$SERVICE_FILE" "$HANDLER_FILE" "$EXCEPTION_FILE" "$ROUTER_FILE"; do
        [ -f "$f" ] && rm "$f" && echo "[ok] Deleted: $f"
    done
    echo ""
    echo "Note: manually remove registrations from repo.go, service.go, handler.go, router.go"
    exit 0
fi

# ============================================
# CREATE MODE
# ============================================
echo "🚀 Generating module: $MODULE_NAME ($MODULE_NAME_UPPER)"

mkdir -p internal/{model,repo,dto,service,handler,exception,router}
mkdir -p internal/database/migration/migrations

# ---- Model ----
if [ ! -f "$MODEL_FILE" ]; then
    cat > "$MODEL_FILE" << EOF
package model

type $MODULE_NAME_UPPER struct {
	Model
	// TODO: add fields
}
EOF
    echo "[ok] Created: $MODEL_FILE"
fi

# ---- Repo ----
if [ ! -f "$REPO_FILE" ]; then
    cat > "$REPO_FILE" << EOF
package repo

import (
	"go-server-starter/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ${MODULE_NAME_UPPER}Repo interface {
	BaseRepo[model.${MODULE_NAME_UPPER}]
}

type ${MODULE_NAME_LOWER}RepoImpl struct {
	BaseRepo[model.${MODULE_NAME_UPPER}]
}

func New${MODULE_NAME_UPPER}Repo(db *gorm.DB, logger *zap.Logger) ${MODULE_NAME_UPPER}Repo {
	return &${MODULE_NAME_LOWER}RepoImpl{
		BaseRepo: NewBaseRepo[model.${MODULE_NAME_UPPER}](db, logger),
	}
}
EOF
    echo "[ok] Created: $REPO_FILE"
fi

# ---- DTO ----
if [ ! -f "$DTO_FILE" ]; then
    cat > "$DTO_FILE" << EOF
package dto

// ${MODULE_NAME_UPPER}CreateReqDto TODO
type ${MODULE_NAME_UPPER}CreateReqDto struct{}

// ${MODULE_NAME_UPPER}UpdateReqDto TODO
type ${MODULE_NAME_UPPER}UpdateReqDto struct{}

// ${MODULE_NAME_UPPER}ResDto TODO
type ${MODULE_NAME_UPPER}ResDto struct {
	ID        uint64 \`json:"id"\`
	CreatedAt string \`json:"createdAt"\`
	UpdatedAt string \`json:"updatedAt"\`
}
EOF
    echo "[ok] Created: $DTO_FILE"
fi

# ---- Service ----
if [ ! -f "$SERVICE_FILE" ]; then
    cat > "$SERVICE_FILE" << EOF
package service

import (
	"context"
	"errors"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ${MODULE_NAME_UPPER}Service interface {
	GetByID(ctx context.Context, id uint64) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.PaginationReqDto) (*dto.PaginationResDto[[]*dto.${MODULE_NAME_UPPER}ResDto], *exception.Exception)
	Create(ctx context.Context, params dto.${MODULE_NAME_UPPER}CreateReqDto) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception)
	Update(ctx context.Context, id uint64, params dto.${MODULE_NAME_UPPER}UpdateReqDto) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception)
	Delete(ctx context.Context, id uint64) *exception.Exception
}

type ${MODULE_NAME_UPPER}ServiceImpl struct {
	repo   repo.Repo
	logger *zap.Logger
}

func New${MODULE_NAME_UPPER}Service(repo repo.Repo, logger *zap.Logger) ${MODULE_NAME_UPPER}Service {
	return &${MODULE_NAME_UPPER}ServiceImpl{repo: repo, logger: logger}
}

func to${MODULE_NAME_UPPER}ResDto(m *model.${MODULE_NAME_UPPER}) *dto.${MODULE_NAME_UPPER}ResDto {
	return &dto.${MODULE_NAME_UPPER}ResDto{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *${MODULE_NAME_UPPER}ServiceImpl) GetByID(ctx context.Context, id uint64) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception) {
	m, err := s.repo.${MODULE_NAME_UPPER}().GetByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if m == nil {
		return nil, exception.NotFound.Append("${MODULE_NAME_LOWER} not found")
	}
	return to${MODULE_NAME_UPPER}ResDto(m), nil
}

func (s *${MODULE_NAME_UPPER}ServiceImpl) GetTable(ctx context.Context, params dto.PaginationReqDto) (*dto.PaginationResDto[[]*dto.${MODULE_NAME_UPPER}ResDto], *exception.Exception) {
	entries, total, err := s.repo.${MODULE_NAME_UPPER}().GetTable(ctx, params.Page, params.PageSize, repo.Order("id ASC"))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	items := make([]*dto.${MODULE_NAME_UPPER}ResDto, 0, len(entries))
	for _, e := range entries {
		items = append(items, to${MODULE_NAME_UPPER}ResDto(e))
	}
	return utils.AssemblePaginationResDto(items, total, params.Page, params.PageSize), nil
}

func (s *${MODULE_NAME_UPPER}ServiceImpl) Create(ctx context.Context, params dto.${MODULE_NAME_UPPER}CreateReqDto) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception) {
	m := &model.${MODULE_NAME_UPPER}{}
	// TODO: set fields from params
	if err := s.repo.${MODULE_NAME_UPPER}().Create(ctx, m); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return to${MODULE_NAME_UPPER}ResDto(m), nil
}

func (s *${MODULE_NAME_UPPER}ServiceImpl) Update(ctx context.Context, id uint64, params dto.${MODULE_NAME_UPPER}UpdateReqDto) (*dto.${MODULE_NAME_UPPER}ResDto, *exception.Exception) {
	m, err := s.repo.${MODULE_NAME_UPPER}().GetByID(ctx, id)
	if err != nil || m == nil {
		return nil, exception.NotFound.Append("${MODULE_NAME_LOWER} not found")
	}
	// TODO: apply params to m
	if err := s.repo.${MODULE_NAME_UPPER}().UpdateByZeroFields(ctx, id, m); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return to${MODULE_NAME_UPPER}ResDto(m), nil
}

func (s *${MODULE_NAME_UPPER}ServiceImpl) Delete(ctx context.Context, id uint64) *exception.Exception {
	_ = s.repo.${MODULE_NAME_UPPER}().SoftDelete(ctx, id)
	return nil
}
EOF
    echo "[ok] Created: $SERVICE_FILE"
fi

# ---- Handler ----
if [ ! -f "$HANDLER_FILE" ]; then
    cat > "$HANDLER_FILE" << EOF
package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ${MODULE_NAME_UPPER}Handler interface {
	GetByID(c *gin.Context)
	GetTable(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type ${MODULE_NAME_UPPER}HandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func New${MODULE_NAME_UPPER}Handler(logger *zap.Logger, service service.Service) ${MODULE_NAME_UPPER}Handler {
	return &${MODULE_NAME_UPPER}HandlerImpl{logger: logger, service: service}
}

func (h *${MODULE_NAME_UPPER}HandlerImpl) GetByID(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.${MODULE_NAME_UPPER}().GetByID(appCtx.Ctx, id)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *${MODULE_NAME_UPPER}HandlerImpl) GetTable(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.PaginationReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.${MODULE_NAME_UPPER}().GetTable(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *${MODULE_NAME_UPPER}HandlerImpl) Create(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.${MODULE_NAME_UPPER}CreateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.${MODULE_NAME_UPPER}().Create(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *${MODULE_NAME_UPPER}HandlerImpl) Update(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	var params dto.${MODULE_NAME_UPPER}UpdateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.${MODULE_NAME_UPPER}().Update(appCtx.Ctx, id, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *${MODULE_NAME_UPPER}HandlerImpl) Delete(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.${MODULE_NAME_UPPER}().Delete(appCtx.Ctx, id); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}
EOF
    echo "[ok] Created: $HANDLER_FILE"
fi

# ---- Router ----
if [ ! -f "$ROUTER_FILE" ]; then
    cat > "$ROUTER_FILE" << EOF
package router

import "go-server-starter/internal/enum"

func (r *Router) Setup${MODULE_NAME_UPPER}Routes() {
	router := r.router.Group("/${MODULE_NAME_LOWER}")
	router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin))
	{
		router.GET("/:id", r.handler.${MODULE_NAME_UPPER}().GetByID)
		router.GET("", r.handler.${MODULE_NAME_UPPER}().GetTable)
		router.POST("", r.handler.${MODULE_NAME_UPPER}().Create)
		router.PUT("/:id", r.handler.${MODULE_NAME_UPPER}().Update)
		router.DELETE("/:id", r.handler.${MODULE_NAME_UPPER}().Delete)
	}
}
EOF
    echo "[ok] Created: $ROUTER_FILE"
fi

# ---- Migration ----
MIG_FILE=$(ls internal/database/migration/migrations/*create_${MODULE_NAME_LOWER}.sql 2>/dev/null | head -1)
if [ -z "$MIG_FILE" ]; then
    cat > "$MIGRATION_UP" << EOF
-- +goose Up
CREATE TABLE ${MODULE_NAME_LOWER} (
    id BIGINT UNSIGNED PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    version BIGINT UNSIGNED DEFAULT 0,
    -- TODO: add columns here
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS ${MODULE_NAME_LOWER};
EOF
    echo "[ok] Created: $MIGRATION_UP"
fi

# ---- Summary with registration instructions ----
echo ""
echo "✅ Module '$MODULE_NAME_UPPER' generated successfully!"
echo ""
echo "  New files:"
echo "    $MODEL_FILE"
echo "    $REPO_FILE"
echo "    $DTO_FILE"
echo "    $SERVICE_FILE"
echo "    $HANDLER_FILE"
echo "    $ROUTER_FILE"
echo "    $MIGRATION_UP"
echo ""
echo "  Register in the aggregation files (3 files, 1 line each):"
echo ""
echo "  repo.go:"
echo "    interface  →  ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Repo"
echo "    struct     →  ${MODULE_NAME_LOWER}Repo ${MODULE_NAME_UPPER}Repo"
echo "    NewRepo    →  ${MODULE_NAME_LOWER}Repo: New${MODULE_NAME_UPPER}Repo(db, logger),"
echo "    accessor   →  func (r *RepoImpl) ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Repo { return r.${MODULE_NAME_LOWER}Repo }"
echo ""
echo "  service.go:"
echo "    interface  →  ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Service"
echo "    struct     →  ${MODULE_NAME_LOWER}Service ${MODULE_NAME_UPPER}Service"
echo "    NewService →  ${MODULE_NAME_LOWER}Service: New${MODULE_NAME_UPPER}Service(repo, logger),"
echo "    accessor   →  func (s *ServiceImpl) ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Service { return s.${MODULE_NAME_LOWER}Service }"
echo ""
echo "  handler.go:"
echo "    interface  →  ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Handler"
echo "    struct     →  ${MODULE_NAME_LOWER}Handler ${MODULE_NAME_UPPER}Handler"
echo "    NewHandler →  ${MODULE_NAME_LOWER}Handler: New${MODULE_NAME_UPPER}Handler(logger, service),"
echo "    accessor   →  func (h *HandlerImpl) ${MODULE_NAME_UPPER}() ${MODULE_NAME_UPPER}Handler { return h.${MODULE_NAME_LOWER}Handler }"
echo ""
echo "  router.go:"
echo "    SetupRoutes →  r.Setup${MODULE_NAME_UPPER}Routes()"
echo ""
echo "  After registration:"
echo "    1. Fill model fields in $MODEL_FILE"
echo "    2. Complete migration SQL in $MIGRATION_UP"
echo "    3. Fill DTO fields in $DTO_FILE"
echo "    4. Implement business logic in $SERVICE_FILE"
echo "    5. Add swagger annotations to $HANDLER_FILE"
echo "    6. Restart to run migrations"
