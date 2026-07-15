package auth

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleCheckType string

const (
	RoleCheckTypeAll RoleCheckType = "all" // 所有角色都符合
	RoleCheckTypeAny RoleCheckType = "any" // 任意一个角色符合
)

type Auth interface {
	RoleCheck(roleCheckType RoleCheckType, roles ...enum.RoleCode) gin.HandlerFunc
	RoleCheckAny(roles ...enum.RoleCode) gin.HandlerFunc
	RoleCheckAll(roles ...enum.RoleCode) gin.HandlerFunc
	// String variants accept dynamic role names (e.g. tenant-specific roles).
	RoleCheckAnyS(roles ...string) gin.HandlerFunc
	RoleCheckAllS(roles ...string) gin.HandlerFunc
	PermissionCheckAny(permissions ...string) gin.HandlerFunc
	PermissionCheckAll(permissions ...string) gin.HandlerFunc
}

type AuthImpl struct {
	service service.Service
	logger  *zap.Logger
}

func NewAuth(service service.Service, logger *zap.Logger) Auth {
	return &AuthImpl{service: service, logger: logger}
}

func (a *AuthImpl) RoleCheck(roleCheckType RoleCheckType, requiredRoles ...enum.RoleCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx = ctx.FromGinCtx(c)
		uniCode, exc := ctx.GetUserUniCode()
		if exc != nil {
			ctx.ToError(exc)
			return
		}
		userRoles, exc := a.service.UserRole().GetCachedRolesCodeByUniCode(ctx.Ctx, uniCode)
		if exc != nil {
			ctx.ToError(exc)
			return
		}
		required := make([]string, len(requiredRoles))
		for i, role := range requiredRoles {
			required[i] = role.String()
		}
		if !checkValues(roleCheckType, userRoles, required) {
			a.logger.Warn("role check failed", zap.Any("userRoles", userRoles), zap.Any("requiredRoles", requiredRoles))
			ctx.ToError(exception.Forbidden)
			return
		}
		c.Next()
	}
}

// checkRoles 检查用户角色是否满足要求
func checkValues(checkType RoleCheckType, actual, required []string) bool {
	if len(required) == 0 {
		return true
	}

	valueSet := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		valueSet[value] = struct{}{}
	}

	switch checkType {
	case RoleCheckTypeAny:
		// 任意一个角色符合即可
		for _, value := range required {
			if _, ok := valueSet[value]; ok {
				return true
			}
		}
		return false
	case RoleCheckTypeAll:
		// 所有角色都必须符合
		for _, value := range required {
			if _, ok := valueSet[value]; !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (a *AuthImpl) permissionCheck(checkType RoleCheckType, required []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCtx := ctx.FromGinCtx(c)
		uniCode, exc := appCtx.GetUserUniCode()
		if exc != nil {
			appCtx.ToError(exc)
			return
		}
		permissions, exc := a.service.Permission().GetCachedPermissionCodesByUniCode(appCtx.Ctx, uniCode)
		if exc != nil {
			appCtx.ToError(exc)
			return
		}
		if !checkValues(checkType, permissions, required) {
			a.logger.Warn("permission check failed", zap.Strings("requiredPermissions", required))
			appCtx.ToError(exception.Forbidden)
			return
		}
		c.Next()
	}
}

func (a *AuthImpl) PermissionCheckAny(permissions ...string) gin.HandlerFunc {
	return a.permissionCheck(RoleCheckTypeAny, permissions)
}

func (a *AuthImpl) PermissionCheckAll(permissions ...string) gin.HandlerFunc {
	return a.permissionCheck(RoleCheckTypeAll, permissions)
}

func (a *AuthImpl) RoleCheckAny(roles ...enum.RoleCode) gin.HandlerFunc {
	return a.RoleCheck(RoleCheckTypeAny, roles...)
}

func (a *AuthImpl) RoleCheckAll(roles ...enum.RoleCode) gin.HandlerFunc {
	return a.RoleCheck(RoleCheckTypeAll, roles...)
}

func (a *AuthImpl) RoleCheckAnyS(roles ...string) gin.HandlerFunc {
	enumRoles := make([]enum.RoleCode, len(roles))
	for i, r := range roles {
		enumRoles[i] = enum.RoleCode(r)
	}
	return a.RoleCheck(RoleCheckTypeAny, enumRoles...)
}

func (a *AuthImpl) RoleCheckAllS(roles ...string) gin.HandlerFunc {
	enumRoles := make([]enum.RoleCode, len(roles))
	for i, r := range roles {
		enumRoles[i] = enum.RoleCode(r)
	}
	return a.RoleCheck(RoleCheckTypeAll, enumRoles...)
}
