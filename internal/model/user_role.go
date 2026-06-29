package model

import "go-server-starter/internal/enum"

type UserRole struct {
	Model
	Code    enum.RoleCode `json:"code"`
	Enabled bool          `json:"enabled"`
}
