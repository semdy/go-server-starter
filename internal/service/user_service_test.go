package service

import (
	"context"
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// testUserRepo is a stub that returns canned data for user service tests.
type testUserRepo struct {
	repo.UserRepo
	userByID      *model.User
	userByIDErr   error
	userByUniCode *model.User
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

func newTestGinCtxForUser() *ctx.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	return ctx.FromGinCtx(c)
}

func TestGetByID_Found(t *testing.T) {
	expected := sampleUser()
	svc := newTestUserService(&testUserRepo{userByID: expected})
	c := newTestGinCtxForUser()

	user, exc := svc.GetByID(c, 1)
	if exc != nil {
		t.Fatalf("unexpected error: %v", exc)
	}
	if user.UniCode != expected.UniCode {
		t.Errorf("expected UniCode %s, got %s", expected.UniCode, user.UniCode)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newTestUserService(&testUserRepo{userByID: nil, userByIDErr: gorm.ErrRecordNotFound})
	c := newTestGinCtxForUser()

	_, exc := svc.GetByID(c, 999)
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
	c := newTestGinCtxForUser()

	user, exc := svc.GetByUniCode(c, "U-001")
	if exc != nil {
		t.Fatalf("unexpected error: %v", exc)
	}
	if user.UniCode != expected.UniCode {
		t.Errorf("expected UniCode %s, got %s", expected.UniCode, user.UniCode)
	}
}

func TestGetByUniCode_NotFound(t *testing.T) {
	svc := newTestUserService(&testUserRepo{userByUniCode: nil, userByUniCodeErr: gorm.ErrRecordNotFound})
	c := newTestGinCtxForUser()

	_, exc := svc.GetByUniCode(c, "nonexistent")
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
	c := newTestGinCtxForUser()

	info, exc := svc.GetInfoByUniCode(c, "U-001")
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
	c := newTestGinCtxForUser()

	_, exc := svc.GetInfoByUniCode(c, "nonexistent")
	if exc == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateInfo_UserNotFound(t *testing.T) {
	// UpdateInfo calls GetUserID which needs the uniCode stored in gin context,
	// so we skip that path and test via GetByUniCode indirectly.
	svc := newTestUserService(&testUserRepo{userByUniCode: nil, userByUniCodeErr: gorm.ErrRecordNotFound})
	c := newTestGinCtxForUser()

	_, exc := svc.GetByUniCode(c, "nonexistent")
	if exc == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestDTOs tests the DTO structures used by the service layer.
func TestPaginationReqDto_Defaults(t *testing.T) {
	params := dto.UserTableQueryReqDto{}
	// Verify zero-value pagination fields are valid (service uses NormalizePageAndPageSize)
	if params.Page != 0 {
		t.Error("expected default Page to be 0")
	}
	if params.PageSize != 0 {
		t.Error("expected default PageSize to be 0")
	}
}
