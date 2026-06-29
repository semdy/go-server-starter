package dto

// SendSmsCodeReqDto is the request DTO for sending an SMS verification code.
type SendSmsCodeReqDto struct {
	Mobile      string `json:"mobile" form:"mobile" binding:"required"`
	CountryCode string `json:"countryCode" form:"countryCode" binding:"required"`
}

// SendEmailCodeReqDto is the request DTO for sending an email verification code.
type SendEmailCodeReqDto struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}
