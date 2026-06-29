package dto

// DeadLetterListReqDto is the request DTO for listing dead letters.
type DeadLetterListReqDto struct {
	PaginationReqDto
	TaskType  *string `json:"taskType" form:"taskType"`
	IsRetried *bool   `json:"isRetried" form:"isRetried"`
}

// DeadLetterItem is a single dead letter entry in list responses.
type DeadLetterItem struct {
	ID        uint64 `json:"id"`
	TaskType  string `json:"taskType"`
	TaskID    string `json:"taskId"`
	Queue     string `json:"queue"`
	Error     string `json:"error"`
	Attempt   int    `json:"attempt"`
	MaxRetry  int    `json:"maxRetry"`
	FailedAt  string `json:"failedAt"`
	IsRetried bool   `json:"isRetried"`
}

// RetryDeadLetterReq is the request body for retrying a single dead letter.
type RetryDeadLetterReq struct {
	ID uint64 `json:"id" form:"id" binding:"required"`
}
