package model

import "time"

// DeadLetter stores a task that exhausted all retries.
type DeadLetter struct {
	Model
	TenantID  string     `json:"tenantId"`
	TaskType  string     `json:"taskType"`
	TaskID    string     `json:"taskId"`
	Queue     string     `json:"queue"`
	Payload   []byte     `gorm:"type:longblob" json:"payload"`
	Error     string     `gorm:"type:text" json:"error"`
	Attempt   int        `json:"attempt"`
	MaxRetry  int        `json:"maxRetry"`
	FailedAt  time.Time  `json:"failedAt"`
	IsRetried bool       `json:"isRetried"`
	RetriedAt *time.Time `json:"retriedAt,omitempty"`
}
