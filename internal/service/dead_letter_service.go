package service

import (
	"context"
	"errors"
	"time"

	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/taskq"
	"go-server-starter/pkg/utils"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeadLetterService manages dead-letter persistence and retry.
// It also implements taskq.Alerter so it can be plugged into the taskq worker directly.
type DeadLetterService interface {
	taskq.Alerter
	// List returns paginated dead letters.
	List(ctx context.Context, params dto.DeadLetterListReqDto) (*dto.PaginationResDto[[]*dto.DeadLetterItem], *exception.Exception)
	// Retry re-enqueues and marks a dead letter as retried.
	Retry(ctx context.Context, id uint64) *exception.Exception
	// RetryAll re-enqueues all un-retried dead letters.
	RetryAll(ctx context.Context, taskType string) (int, *exception.Exception)
	// Delete hard-deletes a dead letter.
	Delete(ctx context.Context, id uint64) *exception.Exception
	// Store persists a dead letter.
	Store(ctx context.Context, info taskq.AlertInfo)
}

type DeadLetterServiceImpl struct {
	repo   repo.Repo
	taskq  *taskq.Client
	logger *zap.Logger
}

func NewDeadLetterService(repo repo.Repo, taskqClient *taskq.Client, logger *zap.Logger) DeadLetterService {
	return &DeadLetterServiceImpl{repo: repo, taskq: taskqClient, logger: logger}
}

// Alert implements taskq.Alerter. Called by the taskq ErrorHandler when retries are exhausted.
func (s *DeadLetterServiceImpl) Alert(ctx context.Context, info taskq.AlertInfo) {
	s.Store(ctx, info)
}

// Store persists a dead letter to MySQL.
func (s *DeadLetterServiceImpl) Store(ctx context.Context, info taskq.AlertInfo) {
	tenantID := info.TenantID
	if tenantID == nil {
		if tid := cctx.GetTenantID(ctx); tid != 0 {
			tenantID = &tid
		}
	}
	if tenantID != nil && *tenantID == 0 {
		tenantID = nil
	}
	dl := &model.DeadLetter{
		TenantID: tenantID,
		TaskType: info.TaskType,
		TaskID:   info.TaskID,
		Payload:  info.Payload,
		Error:    info.Error,
		Attempt:  info.Attempt,
		MaxRetry: info.MaxRetry,
		FailedAt: time.Now(),
	}
	if err := s.repo.DeadLetter().Create(ctx, dl); err != nil {
		s.logger.Error("failed to persist dead letter",
			zap.String("taskType", info.TaskType),
			zap.String("taskId", info.TaskID),
			zap.Error(err),
		)
	}
}

func (s *DeadLetterServiceImpl) List(ctx context.Context, params dto.DeadLetterListReqDto) (*dto.PaginationResDto[[]*dto.DeadLetterItem], *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Order("failed_at DESC"),
		repo.Where("(tenant_id = ? OR tenant_id IS NULL)", cctx.GetTenantID(ctx)),
		repo.WherePtrNonEmpty("task_type = ?", params.TaskType),
		repo.WherePtrNonEmpty("is_retried = ?", params.IsRetried),
	}
	entries, total, err := s.repo.DeadLetter().GetTable(ctx, params.Page, params.PageSize, opts...)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	items := make([]*dto.DeadLetterItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, &dto.DeadLetterItem{
			ID:        e.ID,
			TaskType:  e.TaskType,
			TaskID:    e.TaskID,
			Queue:     e.Queue,
			Error:     e.Error,
			Attempt:   e.Attempt,
			MaxRetry:  e.MaxRetry,
			FailedAt:  e.FailedAt.Format(time.RFC3339),
			IsRetried: e.IsRetried,
		})
	}
	return utils.AssemblePaginationResDto(items, total, params.Page, params.PageSize), nil
}

func (s *DeadLetterServiceImpl) Retry(ctx context.Context, id uint64) *exception.Exception {
	entry, err := s.repo.DeadLetter().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.NotFound.Append("dead letter not found")
		}
		return exception.InternalServerError.Append(err.Error())
	}
	if entry.TenantID != nil && *entry.TenantID != cctx.GetTenantID(ctx) {
		return exception.Forbidden.Append("dead letter belongs to another tenant")
	}
	if entry.IsRetried {
		return exception.BadRequest.Append("already retried")
	}

	// Reconstruct and re-enqueue the task
	task := asynq.NewTask(entry.TaskType, entry.Payload)
	if _, err := s.taskq.Enqueue(ctx, task, taskq.RetryByType(entry.TaskType)...); err != nil {
		return exception.InternalServerError.Append("re-enqueue failed: " + err.Error())
	}

	// Mark as retried
	now := time.Now()
	entry.IsRetried = true
	entry.RetriedAt = &now
	rows, err := s.repo.DeadLetter().MarkRetriedByIDAndTenantOrSystem(ctx, id, cctx.GetTenantID(ctx), now)
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if rows == 0 {
		return exception.Forbidden.Append("dead letter belongs to another tenant")
	}

	return nil
}

func (s *DeadLetterServiceImpl) RetryAll(ctx context.Context, taskType string) (int, *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Where("task_type = ?", taskType),
		repo.Where("is_retried = ?", false),
		repo.Where("(tenant_id = ? OR tenant_id IS NULL)", cctx.GetTenantID(ctx)),
		repo.Order("failed_at ASC"),
	}
	entries, _, err := s.repo.DeadLetter().GetTable(ctx, 1, 1000, opts...)
	if err != nil {
		return 0, exception.InternalServerError.Append(err.Error())
	}

	count := 0
	for _, entry := range entries {
		task := asynq.NewTask(entry.TaskType, entry.Payload)
		if _, err := s.taskq.Enqueue(ctx, task, taskq.RetryByType(entry.TaskType)...); err != nil {
			s.logger.Warn("failed to re-enqueue dead letter", zap.Uint64("id", entry.ID), zap.Error(err))
			continue
		}
		now := time.Now()
		entry.IsRetried = true
		entry.RetriedAt = &now
		rows, err := s.repo.DeadLetter().MarkRetriedByIDAndTenantOrSystem(ctx, entry.ID, cctx.GetTenantID(ctx), now)
		if err != nil {
			s.logger.Warn("failed to mark dead letter retried", zap.Uint64("id", entry.ID), zap.Error(err))
			continue
		}
		if rows == 0 {
			s.logger.Warn("dead letter no longer belongs to current tenant", zap.Uint64("id", entry.ID))
			continue
		}
		count++
	}

	return count, nil
}

func (s *DeadLetterServiceImpl) Delete(ctx context.Context, id uint64) *exception.Exception {
	entry, err := s.repo.DeadLetter().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.NotFound.Append("dead letter not found")
		}
		return exception.InternalServerError.Append(err.Error())
	}
	if entry.TenantID != nil && *entry.TenantID != cctx.GetTenantID(ctx) {
		return exception.Forbidden.Append("dead letter belongs to another tenant")
	}
	rows, err := s.repo.DeadLetter().HardDeleteByIDAndTenantOrSystem(ctx, id, cctx.GetTenantID(ctx))
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if rows == 0 {
		return exception.Forbidden.Append("dead letter belongs to another tenant")
	}
	return nil
}
