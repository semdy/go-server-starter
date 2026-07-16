package service

import (
	"context"
	"go-server-starter/internal/config"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/jwt"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type mockRepo struct {
	userRepo     repo.UserRepo
	userRoleRepo repo.UserRoleRepo
}

func (m *mockRepo) DB() *gorm.DB                                                   { return nil }
func (m *mockRepo) Logger() *zap.Logger                                            { return zap.NewNop() }
func (m *mockRepo) Transaction(_ context.Context, _ func(tx *gorm.DB) error) error { return nil }
func (m *mockRepo) User() repo.UserRepo                                            { return m.userRepo }
func (m *mockRepo) UserRole() repo.UserRoleRepo                                    { return m.userRoleRepo }
func (m *mockRepo) Permission() repo.PermissionRepo                                { return nil }
func (m *mockRepo) DeadLetter() repo.DeadLetterRepo                                { return nil }
func (m *mockRepo) Tenant() repo.TenantRepo                                        { return nil }

type stubUserRepo struct {
	repo.UserRepo
	user    *model.User
	tenants []*model.Tenant
}

func (s *stubUserRepo) GetOne(_ context.Context, _ ...repo.QueryOption) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) WithTx(_ *gorm.DB) repo.UserRepo { return s }

func (s *stubUserRepo) GetByUniCode(context.Context, string) (*model.User, error) {
	return s.user, nil
}

func (s *stubUserRepo) GetTenantsByUserID(context.Context, uint64) ([]*model.Tenant, error) {
	return s.tenants, nil
}

type stubUserRoleRepo struct {
	repo.UserRoleRepo
}

func (s *stubUserRoleRepo) WithTx(_ *gorm.DB) repo.UserRoleRepo { return s }

func testJWT() *jwt.JWT {
	return jwt.NewJWT(&config.JWTConfig{
		Issuer:      "test",
		TokenSecret: "test-secret",
	}, zap.NewNop())
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(
		&mockRepo{userRepo: &stubUserRepo{}, userRoleRepo: &stubUserRoleRepo{}},
		testJWT(),
		nil, // access
		nil, // taskq
		zap.NewNop(),
	)
	if svc == nil {
		t.Fatal("expected non-nil auth service")
	}
}

func TestGetMyTenantsReturnsCurrentAndAvailableTenants(t *testing.T) {
	userRepo := &stubUserRepo{
		user: &model.User{Model: model.Model{ID: 7}, UniCode: "U-1", Active: true},
		tenants: []*model.Tenant{
			{Model: model.Model{ID: 42}, Code: "default", Name: "Default", Active: true},
			{Model: model.Model{ID: 99}, Code: "team", Name: "Team", Active: true},
		},
	}
	svc := NewAuthService(&mockRepo{userRepo: userRepo}, testJWT(), nil, nil, zap.NewNop())
	ctx := cctx.WithTenant(context.Background(), 99)

	result, exc := svc.GetMyTenants(ctx, "U-1")
	if exc != nil {
		t.Fatalf("unexpected exception: %v", exc)
	}
	if result.CurrentTenantID != 99 || len(result.Tenants) != 2 {
		t.Fatalf("unexpected tenant response: %+v", result)
	}
	if result.Tenants[0].ID != 42 || result.Tenants[1].ID != 99 {
		t.Fatalf("unexpected tenant list: %+v", result.Tenants)
	}
}
