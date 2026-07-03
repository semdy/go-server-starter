package model

type User struct {
	Model
	TenantID    uint64     `json:"tenantId"`
	UniCode     string     `json:"uniCode"`
	Active      bool       `json:"active"`
	Email       string     `json:"email"`
	Mobile      string     `json:"mobile"`
	CountryCode string     `json:"countryCode"`
	Desc        string     `json:"desc"`
	Password    string     `json:"-"`
	Salt        string     `json:"-"`
	Nickname    string     `json:"nickname"`
	AvatarURL   string     `json:"avatarURL"`
	Roles       []UserRole `gorm:"many2many:user_role_refs" json:"roles"`
	Tenants     []Tenant   `gorm:"many2many:user_tenant_refs" json:"tenants,omitempty"`
}

func (User) TableName() string {
	return "users"
}
