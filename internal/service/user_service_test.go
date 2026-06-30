package service

import (
	"context"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// testUserRepo is a stub that returns canned data for user service tests.
type testUserRepo struct {
	repo.UserRepo
	userByID        *model.User
	userByIDErr     error
	userByUniCode    *model.User
	userByUniCodeErr error
}

func (r *testUserRepo) WithTx(_ *gorm.DB) repo.UserRepo { return r }

func (r *testUserRepo) GetByID(_ context.Context, _ uint64, _ ...repo.QueryOption) (*model.User, error) {
	return r.userByID, r.userByIDErr
}

func (r *testUserRepo) GetByUniCode(_ context.Context, _ string) (*model.User, error) {
	return r.userByUniCode, r.userByUniCodeErr
}

func (r *testUserRepo) GetOne(_ context.Context, _ ...repo.QueryOption) (*model.User, error) {
	return r.userByUniCode, r.userByUniCodeErr
}

func (r *testUserRepo) GetByIDs(_ context.Context, _ []uint64, _ ...repo.QueryOption) ([]*model.User, error) {
	return nil, nil
}

func (r *testUserRepo) GetTable(_ context.Context, _, _ int, _ ...repo.QueryOption) ([]*model.User, int64, error) {
	return nil, 0, nil
}

// testRepo wraps test repos into a full repo.Repo.
type testRepo struct {
	userRepo     repo.UserRepo
	userRoleRepo repo.UserRoleRepo
}

func (r *testRepo) DB() *gorm.DB                             { return nil }
func (r *testRepo) Logger() *zap.Logger                      { return zap.NewNop() }
func (r *testRepo) Transaction(_ context.Context, _ func(tx *gorm.DB) error) error { return nil }
func (r *testRepo) User() repo.UserRepo                      { return r.userRepo }
func (r *testRepo) UserRole() repo.UserRoleRepo              { return r.userRoleRepo }
func (r *testRepo) DeadLetter() repo.DeadLetterRepo          { return nil }
func (r *testRepo) Tenant() repo.TenantRepo                   { return nil }

func sampleUser() *model.User {
	return &model.User{
		UniCode:     "U-001",
		Email:       "test@example.com",
		Mobile:      "13800138000",
		CountryCode: "86",
		Nickname:    "TestUser",
		Roles: []model.UserRole{
			{Code: enum.RoleCodeUser, Enabled: true},
		},
	}
}

func newTestUserService(userRepo repo.UserRepo) UserService {
	return NewUserService(
		&testRepo{userRepo: userRepo, userRoleRepo: &stubUserRoleRepo{}},
		nil, // redis not needed for these tests
		zap.NewNop(),
	)
}

func TestGetByID_Found(t *testing.T) {
	expected := sampleUser()
	svc := newTestUserService(&testUserRepo{userByID: expected})

	user, exc := svc.GetByID(context.Background(), 1)
	if exc != nil {
		t.Fatalf("unexpected error: %v", exc)
	}
	if user.UniCode != expected.UniCode {
		t.Errorf("expected UniCode %s, got %s", expected.UniCode, user.UniCode)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newTestUserService(&testUserRepo{userByID: nil, userByIDErr: gorm.ErrRecordNotFound})

	_, exc := svc.GetByID(context.Background(), 999)
	if exc == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if exc.Code != exception.UserNotFound.Code {
		t.Errorf("expected code %d, got %d", exception.UserNotFound.Code, exc.Code)
	}
}

func TestGetByUniCode_Found(t *testing.T) {
	expected := sampleUser()
	svc := newTestUserService(&testUserRepo{userByUniCode: expected})

	user, exc := svc.GetByUniCode(context.Background(), "U-001")
	if exc != nil {
		t.Fatalf("unexpected error: %v", exc)
	}
	if user.UniCode != expected.UniCode {
		t.Errorf("expected UniCode %s, got %s", expected.UniCode, user.UniCode)
	}
}

func TestGetByUniCode_NotFound(t *testing.T) {
	svc := newTestUserService(&testUserRepo{userByUniCode: nil, userByUniCodeErr: gorm.ErrRecordNotFound})

	_, exc := svc.GetByUniCode(context.Background(), "nonexistent")
	if exc == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if exc.Code != exception.UserNotFound.Code {
		t.Errorf("expected code %d, got %d", exception.UserNotFound.Code, exc.Code)
	}
}

func TestGetInfoByUniCode_Success(t *testing.T) {
	expected := sampleUser()
	svc := newTestUserService(&testUserRepo{userByUniCode: expected})

	info, exc := svc.GetInfoByUniCode(context.Background(), "U-001")
	if exc != nil {
		t.Fatalf("unexpected error: %v", exc)
	}
	if info.UniCode != expected.UniCode {
		t.Errorf("expected UniCode %s, got %s", expected.UniCode, info.UniCode)
	}
	if len(info.Roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(info.Roles))
	}
	if info.Roles[0] != "user" {
		t.Errorf("expected role 'user', got '%s'", info.Roles[0])
	}
}

func TestGetInfoByUniCode_NotFound(t *testing.T) {
	svc := newTestUserService(&testUserRepo{userByUniCode: nil, userByUniCodeErr: gorm.ErrRecordNotFound})

	_, exc := svc.GetInfoByUniCode(context.Background(), "nonexistent")
	if exc == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPaginationReqDto_Defaults(t *testing.T) {
	params := dto.UserTableQueryReqDto{}
	if params.Page != 0 {
		t.Error("expected default Page to be 0")
	}
	if params.PageSize != 0 {
		t.Error("expected default PageSize to be 0")
	}
}
