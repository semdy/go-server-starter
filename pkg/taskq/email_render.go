package taskq

import (
	"fmt"

	"go-server-starter/pkg/notify/template"
)

// WelcomeEmailData is the data passed to the welcome email template.
type WelcomeEmailData struct {
	Nickname    string
	UserUniCode string
	Email       string
	Mobile      string
}

// welcomeEmailHTML renders the welcome email template with the given payload.
func welcomeEmailHTML(payload EmailWelcomePayload) (string, error) {
	data := WelcomeEmailData{
		Nickname:    payload.Nickname,
		UserUniCode: payload.UserUniCode,
		Email:       payload.Email,
	}
	html, err := template.GetEngine().Render("welcome_email.html", data)
	if err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return html, nil
}

// verifyCodeEmailHTML renders the verification code email template.
func verifyCodeEmailHTML(payload SendEmailCodePayload) (string, error) {
	data := map[string]string{"Code": payload.Code}
	html, err := template.GetEngine().Render("verify_code_email.html", data)
	if err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return html, nil
}
