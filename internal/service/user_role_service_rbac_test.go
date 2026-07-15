package service

import (
	"context"
	"testing"

	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"

	"go.uber.org/zap"
)

type builtInRoleRepoStub struct {
	repo.UserRoleRepo
	role *model.UserRole
}

func (r *builtInRoleRepoStub) GetOne(context.Context, ...repo.QueryOption) (*model.UserRole, error) {
	return r.role, nil
}

func TestBuiltInRoleCannotBeUpdatedOrDeleted(t *testing.T) {
	roleRepo := &builtInRoleRepoStub{role: &model.UserRole{Model: model.Model{ID: 1}, TenantID: 0, Code: "admin", BuiltIn: true, Enabled: true}}
	harness := &permissionRepoHarness{role: roleRepo}
	service := NewUserRoleService(harness, nil, nil, zap.NewNop())
	ctx := cctx.WithTenant(context.Background(), 42)

	name := "Changed"
	if _, exc := service.Update(ctx, 1, dto.UserRoleUpdateReqDto{Name: &name}); exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected built-in update to be forbidden, got %v", exc)
	}
	if exc := service.Delete(ctx, 1); exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected built-in delete to be forbidden, got %v", exc)
	}
}
