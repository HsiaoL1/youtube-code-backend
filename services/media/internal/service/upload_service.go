package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/media/internal/model"
	"youtube-code-backend/services/media/internal/repository"
)

type UploadService struct {
	repo *repository.UploadRepository
}

func NewUploadService(repo *repository.UploadRepository) *UploadService {
	return &UploadService{repo: repo}
}

type InitUploadRequest struct {
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	MimeType   string `json:"mime_type"`
	PartsTotal int    `json:"parts_total"`
}

type UploadPartRequest struct {
	PartNumber int    `json:"part_number"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
}

func (s *UploadService) InitUpload(userID uint64, req InitUploadRequest) (*model.UploadSession, error) {
	if req.Filename == "" {
		return nil, errors.ErrValidation.WithMessage("filename is required")
	}
	if req.FileSize <= 0 {
		return nil, errors.ErrValidation.WithMessage("file_size must be positive")
	}
	if req.MimeType == "" {
		return nil, errors.ErrValidation.WithMessage("mime_type is required")
	}
	if req.PartsTotal < 1 {
		return nil, errors.ErrValidation.WithMessage("parts_total must be at least 1")
	}

	sessionUUID := uuid.New().String()
	storageKey := fmt.Sprintf("uploads/%d/%s/%s", userID, sessionUUID, req.Filename)

	session := &model.UploadSession{
		SessionUUID:   sessionUUID,
		UserID:        userID,
		Filename:      req.Filename,
		FileSize:      req.FileSize,
		MimeType:      req.MimeType,
		Status:        model.UploadStatusInitiated,
		StorageKey:    storageKey,
		PartsTotal:    req.PartsTotal,
		PartsUploaded: 0,
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create upload session")
	}

	return session, nil
}

func (s *UploadService) UploadPart(userID uint64, sessionUUID string, req UploadPartRequest) (*model.UploadPart, error) {
	session, err := s.repo.FindSessionByUUID(sessionUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("upload session not found")
		}
		return nil, errors.ErrInternal
	}

	if session.UserID != userID {
		return nil, errors.ErrForbidden.WithMessage("not your upload session")
	}

	if session.Status == model.UploadStatusCompleted || session.Status == model.UploadStatusFailed {
		return nil, errors.ErrBadRequest.WithMessage("upload session is " + string(session.Status))
	}

	if req.PartNumber < 1 || req.PartNumber > session.PartsTotal {
		return nil, errors.ErrValidation.WithMessage("invalid part number")
	}

	if req.Size <= 0 {
		return nil, errors.ErrValidation.WithMessage("part size must be positive")
	}

	// Update session status to uploading if it was initiated
	if session.Status == model.UploadStatusInitiated {
		_ = s.repo.UpdateSessionStatus(session.ID, model.UploadStatusUploading)
	}

	part := &model.UploadPart{
		SessionID:  session.ID,
		PartNumber: req.PartNumber,
		Size:       req.Size,
		ETag:       req.ETag,
		Status:     model.PartStatusUploaded,
	}

	if err := s.repo.CreatePart(part); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create upload part")
	}

	_ = s.repo.IncrementPartsUploaded(session.ID)

	return part, nil
}

func (s *UploadService) CompleteUpload(userID uint64, sessionUUID string) (*model.UploadSession, error) {
	session, err := s.repo.FindSessionByUUID(sessionUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("upload session not found")
		}
		return nil, errors.ErrInternal
	}

	if session.UserID != userID {
		return nil, errors.ErrForbidden.WithMessage("not your upload session")
	}

	if session.Status == model.UploadStatusCompleted {
		return nil, errors.ErrBadRequest.WithMessage("upload already completed")
	}

	if session.Status == model.UploadStatusFailed {
		return nil, errors.ErrBadRequest.WithMessage("upload session has failed")
	}

	session.Status = model.UploadStatusCompleted
	if err := s.repo.UpdateSession(session); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to complete upload")
	}

	return session, nil
}

func (s *UploadService) AbortUpload(userID uint64, sessionUUID string) error {
	session, err := s.repo.FindSessionByUUID(sessionUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("upload session not found")
		}
		return errors.ErrInternal
	}

	if session.UserID != userID {
		return errors.ErrForbidden.WithMessage("not your upload session")
	}

	if session.Status == model.UploadStatusCompleted {
		return errors.ErrBadRequest.WithMessage("cannot abort a completed upload")
	}

	return s.repo.UpdateSessionStatus(session.ID, model.UploadStatusFailed)
}

func (s *UploadService) GetSessionStatus(userID uint64, sessionUUID string) (*model.UploadSession, []model.UploadPart, error) {
	session, err := s.repo.FindSessionByUUID(sessionUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.ErrNotFound.WithMessage("upload session not found")
		}
		return nil, nil, errors.ErrInternal
	}

	if session.UserID != userID {
		return nil, nil, errors.ErrForbidden.WithMessage("not your upload session")
	}

	parts, err := s.repo.FindPartsBySessionID(session.ID)
	if err != nil {
		return nil, nil, errors.ErrInternal.WithMessage("failed to fetch upload parts")
	}

	return session, parts, nil
}
