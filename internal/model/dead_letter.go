package model

import "time"

// DeadLetter stores a task that exhausted all retries.
type DeadLetter struct {
	Model
	TaskType    string    `gorm:"not null;index" json:"taskType"`
	TaskID      string    `gorm:"not null;uniqueIndex" json:"taskId"`
	Queue       string    `gorm:"not null;index" json:"queue"`
	Payload     []byte    `gorm:"type:longblob" json:"payload"`
	Error       string    `gorm:"type:text" json:"error"`
	Attempt     int       `json:"attempt"`
	MaxRetry    int       `json:"maxRetry"`
	FailedAt    time.Time `gorm:"not null;index" json:"failedAt"`
	IsRetried   bool      `gorm:"default:false;index" json:"isRetried"`
	RetriedAt   *time.Time `json:"retriedAt,omitempty"`
}
