package dto

type TenantCreateReqDto struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Code string `json:"code" binding:"required,min=2,max=64"`
}

type TenantUpdateReqDto struct {
	Name   *string `json:"name" binding:"omitempty,min=2,max=100"`
	Status *string `json:"status" binding:"omitempty,oneof=active disabled"`
}

type TenantResDto struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type TenantTableQueryReqDto struct {
	PaginationReqDto
	Status *string `json:"status" form:"status"`
}
