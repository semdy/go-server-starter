package dto

type PermissionResDto struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type PermissionTableQueryReqDto struct {
	PaginationReqDto
	Code    *string `json:"code" form:"code"`
	Enabled *bool   `json:"enabled" form:"enabled"`
}

type MyAccessResDto struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
