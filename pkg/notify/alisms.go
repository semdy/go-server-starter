package notify

import (
	"context"
	"encoding/json"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"go.uber.org/zap"
)

// AlibabaSmsSender sends SMS via Alibaba Cloud Dysmsapi.
type AlibabaSmsSender struct {
	client *dysmsapi.Client
	logger *zap.Logger
}

// NewAlibabaSmsSender creates an Alibaba Cloud SMS sender.
// accessKeyID/accessKeySecret are obtained from the Alibaba Cloud RAM console.
func NewAlibabaSmsSender(accessKeyID, accessKeySecret string, logger *zap.Logger) (*AlibabaSmsSender, error) {
	cfg := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
	}
	cfg.Endpoint = stringPtr("dysmsapi.aliyuncs.com")

	client, err := dysmsapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create alibaba sms client: %w", err)
	}
	return &AlibabaSmsSender{client: client, logger: logger}, nil
}

// SendSMS sends an SMS verification code via Alibaba Cloud.
func (s *AlibabaSmsSender) SendSMS(ctx context.Context, mobile, signName, templateCode string, params map[string]string) error {
	paramJSON, _ := json.Marshal(params)
	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  stringPtr(mobile),
		SignName:      stringPtr(signName),
		TemplateCode:  stringPtr(templateCode),
		TemplateParam: stringPtr(string(paramJSON)),
	}

	resp, err := s.client.SendSms(req)
	if err != nil {
		s.logger.Error("alibaba sms send failed", zap.String("mobile", mobile), zap.Error(err))
		return fmt.Errorf("send sms: %w", err)
	}
	if resp.Body.Code != nil && *resp.Body.Code != "OK" {
		s.logger.Error("alibaba sms API returned error",
			zap.String("code", *resp.Body.Code),
			zap.String("message", *resp.Body.Message),
		)
		return fmt.Errorf("sms API error: %s - %s", *resp.Body.Code, *resp.Body.Message)
	}
	return nil
}
