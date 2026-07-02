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
	Token string `json:"token"`
}

// SwitchTenantReqDto is the request body for switching to a different tenant.
type SwitchTenantReqDto struct {
	TenantID uint64 `json:"tenantId" binding:"required"`
}
