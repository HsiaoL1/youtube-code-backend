package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/moderation/internal/model"
	"youtube-code-backend/services/moderation/internal/repository"
)

type ModerationService struct {
	itemRepo        *repository.ModerationItemRepository
	reportRepo      *repository.ReportRepository
	enforcementRepo *repository.EnforcementActionRepository
	auditRepo       *repository.AuditLogRepository
}

func NewModerationService(
	itemRepo *repository.ModerationItemRepository,
	reportRepo *repository.ReportRepository,
	enforcementRepo *repository.EnforcementActionRepository,
	auditRepo *repository.AuditLogRepository,
) *ModerationService {
	return &ModerationService{
		itemRepo:        itemRepo,
		reportRepo:      reportRepo,
		enforcementRepo: enforcementRepo,
		auditRepo:       auditRepo,
	}
}

// --- Request DTOs ---

type CreateReportRequest struct {
	ContentType string `json:"content_type"`
	ContentID   uint64 `json:"content_id"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

type RejectItemRequest struct {
	Reason string `json:"reason"`
}

type UpdateReportStatusRequest struct {
	Status string `json:"status"`
}

type CreateEnforcementRequest struct {
	UserID     uint64     `json:"user_id"`
	ActionType string     `json:"action_type"`
	Reason     string     `json:"reason"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// --- Moderation Queue ---

func (s *ModerationService) ListQueue(contentType string, offset, limit int) ([]model.ModerationItem, int64, error) {
	items, total, err := s.itemRepo.FindPending(contentType, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list moderation queue")
	}
	return items, total, nil
}

func (s *ModerationService) GetItem(id uint64) (*model.ModerationItem, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("moderation item not found")
		}
		return nil, errors.ErrInternal
	}
	return item, nil
}

func (s *ModerationService) ApproveItem(id uint64, reviewerID uint64, ipAddress string) (*model.ModerationItem, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("moderation item not found")
		}
		return nil, errors.ErrInternal
	}

	if item.Status != string(model.ModerationStatusPending) {
		return nil, errors.ErrBadRequest.WithMessage("item is not in pending status")
	}

	now := time.Now()
	item.Status = string(model.ModerationStatusApproved)
	item.Decision = "approved"
	item.ReviewedBy = reviewerID
	item.ReviewedAt = &now

	if err := s.itemRepo.Update(item); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to approve item")
	}

	_ = s.auditRepo.Create(&model.AuditLog{
		ActorID:    reviewerID,
		Action:     "approve_item",
		TargetType: "moderation_item",
		TargetID:   id,
		Details:    fmt.Sprintf("approved moderation item %d (content_type=%s, content_id=%d)", id, item.ContentType, item.ContentID),
		IPAddress:  ipAddress,
	})

	return item, nil
}

func (s *ModerationService) RejectItem(id uint64, reviewerID uint64, req RejectItemRequest, ipAddress string) (*model.ModerationItem, error) {
	if req.Reason == "" {
		return nil, errors.ErrValidation.WithMessage("reason is required")
	}

	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("moderation item not found")
		}
		return nil, errors.ErrInternal
	}

	if item.Status != string(model.ModerationStatusPending) {
		return nil, errors.ErrBadRequest.WithMessage("item is not in pending status")
	}

	now := time.Now()
	item.Status = string(model.ModerationStatusRejected)
	item.Decision = "rejected"
	item.RejectionReason = req.Reason
	item.ReviewedBy = reviewerID
	item.ReviewedAt = &now

	if err := s.itemRepo.Update(item); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to reject item")
	}

	_ = s.auditRepo.Create(&model.AuditLog{
		ActorID:    reviewerID,
		Action:     "reject_item",
		TargetType: "moderation_item",
		TargetID:   id,
		Details:    fmt.Sprintf("rejected moderation item %d: %s", id, req.Reason),
		IPAddress:  ipAddress,
	})

	return item, nil
}

// --- Reports ---

