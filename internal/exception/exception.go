package exception

import (
	"fmt"
	"go-server-starter/internal/i18n"
)

type Exception struct {
	StatusCode int       `json:"-"`
	Code       int       `json:"code"`
	Message    string    `json:"message"` // Will be set during translation
	Details    []string  `json:"details"`
	I18nMsg    i18n.Text `json:"-"` // i18n message for translation
}

var codes = map[int]struct{}{}

// mustRegister registers an exception code and panics if it already exists.
// Panicking is intentional here: exception codes are defined in package-level var
// declarations during init, so duplicate codes indicate a programming error that
// must be caught at startup. This follows the Go MustXxx convention.
func mustRegister(statusCode int, code int, message string, i18nMsg i18n.Text) *Exception {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("Exception code %d already exists", code))
	}
	codes[code] = struct{}{}
	return &Exception{StatusCode: statusCode, Code: code, Message: message, Details: []string{}, I18nMsg: i18nMsg}
}

func NewException(statusCode int, code int, message string, i18nMsg i18n.Text) *Exception {
	return mustRegister(statusCode, code, message, i18nMsg)
}

func (e *Exception) clone() *Exception {
	ne := *e
	return &ne
}

func (e *Exception) Append(details ...string) *Exception {
	e = e.clone()
	e.Details = append(e.Details, details...)
	return e
}

// Error implements the error interface so *Exception can be used with standard
// error-returning patterns (e.g. singleflight.Do).
func (e *Exception) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// CodeEquals returns true if both exceptions share the same business error code.
func (e *Exception) CodeEquals(err *Exception) bool {
	if e == nil || err == nil {
		return false
	}
	return e.Code == err.Code
}
