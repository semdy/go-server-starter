package constant

const (
	PermissionUserRead        = "user.read"
	PermissionUserCreate      = "user.create"
	PermissionUserUpdate      = "user.update"
	PermissionUserDelete      = "user.delete"
	PermissionUserAssignRoles = "user.assign_roles"

	PermissionRoleRead              = "role.read"
	PermissionRoleCreate            = "role.create"
	PermissionRoleUpdate            = "role.update"
	PermissionRoleDelete            = "role.delete"
	PermissionRoleAssignPermissions = "role.assign_permissions"
	PermissionPermissionRead        = "permission.read"

	PermissionTenantRead   = "tenant.read"
	PermissionTenantCreate = "tenant.create"
	PermissionTenantUpdate = "tenant.update"
	PermissionTenantDelete = "tenant.delete"

	PermissionDeadLetterRead   = "dead_letter.read"
	PermissionDeadLetterRetry  = "dead_letter.retry"
	PermissionDeadLetterDelete = "dead_letter.delete"
)

var BuiltInPermissions = []struct {
	Code string
	Name string
}{
	{PermissionUserRead, "查看用户"},
	{PermissionUserCreate, "创建用户"},
	{PermissionUserUpdate, "更新用户"},
	{PermissionUserDelete, "删除用户"},
	{PermissionUserAssignRoles, "分配用户角色"},
	{PermissionRoleRead, "查看角色"},
	{PermissionRoleCreate, "创建角色"},
	{PermissionRoleUpdate, "更新角色"},
	{PermissionRoleDelete, "删除角色"},
	{PermissionRoleAssignPermissions, "配置角色权限"},
	{PermissionPermissionRead, "查看权限字典"},
	{PermissionTenantRead, "查看租户"},
	{PermissionTenantCreate, "创建租户"},
	{PermissionTenantUpdate, "更新租户"},
	{PermissionTenantDelete, "删除租户"},
	{PermissionDeadLetterRead, "查看死信"},
	{PermissionDeadLetterRetry, "重试死信"},
	{PermissionDeadLetterDelete, "删除死信"},
}

var AdminPermissions = []string{
	PermissionUserRead,
	PermissionUserCreate,
	PermissionUserUpdate,
	PermissionUserDelete,
	PermissionUserAssignRoles,
	PermissionRoleRead,
	PermissionRoleCreate,
	PermissionRoleUpdate,
	PermissionRoleDelete,
	PermissionRoleAssignPermissions,
	PermissionPermissionRead,
}

// TenantManagementPermissions are platform-level capabilities. They belong
// exclusively to the built-in super_admin role and cannot be granted to
// tenant custom roles.
var TenantManagementPermissions = []string{
	PermissionTenantRead,
	PermissionTenantCreate,
	PermissionTenantUpdate,
	PermissionTenantDelete,
}

func IsTenantManagementPermission(code string) bool {
	for _, permission := range TenantManagementPermissions {
		if code == permission {
			return true
		}
	}
	return false
}
