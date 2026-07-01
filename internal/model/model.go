package model

import (
	"time"

	"gorm.io/gorm"
)

// GenerateID is set by app.go to the Snowflake generator.
// GORM BeforeCreate hooks use it to auto-fill primary keys.
var GenerateID func() uint64

// 基础模型 有主键
type Model struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`     // 主键
	CreatedAt *time.Time     `json:"createdAt"`                // 创建时间
	UpdatedAt *time.Time     `json:"updatedAt"`                // 更新时间
	DeletedAt gorm.DeletedAt `json:"-"`           // 软删除
	Version   uint64         `json:"version"`     // 乐观锁
}

// BeforeCreate auto-fills Snowflake ID when the caller hasn't set one.
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == 0 && GenerateID != nil {
		m.ID = GenerateID()
	}
	return nil
}
