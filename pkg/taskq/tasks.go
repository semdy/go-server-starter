package taskq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// ----- Task type constants -----
// Register new task types here. Convention: "domain:action".
const (
	TaskEmailWelcome = "email:welcome" // 发送欢迎邮件
)

// ----- Payload types -----

// EmailWelcomePayload is the payload for welcome email tasks.
type EmailWelcomePayload struct {
	UserUniCode string `json:"userUniCode"`
	Email       string `json:"email"`
}

// ----- Retry helpers -----

// RetryByType returns task-specific retry options.
// Use to customize retry count and backoff per task type.
func RetryByType(taskType string) []asynq.Option {
	switch taskType {
	case TaskEmailWelcome:
		return []asynq.Option{
			asynq.MaxRetry(3),
			asynq.Timeout(30 * time.Second),
		}
	default:
		return []asynq.Option{
			asynq.MaxRetry(10),
			asynq.Timeout(60 * time.Second),
		}
	}
}

// ----- Task constructors -----

// NewEmailWelcomeTask creates a task for sending a welcome email.
func NewEmailWelcomeTask(payload EmailWelcomePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEmailWelcome, data), nil
}

// ----- Handlers -----

// HandleEmailWelcome sends a welcome email to a newly registered user.
// TODO: Replace with real email sending logic (SMTP / SendGrid / etc.).
func HandleEmailWelcome(ctx context.Context, task *asynq.Task) error {
	var payload EmailWelcomePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		// Payload corruption — don't retry
		return asynq.SkipRetry
	}

	// TODO: Send real welcome email here
	_ = payload
	return nil
}

// ----- Alerter -----

// Alerter is called when a task exhausts all retries.
// Implementations can send webhooks, Slack messages, emails, etc.
type Alerter interface {
	Alert(ctx context.Context, info AlertInfo)
}

// AlertInfo contains details about a task that failed permanently.
type AlertInfo struct {
	TaskType string `json:"taskType"`
	TaskID   string `json:"taskId"`
	Payload  []byte `json:"payload"`
	Error    string `json:"error"`
	Attempt  int    `json:"attempt"`
	MaxRetry int    `json:"maxRetry"`
}

// noopAlerter is the default, discards alerts.
type noopAlerter struct{}

func (noopAlerter) Alert(_ context.Context, _ AlertInfo) {}
