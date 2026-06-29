package notify

import "context"

// SmsSender sends SMS messages (e.g. verification codes).
type SmsSender interface {
	SendSMS(ctx context.Context, mobile, signName, templateCode string, params map[string]string) error
}

// EmailSender sends emails (e.g. welcome emails, verification codes).
type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, bodyHTML string) error
}

// LogSender implements both interfaces by logging. Use in dev/test.
type LogSender struct{}

func (LogSender) SendSMS(ctx context.Context, mobile, signName, templateCode string, params map[string]string) error {
	return nil
}
func (LogSender) SendEmail(ctx context.Context, to, subject, bodyHTML string) error {
	return nil
}
