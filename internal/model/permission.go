package model

type Permission struct {
	Model
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func (Permission) TableName() string {
	return "permissions"
}
