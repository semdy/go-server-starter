package dto

type AuthLoginByMobileAndCodeReqDto struct {
	Mobile      string `json:"mobile" binding:"required"`
	Code        string `json:"code" binding:"required"`
	CountryCode string `json:"countryCode" binding:"required"`
}

type AuthLoginByEmailAndCodeReqDto struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type AuthTokenResDto struct {
	Token           string             `json:"token"`
	CurrentTenantID uint64             `json:"currentTenantId"`
	Tenants         []AuthTenantResDto `json:"tenants"`
	Roles           []string           `json:"roles"`
	Permissions     []string           `json:"permissions"`
}

type AuthTenantResDto struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type MyTenantsResDto struct {
	CurrentTenantID uint64             `json:"currentTenantId"`
	Tenants         []AuthTenantResDto `json:"tenants"`
}

// SwitchTenantReqDto is the request body for switching to a different tenant.
type SwitchTenantReqDto struct {
	TenantID uint64 `json:"tenantId" binding:"required"`
}
