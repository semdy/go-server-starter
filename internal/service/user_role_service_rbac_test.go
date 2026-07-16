package service

import (
	"context"
	"testing"

	"go-server-starter/internal/constant"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type builtInRoleRepoStub struct {
	repo.UserRoleRepo
	role *model.UserRole
}

type tenantPermissionRepoStub struct {
	repo.PermissionRepo
	permissions []*model.Permission
}

func (r *tenantPermissionRepoStub) GetByIDs(_ context.Context, ids []uint64, _ ...repo.QueryOption) ([]*model.Permission, error) {
	result := make([]*model.Permission, 0, len(ids))
	for _, id := range ids {
		for _, permission := range r.permissions {
			if permission.ID == id {
				result = append(result, permission)
				break
			}
		}
	}
	return result, nil
}

func (r *tenantPermissionRepoStub) GetMany(context.Context, ...repo.QueryOption) ([]*model.Permission, error) {
	return r.permissions, nil
}

func (r *tenantPermissionRepoStub) GetByID(_ context.Context, id uint64, _ ...repo.QueryOption) (*model.Permission, error) {
	for _, permission := range r.permissions {
		if permission.ID == id {
			return permission, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type permissionAccessStub struct {
	PermissionService
	codes []string
}

type toggleRoleRepoStub struct {
	builtInRoleRepoStub
	actorRoles []*model.UserRole
}

func (r *toggleRoleRepoStub) GetRolesByUserAndTenant(context.Context, uint64, uint64) ([]*model.UserRole, error) {
	return r.actorRoles, nil
}

func (r *toggleRoleRepoStub) SetRolePermission(_ context.Context, _, permissionID uint64, checked bool) error {
	if checked {
		r.role.Permissions = []model.Permission{{Model: model.Model{ID: permissionID}, Code: constant.PermissionUserDelete, Enabled: true}}
	} else {
		r.role.Permissions = nil
	}
	return nil
}

func (s *permissionAccessStub) GetCachedPermissionCodesByUniCode(context.Context, string) ([]string, *exception.Exception) {
	return s.codes, nil
}

func (s *permissionAccessStub) InvalidateTenantAccessCaches(context.Context, uint64) {}

func (s *permissionAccessStub) InvalidateGlobalAccessCaches(context.Context) {}

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

	exc := service.validatePermissionIDs(context.Background(), []uint64{1}, nil)
	if exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected tenant management permission assignment to be forbidden, got %v", exc)
	}
}

func TestGetPermissionConfigReturnsGlobalPermissionsWithSwitchState(t *testing.T) {
	roleRepo := &builtInRoleRepoStub{role: &model.UserRole{
		Model: model.Model{ID: 9}, TenantID: 42, Code: "editor", Name: "Editor", Enabled: true,
		Permissions: []model.Permission{{Model: model.Model{ID: 1}, Code: constant.PermissionUserRead, Enabled: true}},
	}}
	permissionRepo := &tenantPermissionRepoStub{permissions: []*model.Permission{
		{Model: model.Model{ID: 1}, Code: constant.PermissionUserRead, Enabled: true},
		{Model: model.Model{ID: 2}, Code: constant.PermissionUserDelete, Enabled: true},
		{Model: model.Model{ID: 3}, Code: constant.PermissionTenantRead, Enabled: true},
	}}
	access := &permissionAccessStub{codes: []string{
		constant.PermissionRoleAssignPermissions,
		constant.PermissionUserRead,
		constant.PermissionUserDelete,
		constant.PermissionTenantRead,
	}}
	harness := &permissionRepoHarness{role: roleRepo, permission: permissionRepo}
	service := &UserRoleServiceImpl{repo: harness, access: access, logger: zap.NewNop()}
	ctx := cctx.WithUserUniCode(cctx.WithTenant(context.Background(), 42), "U-1")

	config, exc := service.GetPermissionConfig(ctx, 9)
	if exc != nil {
		t.Fatalf("unexpected exception: %v", exc)
	}
	if !config.Editable || len(config.Permissions) != 3 {
		t.Fatalf("unexpected config: %+v", config)
	}
	if !config.Permissions[0].Checked || !config.Permissions[0].Editable {
		t.Fatalf("assigned permission should be checked and editable: %+v", config.Permissions[0])
	}
	if config.Permissions[1].Checked || !config.Permissions[1].Editable {
		t.Fatalf("grantable permission should be unchecked and editable: %+v", config.Permissions[1])
	}
	if config.Permissions[2].Checked || config.Permissions[2].Editable {
		t.Fatalf("tenant permission must not be editable: %+v", config.Permissions[2])
	}
}

func TestTogglePermissionUpdatesSingleRolePermission(t *testing.T) {
	role := &model.UserRole{Model: model.Model{ID: 9}, TenantID: 0, Code: "admin", Name: "Admin", BuiltIn: true, Enabled: true}
	roleRepo := &toggleRoleRepoStub{
		builtInRoleRepoStub: builtInRoleRepoStub{role: role},
		actorRoles:          []*model.UserRole{{TenantID: 0, Code: "super_admin", BuiltIn: true, Enabled: true}},
	}
	permissionRepo := &tenantPermissionRepoStub{permissions: []*model.Permission{
		{Model: model.Model{ID: 2}, Code: constant.PermissionUserDelete, Enabled: true},
	}}
	access := &permissionAccessStub{codes: []string{
		constant.PermissionRoleAssignPermissions,
		constant.PermissionUserDelete,
	}}
	harness := &permissionRepoHarness{
		user:       &permissionUserRepoStub{user: &model.User{Model: model.Model{ID: 7}, TenantID: 42, Active: true}},
		role:       roleRepo,
		permission: permissionRepo,
		tenant:     &permissionTenantRepoStub{tenant: &model.Tenant{Active: true}},
	}
	service := &UserRoleServiceImpl{repo: harness, access: access, logger: zap.NewNop()}
	ctx := cctx.WithUserUniCode(cctx.WithTenant(context.Background(), 42), "U-1")

	config, exc := service.TogglePermission(ctx, 9, 2, true)
	if exc != nil {
		t.Fatalf("unexpected exception enabling permission: %v", exc)
	}
	if !config.Editable || len(config.Permissions) != 1 || !config.Permissions[0].Checked {
		t.Fatalf("permission should be checked after enabling: %+v", config.Permissions)
	}

	config, exc = service.TogglePermission(ctx, 9, 2, false)
	if exc != nil {
		t.Fatalf("unexpected exception disabling permission: %v", exc)
	}
	if config.Permissions[0].Checked {
		t.Fatalf("permission should be unchecked after disabling: %+v", config.Permissions[0])
	}
}

func TestNonSuperAdminCannotModifyBuiltInRolePermissions(t *testing.T) {
	role := &model.UserRole{Model: model.Model{ID: 9}, TenantID: 0, Code: "admin", Name: "Admin", BuiltIn: true, Enabled: true}
	roleRepo := &toggleRoleRepoStub{
		builtInRoleRepoStub: builtInRoleRepoStub{role: role},
		actorRoles:          []*model.UserRole{{TenantID: 0, Code: "admin", BuiltIn: true, Enabled: true}},
	}
	harness := &permissionRepoHarness{
		user:   &permissionUserRepoStub{user: &model.User{Model: model.Model{ID: 7}, TenantID: 42, Active: true}},
		role:   roleRepo,
		tenant: &permissionTenantRepoStub{tenant: &model.Tenant{Active: true}},
	}
	service := &UserRoleServiceImpl{repo: harness, logger: zap.NewNop()}
	ctx := cctx.WithUserUniCode(cctx.WithTenant(context.Background(), 42), "U-1")

	if _, exc := service.TogglePermission(ctx, 9, 2, true); exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected non-super-admin built-in role update to be forbidden, got %v", exc)
	}
}

func TestSuperAdminPermissionsCannotBeToggled(t *testing.T) {
	roleRepo := &builtInRoleRepoStub{role: &model.UserRole{
		Model: model.Model{ID: 1}, TenantID: 0, Code: "super_admin", Name: "Super Admin", BuiltIn: true, Enabled: true,
	}}
	harness := &permissionRepoHarness{role: roleRepo}
	service := &UserRoleServiceImpl{repo: harness, logger: zap.NewNop()}

	if _, exc := service.TogglePermission(cctx.WithTenant(context.Background(), 42), 1, 2, true); exc == nil || exc.StatusCode != 403 {
		t.Fatalf("expected super_admin permission toggle to be forbidden, got %v", exc)
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