func (s *ModerationService) CreateReport(reporterID uint64, req CreateReportRequest, ipAddress string) (*model.Report, error) {
	if req.ContentType == "" {
		return nil, errors.ErrValidation.WithMessage("content_type is required")
	}
	if !model.ValidContentTypes[req.ContentType] {
		return nil, errors.ErrValidation.WithMessage("invalid content_type")
	}
	if req.ContentID == 0 {
		return nil, errors.ErrValidation.WithMessage("content_id is required")
	}
	if req.Reason == "" {
		return nil, errors.ErrValidation.WithMessage("reason is required")
	}

	report := &model.Report{
		ReporterID:  reporterID,
		ContentType: req.ContentType,
		ContentID:   req.ContentID,
		Reason:      req.Reason,
		Description: req.Description,
		Status:      string(model.ReportStatusOpen),
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create report")
	}

	_ = s.auditRepo.Create(&model.AuditLog{
		ActorID:    reporterID,
		Action:     "create_report",
		TargetType: req.ContentType,
		TargetID:   req.ContentID,
		Details:    fmt.Sprintf("report created: %s", req.Reason),
		IPAddress:  ipAddress,
	})

	return report, nil
}

func (s *ModerationService) ListReports(status string, offset, limit int) ([]model.Report, int64, error) {
	reports, total, err := s.reportRepo.FindAll(status, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list reports")
	}
	return reports, total, nil
}

func (s *ModerationService) UpdateReportStatus(id uint64, actorID uint64, req UpdateReportStatusRequest, ipAddress string) (*model.Report, error) {
	if req.Status == "" {
		return nil, errors.ErrValidation.WithMessage("status is required")
	}
	if !model.ValidReportStatuses[req.Status] {
		return nil, errors.ErrValidation.WithMessage("invalid status value")
	}

	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("report not found")
		}
		return nil, errors.ErrInternal
	}

	report.Status = req.Status
	if err := s.reportRepo.Update(report); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update report status")
	}

	_ = s.auditRepo.Create(&model.AuditLog{
		ActorID:    actorID,
		Action:     "update_report_status",
		TargetType: "report",
		TargetID:   id,
		Details:    fmt.Sprintf("report %d status changed to %s", id, req.Status),
		IPAddress:  ipAddress,
	})

	return report, nil
}

// --- Enforcement Actions ---

func (s *ModerationService) CreateEnforcement(performedBy uint64, req CreateEnforcementRequest, ipAddress string) (*model.EnforcementAction, error) {
	if req.UserID == 0 {
		return nil, errors.ErrValidation.WithMessage("user_id is required")
	}
	if req.ActionType == "" {
		return nil, errors.ErrValidation.WithMessage("action_type is required")
	}
	if !model.ValidActionTypes[req.ActionType] {
		return nil, errors.ErrValidation.WithMessage("invalid action_type")
	}

	action := &model.EnforcementAction{
		UserID:      req.UserID,
		ActionType:  req.ActionType,
		Reason:      req.Reason,
		PerformedBy: performedBy,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.enforcementRepo.Create(action); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create enforcement action")
	}

	_ = s.auditRepo.Create(&model.AuditLog{
		ActorID:    performedBy,
		Action:     "create_enforcement",
		TargetType: "user",
		TargetID:   req.UserID,
		Details:    fmt.Sprintf("enforcement action %s on user %d: %s", req.ActionType, req.UserID, req.Reason),
		IPAddress:  ipAddress,
	})

	return action, nil
}

func (s *ModerationService) ListEnforcementsByUser(userID uint64, offset, limit int) ([]model.EnforcementAction, int64, error) {
	actions, total, err := s.enforcementRepo.FindByUserID(userID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list enforcement actions")
	}
	return actions, total, nil
}

// --- Audit Logs ---

func (s *ModerationService) ListAuditLogs(actorID uint64, action string, offset, limit int) ([]model.AuditLog, int64, error) {
	logs, total, err := s.auditRepo.FindAll(actorID, action, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list audit logs")
	}
	return logs, total, nil
}
