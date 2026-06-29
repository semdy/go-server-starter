package dto

// ArchivedTaskItem represents an archived (retry-exhausted) task.
type ArchivedTaskItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Queue    string `json:"queue"`
	Payload  []byte `json:"payload"`
	Retried  int    `json:"retried"`
	MaxRetry int    `json:"maxRetry"`
	LastErr  string `json:"lastErr"`
}

// RunTaskReq is the request body to retry a specific archived task.
type RunTaskReq struct {
	Queue  string `json:"queue" binding:"required"`
	TaskID string `json:"taskId" binding:"required"`
}
