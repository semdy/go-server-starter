package dto

// UserRoleCreateReqDto is the request DTO for creating a user role.
type UserRoleCreateReqDto struct {
	Code    string `json:"code" form:"code" validate:"required,min=1,max=50"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

// UserRoleUpdateReqDto is the request DTO for updating a user role.
type UserRoleUpdateReqDto struct {
	Code    *string `json:"code" form:"code" validate:"omitempty,min=1,max=50"`
	Enabled *bool   `json:"enabled" form:"enabled"`
}

// UserRoleResDto is the response DTO for a user role.
type UserRoleResDto struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// UserRoleTableQueryReqDto is the request DTO for querying user roles with pagination.
type UserRoleTableQueryReqDto struct {
	PaginationReqDto
	Code    *string `json:"code" form:"code"`
	Enabled *bool   `json:"enabled" form:"enabled"`
}
