package model

type Tenant struct {
	Model
	Name   string `json:"name"`
	Code   string `json:"code"`
	Active bool   `json:"active"`
}
