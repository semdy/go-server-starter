package service

import (
	"context"
	"time"

	"go-server-starter/internal/constant"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/pkg/taskq"
	"go-server-starter/pkg/verify_code"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// VerifyCodeService handles verification code generation and validation.
type VerifyCodeService interface {
	SendSmsCode(ctx context.Context, params dto.SendSmsCodeReqDto) *exception.Exception
	SendEmailCode(ctx context.Context, params dto.SendEmailCodeReqDto) *exception.Exception
	Validate(typ verify_code.Type, target, code string) *exception.Exception
}

type VerifyCodeServiceImpl struct {
	store  *verify_code.Store
	taskq  *taskq.Client
	logger *zap.Logger
}

func NewVerifyCodeService(redisClient *redis.Client, taskqClient *taskq.Client, logger *zap.Logger) VerifyCodeService {
	return &VerifyCodeServiceImpl{
		store: verify_code.NewStore(redisClient,
			constant.REDIS_EXPIRE_OF_VERIFY_CODE,
			constant.REDIS_EXPIRE_OF_VERIFY_LIMIT,
		),
		taskq:  taskqClient,
		logger: logger,
	}
}

func (s *VerifyCodeServiceImpl) SendSmsCode(ctx context.Context, params dto.SendSmsCodeReqDto) *exception.Exception {
	// Generate and store code
	code, err := s.store.Generate(ctx, verify_code.SMSTypeLogin, params.Mobile)
	if err != nil {
		if err == verify_code.ErrResendCooldown {
			return exception.BadRequest.Append("please wait 60 seconds before requesting another code")
		}
		return exception.InternalServerError.Append(err.Error())
	}

	// Enqueue SMS send task (idempotent within 5 min cooldown window)
	if s.taskq != nil {
		task, _ := taskq.NewSendSMSCodeTask(taskq.SendSMSCodePayload{
			Mobile: params.Mobile,
			Code:   code,
		})
		if task != nil {
			s.taskq.EnqueueUnique(ctx, task,
				taskq.SendSMSCodeUniqueKey(params.Mobile),
				5*time.Minute,
				taskq.RetryByType(taskq.TaskSendSMSCode)...,
			)
		}
	}
	return nil
}

func (s *VerifyCodeServiceImpl) SendEmailCode(ctx context.Context, params dto.SendEmailCodeReqDto) *exception.Exception {
	code, err := s.store.Generate(ctx, verify_code.EmailTypeLogin, params.Email)
	if err != nil {
		if err == verify_code.ErrResendCooldown {
			return exception.BadRequest.Append("please wait 60 seconds before requesting another code")
		}
		return exception.InternalServerError.Append(err.Error())
	}

	// Enqueue email code send task (idempotent within 5 min cooldown window)
	if s.taskq != nil {
		task, _ := taskq.NewSendEmailCodeTask(taskq.SendEmailCodePayload{
			Email: params.Email,
			Code:  code,
		})
		if task != nil {
			s.taskq.EnqueueUnique(ctx, task,
				taskq.SendEmailCodeUniqueKey(params.Email),
				5*time.Minute,
				taskq.RetryByType(taskq.TaskSendEmailCode)...,
			)
		}
	}
	return nil
}

func (s *VerifyCodeServiceImpl) Validate(typ verify_code.Type, target, code string) *exception.Exception {
	if err := s.store.Validate(context.Background(), typ, target, code); err != nil {
		switch err {
		case verify_code.ErrCodeExpired:
			return exception.BadRequest.Append("verification code expired")
		case verify_code.ErrCodeMismatch:
			return exception.UserMobileVerificationCodeIsIncorrect
		default:
			return exception.InternalServerError.Append(err.Error())
		}
	}
	return nil
}
