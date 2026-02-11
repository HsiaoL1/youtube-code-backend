package service

import (
	"strings"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/comment/internal/model"
	"youtube-code-backend/services/comment/internal/repository"
)

type SensitiveWordService struct {
	repo *repository.SensitiveWordRepository
}

func NewSensitiveWordService(repo *repository.SensitiveWordRepository) *SensitiveWordService {
	return &SensitiveWordService{repo: repo}
}

// AddSensitiveWordRequest is the payload for adding a sensitive word.
type AddSensitiveWordRequest struct {
	Word     string `json:"word"`
	Severity string `json:"severity"`
	IsActive *bool  `json:"is_active,omitempty"`
}

func (s *SensitiveWordService) List() ([]model.SensitiveWord, error) {
	words, err := s.repo.FindAll()
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to list sensitive words")
	}
	return words, nil
}

func (s *SensitiveWordService) Add(req AddSensitiveWordRequest) (*model.SensitiveWord, error) {
	if strings.TrimSpace(req.Word) == "" {
		return nil, errors.ErrValidation.WithMessage("word is required")
	}

	// Check uniqueness
	if _, err := s.repo.FindByWord(req.Word); err == nil {
		return nil, errors.ErrConflict.WithMessage("sensitive word already exists")
	}

	severity := req.Severity
	if severity == "" {
		severity = "warn"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	word := &model.SensitiveWord{
		Word:     strings.TrimSpace(req.Word),
		Severity: severity,
		IsActive: isActive,
	}

	if err := s.repo.Create(word); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to add sensitive word")
	}
	return word, nil
}

func (s *SensitiveWordService) Delete(id uint64) error {
	if err := s.repo.Delete(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete sensitive word")
	}
	return nil
}
