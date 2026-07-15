package service

import (
	"context"
	"testing"

	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type permissionUserRepoStub struct {
	repo.UserRepo
	user     *model.User
	isMember bool
}

func (r *permissionUserRepoStub) GetByUniCode(context.Context, string) (*model.User, error) {
	return r.user, nil
}

func (r *permissionUserRepoStub) HasTenantMembership(context.Context, uint64, uint64) (bool, error) {
	return r.isMember, nil
}

type permissionTenantRepoStub struct {
	repo.TenantRepo
	tenant *model.Tenant
}

func (r *permissionTenantRepoStub) GetByID(context.Context, uint64, ...repo.QueryOption) (*model.Tenant, error) {
	return r.tenant, nil
}

type permissionRoleRepoStub struct {
	repo.UserRoleRepo
	codes           []string
	queriedTenantID uint64
}

func (r *permissionRoleRepoStub) GetPermissionCodesByUserAndTenant(_ context.Context, _, tenantID uint64) ([]string, error) {
	r.queriedTenantID = tenantID
	return r.codes, nil
}

type permissionRepoHarness struct {
	user   repo.UserRepo
	role   repo.UserRoleRepo
	tenant repo.TenantRepo
}

func (r *permissionRepoHarness) DB() *gorm.DB                                            { return nil }
func (r *permissionRepoHarness) Logger() *zap.Logger                                     { return zap.NewNop() }
func (r *permissionRepoHarness) Transaction(context.Context, func(*gorm.DB) error) error { return nil }
func (r *permissionRepoHarness) User() repo.UserRepo                                     { return r.user }
func (r *permissionRepoHarness) UserRole() repo.UserRoleRepo                             { return r.role }
func (r *permissionRepoHarness) Permission() repo.PermissionRepo                         { return nil }
func (r *permissionRepoHarness) DeadLetter() repo.DeadLetterRepo                         { return nil }
func (r *permissionRepoHarness) Tenant() repo.TenantRepo                                 { return r.tenant }

func TestPermissionCodesAreTenantScoped(t *testing.T) {
	roleRepo := &permissionRoleRepoStub{codes: []string{"user.read"}}
	harness := &permissionRepoHarness{
		user:   &permissionUserRepoStub{user: &model.User{Model: model.Model{ID: 7}, TenantID: 42, Active: true}},
		role:   roleRepo,
		tenant: &permissionTenantRepoStub{tenant: &model.Tenant{Active: true}},
	}
	service := NewPermissionService(harness, nil, zap.NewNop())

	codes, exc := service.GetPermissionCodesByUniCode(cctx.WithTenant(context.Background(), 42), "U-1")
	if exc != nil {
		t.Fatalf("unexpected exception: %v", exc)
	}
	if len(codes) != 1 || codes[0] != "user.read" {
		t.Fatalf("unexpected permissions: %v", codes)
	}
	if roleRepo.queriedTenantID != 42 {
		t.Fatalf("queried tenant = %d, want 42", roleRepo.queriedTenantID)
	}
}

func TestPermissionCodesRejectNonMemberTenant(t *testing.T) {
	harness := &permissionRepoHarness{
		user:   &permissionUserRepoStub{user: &model.User{Model: model.Model{ID: 7}, TenantID: 1, Active: true}, isMember: false},
		role:   &permissionRoleRepoStub{},
		tenant: &permissionTenantRepoStub{tenant: &model.Tenant{Active: true}},
	}
	service := NewPermissionService(harness, nil, zap.NewNop())

	if _, exc := service.GetPermissionCodesByUniCode(cctx.WithTenant(context.Background(), 42), "U-1"); exc == nil {
		t.Fatal("expected non-member tenant to be rejected")
	}
}
