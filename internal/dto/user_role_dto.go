package dto

// UserRoleCreateReqDto is the request DTO for creating a user role.
type UserRoleCreateReqDto struct {
	Code          string   `json:"code" form:"code" binding:"required,min=2,max=50"`
	Name          string   `json:"name" form:"name" binding:"required,min=2,max=100"`
	Description   string   `json:"description" form:"description" binding:"max=255"`
	Enabled       *bool    `json:"enabled" form:"enabled"`
	PermissionIDs []uint64 `json:"permissionIds" form:"permissionIds"`
}

// UserRoleUpdateReqDto is the request DTO for updating a user role.
type UserRoleUpdateReqDto struct {
	Code        *string `json:"code" form:"code" binding:"omitempty,min=2,max=50"`
	Name        *string `json:"name" form:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" form:"description" binding:"omitempty,max=255"`
	Enabled     *bool   `json:"enabled" form:"enabled"`
}

type UserRoleSetPermissionsReqDto struct {
	PermissionIDs []uint64 `json:"permissionIds"`
}

// UserRoleResDto is the response DTO for a user role.
type UserRoleResDto struct {
	ID          uint64             `json:"id"`
	TenantID    uint64             `json:"tenantId"`
	Code        string             `json:"code"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	BuiltIn     bool               `json:"builtIn"`
	Enabled     bool               `json:"enabled"`
	Permissions []PermissionResDto `json:"permissions"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
}

// UserRoleTableQueryReqDto is the request DTO for querying user roles with pagination.
type UserRoleTableQueryReqDto struct {
	PaginationReqDto
	Code    *string `json:"code" form:"code"`
	Enabled *bool   `json:"enabled" form:"enabled"`
}
