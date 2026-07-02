package taskq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// ----- Task type constants -----
// Register new task types here. Convention: "domain:action".
const (
	TaskEmailWelcome  = "email:welcome"   // 发送欢迎邮件
	TaskSendSMSCode   = "sms:send_code"   // 发送短信验证码
	TaskSendEmailCode = "email:send_code" // 发送邮箱验证码
)

// ----- Payload types -----

// EmailWelcomePayload is the payload for welcome email tasks.
type EmailWelcomePayload struct {
	UserUniCode string `json:"userUniCode"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
}

// SendSMSCodePayload is the payload for sending an SMS verification code.
type SendSMSCodePayload struct {
	Mobile string `json:"mobile"`
	Code   string `json:"code"`
}

// SendEmailCodePayload is the payload for sending an email verification code.
type SendEmailCodePayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// ----- Handler dependencies (injected at startup) -----

// HandlerDeps holds the dependencies task handlers need.
// Set these before registering handlers in app.go.
var HandlerDeps struct {
	Logger *zap.Logger
	EmailSender interface {
		SendEmail(ctx context.Context, to, subject, bodyHTML string) error
	}
	SmsSender interface {
		SendSMS(ctx context.Context, mobile, signName, templateCode string, params map[string]string) error
	}
	SMSSignName     string
	SMSTemplateCode string
}

// ----- Unique key helpers -----

// UniqueKey returns a unique dedup key for the given task type and ID.
func UniqueKey(taskType, id string) string {
	return taskType + ":" + id
}

// WelcomeEmailUniqueKey returns the dedup key for a welcome email task.
func WelcomeEmailUniqueKey(userUniCode string) string {
	return UniqueKey(TaskEmailWelcome, userUniCode)
}

// SendSMSCodeUniqueKey returns the dedup key for an SMS code task.
func SendSMSCodeUniqueKey(mobile string) string {
	return UniqueKey(TaskSendSMSCode, mobile)
}

// SendEmailCodeUniqueKey returns the dedup key for an email code task.
func SendEmailCodeUniqueKey(email string) string {
	return UniqueKey(TaskSendEmailCode, email)
}

// ----- Retry helpers -----

// RetryByType returns task-specific retry options.
func RetryByType(taskType string) []asynq.Option {
	switch taskType {
	case TaskEmailWelcome:
		return []asynq.Option{
			asynq.MaxRetry(3),
			asynq.Timeout(30 * time.Second),
		}
	case TaskSendSMSCode, TaskSendEmailCode:
		return []asynq.Option{
			asynq.MaxRetry(1), // send-code failures rarely benefit from retry
			asynq.Timeout(15 * time.Second),
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

// NewSendSMSCodeTask creates a task for sending an SMS verification code.
func NewSendSMSCodeTask(payload SendSMSCodePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendSMSCode, data), nil
}

// NewSendEmailCodeTask creates a task for sending an email verification code.
func NewSendEmailCodeTask(payload SendEmailCodePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskSendEmailCode, data), nil
}

// ----- Handlers -----

// HandleEmailWelcome sends a welcome email to a newly registered user.
func HandleEmailWelcome(ctx context.Context, task *asynq.Task) error {
	var payload EmailWelcomePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return asynq.SkipRetry
	}

	if HandlerDeps.Logger != nil {
		HandlerDeps.Logger.Info("sending welcome email",
			zap.String("email", payload.Email),
			zap.String("uniCode", payload.UserUniCode),
		)
	}

	// Render welcome email from template
	html, err := welcomeEmailHTML(payload)
	if err != nil {
		return fmt.Errorf("render welcome email: %w", err)
	}

	return HandlerDeps.EmailSender.SendEmail(ctx, payload.Email, "Welcome to Go Server Starter", html)
}

// HandleSendSMSCode sends an SMS verification code via Alibaba Cloud SMS.
func HandleSendSMSCode(ctx context.Context, task *asynq.Task) error {
	var payload SendSMSCodePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return asynq.SkipRetry
	}

	params := map[string]string{"code": payload.Code}
	return HandlerDeps.SmsSender.SendSMS(ctx, payload.Mobile, HandlerDeps.SMSSignName, HandlerDeps.SMSTemplateCode, params)
}

// HandleSendEmailCode sends an email verification code via the configured email sender.
func HandleSendEmailCode(ctx context.Context, task *asynq.Task) error {
	var payload SendEmailCodePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return asynq.SkipRetry
	}

	// Render verification code email from template
	html, err := verifyCodeEmailHTML(payload)
	if err != nil {
		return fmt.Errorf("render verify code email: %w", err)
	}

	return HandlerDeps.EmailSender.SendEmail(ctx, payload.Email,
		fmt.Sprintf("Your verification code: %s", payload.Code), html)
}

// ----- Alerter -----

// Alerter is called when a task exhausts all retries.
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
