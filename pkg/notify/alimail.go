package notify

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dm "github.com/alibabacloud-go/dm-20151123/v2/client"
	"go.uber.org/zap"
)

// AlibabaEmailSender sends emails via Alibaba Cloud DirectMail.
type AlibabaEmailSender struct {
	client   *dm.Client
	fromAddr string
	fromName string
	logger   *zap.Logger
}

// NewAlibabaEmailSender creates an Alibaba Cloud DirectMail sender.
func NewAlibabaEmailSender(accessKeyID, accessKeySecret, fromAddress, fromName string, logger *zap.Logger) (*AlibabaEmailSender, error) {
	cfg := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
	}
	cfg.Endpoint = stringPtr("dm.aliyuncs.com")

	client, err := dm.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create alibaba dm client: %w", err)
	}
	return &AlibabaEmailSender{
		client:   client,
		fromAddr: fromAddress,
		fromName: fromName,
		logger:   logger,
	}, nil
}

// SendEmail sends an email via Alibaba Cloud DirectMail (single send).
func (s *AlibabaEmailSender) SendEmail(ctx context.Context, to, subject, bodyHTML string) error {
	fromAlias := fmt.Sprintf("%s<%s>", s.fromName, s.fromAddr)
	req := &dm.SingleSendMailRequest{
		AccountName:    stringPtr(s.fromAddr),
		FromAlias:      stringPtr(fromAlias),
		ToAddress:      stringPtr(to),
		Subject:        stringPtr(subject),
		HtmlBody:       stringPtr(bodyHTML),
		ReplyToAddress: boolPtr(true),
		AddressType:    int32Ptr(1),
	}

	resp, err := s.client.SingleSendMail(req)
	if err != nil {
		s.logger.Error("alibaba email send failed", zap.String("to", to), zap.Error(err))
		return fmt.Errorf("send email: %w", err)
	}
	if resp.Body == nil {
		return fmt.Errorf("empty response from DM API")
	}
	return nil
}
