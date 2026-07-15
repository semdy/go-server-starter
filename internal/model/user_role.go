package model

type UserRole struct {
	Model
	TenantID    uint64       `json:"tenantId"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	BuiltIn     bool         `json:"builtIn"`
	Enabled     bool         `json:"enabled"`
	Permissions []Permission `gorm:"many2many:role_permission_refs" json:"permissions,omitempty"`
}
