package service

import (
	"context"
	"go-server-starter/internal/config"
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
}

func (s *stubUserRepo) GetOne(_ context.Context, _ ...repo.QueryOption) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) WithTx(_ *gorm.DB) repo.UserRepo { return s }

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
