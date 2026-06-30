package model

type Tenant struct {
	Model
	Name   string `json:"name"`
	Code   string `gorm:"uniqueIndex;not null" json:"code"`
	Status string `gorm:"default:'active'" json:"status"` // active, disabled
}
