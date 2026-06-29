package service

import (
	"context"
	"go-server-starter/internal/config"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/jwt"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mockRepo implements repo.Repo for testing.
type mockRepo struct {
	userRepo     repo.UserRepo
	userRoleRepo repo.UserRoleRepo
}

func (m *mockRepo) DB() *gorm.DB                             { return nil }
func (m *mockRepo) Logger() *zap.Logger                      { return zap.NewNop() }
func (m *mockRepo) Transaction(_ context.Context, _ func(tx *gorm.DB) error) error { return nil }
func (m *mockRepo) User() repo.UserRepo                      { return m.userRepo }
func (m *mockRepo) UserRole() repo.UserRoleRepo              { return m.userRoleRepo }

// stubUserRepo is a minimal stub — only methods actually called are implemented.
type stubUserRepo struct {
	repo.UserRepo
}

func (s *stubUserRepo) GetOne(_ context.Context, _ ...repo.QueryOption) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (s *stubUserRepo) WithTx(_ *gorm.DB) repo.UserRepo { return s }

// stubUserRoleRepo is a minimal stub.
type stubUserRoleRepo struct {
	repo.UserRoleRepo
}

func (s *stubUserRoleRepo) WithTx(_ *gorm.DB) repo.UserRoleRepo { return s }

// testJWT creates a minimal JWT instance for testing.
func testJWT() *jwt.JWT {
	return jwt.NewJWT(&config.JWTConfig{
		Issuer:      "test",
		TokenSecret: "test-secret",
	}, zap.NewNop())
}

func TestVerifyCode(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		invalidErr *exception.Exception
		wantErr    bool
	}{
		{"valid code", "123456", exception.UserMobileVerificationCodeIsIncorrect, false},
		{"empty code", "", exception.UserMobileVerificationCodeIsIncorrect, true},
		{"valid code email", "888888", exception.UserEmailVerificationCodeIsIncorrect, false},
		{"empty code email", "", exception.UserEmailVerificationCodeIsIncorrect, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyCode(tt.code, tt.invalidErr)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginByMobileAndCode_InvalidCode(t *testing.T) {
	svc := NewAuthService(
		&mockRepo{userRepo: &stubUserRepo{}, userRoleRepo: &stubUserRoleRepo{}},
		testJWT(),
		nil, // taskq client not needed for this test
		zap.NewNop(),
	)

	_, exc := svc.LoginByMobileAndCode(context.Background(), enum.DeviceTypeWeb, dto.AuthLoginByMobileAndCodeReqDto{
		Mobile:      "13800138000",
		CountryCode: "86",
		Code:        "",
	})

	if exc == nil {
		t.Fatal("expected error for empty code, got nil")
	}
	if exc.Code != exception.UserMobileVerificationCodeIsIncorrect.Code {
		t.Errorf("expected code %d, got %d", exception.UserMobileVerificationCodeIsIncorrect.Code, exc.Code)
	}
}

func TestLoginByEmailAndCode_InvalidCode(t *testing.T) {
	svc := NewAuthService(
		&mockRepo{userRepo: &stubUserRepo{}, userRoleRepo: &stubUserRoleRepo{}},
		testJWT(),
		nil, // taskq client not needed for this test
		zap.NewNop(),
	)

	_, exc := svc.LoginByEmailAndCode(context.Background(), enum.DeviceTypeWeb, dto.AuthLoginByEmailAndCodeReqDto{
		Email: "test@example.com",
		Code:  "",
	})

	if exc == nil {
		t.Fatal("expected error for empty code, got nil")
	}
	if exc.Code != exception.UserEmailVerificationCodeIsIncorrect.Code {
		t.Errorf("expected code %d, got %d", exception.UserEmailVerificationCodeIsIncorrect.Code, exc.Code)
	}
}
