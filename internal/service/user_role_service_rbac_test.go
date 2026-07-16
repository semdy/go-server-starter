package service

import (
	"context"
	"testing"

	"go-server-starter/internal/constant"
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

type tenantPermissionRepoStub struct {
	repo.PermissionRepo
	permissions []*model.Permission
}

func (r *tenantPermissionRepoStub) GetByIDs(context.Context, []uint64, ...repo.QueryOption) ([]*model.Permission, error) {
	return r.permissions, nil
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

func TestTenantManagementPermissionsCannotBeAssignedToCustomRole(t *testing.T) {
	permissionRepo := &tenantPermissionRepoStub{permissions: []*model.Permission{
		{Model: model.Model{ID: 1}, Code: constant.PermissionTenantRead, Enabled: true},
	}}
	harness := &permissionRepoHarness{permission: permissionRepo}
	service := &UserRoleServiceImpl{repo: harness, logger: zap.NewNop()}

	exc := service.validatePermissionIDs(context.Background(), []uint64{1})
	if exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected tenant management permission assignment to be forbidden, got %v", exc)
	}
}

func TestBuiltInRoleCodeCannotBeUsedByCustomRole(t *testing.T) {
	service := &UserRoleServiceImpl{logger: zap.NewNop()}
	_, exc := service.Create(context.Background(), dto.UserRoleCreateReqDto{
		Code: "super_admin",
		Name: "Fake super admin",
	})
	if exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected built-in role code to be reserved, got %v", exc)
	}
}

func TestCustomRoleUsingReservedCodeCannotSatisfyRoleCheck(t *testing.T) {
	codes := effectiveRoleCodes([]*model.UserRole{
		{TenantID: 42, Code: "super_admin", BuiltIn: false, Enabled: true},
		{TenantID: 0, Code: "admin", BuiltIn: true, Enabled: true},
		{TenantID: 42, Code: "editor", BuiltIn: false, Enabled: true},
	})
	if len(codes) != 2 || codes[0] != "admin" || codes[1] != "editor" {
		t.Fatalf("unexpected effective role codes: %v", codes)
	}
}
