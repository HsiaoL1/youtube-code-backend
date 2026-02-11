package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/media/internal/model"
	"youtube-code-backend/services/media/internal/repository"
)

type AssetService struct {
	repo *repository.AssetRepository
}

func NewAssetService(repo *repository.AssetRepository) *AssetService {
	return &AssetService{repo: repo}
}

// Asset operations

func (s *AssetService) GetAssetsByVideoID(videoID uint64) ([]model.MediaAsset, error) {
	assets, err := s.repo.FindAssetsByVideoID(videoID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to fetch media assets")
	}
	return assets, nil
}

// Variant operations

func (s *AssetService) GetVariantsByAssetID(assetID uint64) ([]model.MediaVariant, error) {
	// Verify asset exists
	if _, err := s.repo.FindAssetByID(assetID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("media asset not found")
		}
		return nil, errors.ErrInternal
	}

	variants, err := s.repo.FindVariantsByAssetID(assetID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to fetch variants")
	}
	return variants, nil
}

// Thumbnail operations

type AddThumbnailRequest struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsDefault bool   `json:"is_default"`
}

func (s *AssetService) GetThumbnailsByVideoID(videoID uint64) ([]model.MediaThumbnail, error) {
	thumbnails, err := s.repo.FindThumbnailsByVideoID(videoID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to fetch thumbnails")
	}
	return thumbnails, nil
}

func (s *AssetService) AddThumbnail(videoID uint64, req AddThumbnailRequest) (*model.MediaThumbnail, error) {
	if req.URL == "" {
		return nil, errors.ErrValidation.WithMessage("url is required")
	}

	thumbnail := &model.MediaThumbnail{
		VideoID:   videoID,
		URL:       req.URL,
		Width:     req.Width,
		Height:    req.Height,
		IsDefault: req.IsDefault,
	}

	if err := s.repo.CreateThumbnail(thumbnail); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create thumbnail")
	}

	// If this thumbnail is set as default, update others
	if req.IsDefault {
		_ = s.repo.SetDefaultThumbnail(videoID, thumbnail.ID)
	}

	return thumbnail, nil
}

// Subtitle operations

type AddSubtitleRequest struct {
	Language string `json:"language"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	IsAuto   bool   `json:"is_auto"`
}

func (s *AssetService) GetSubtitlesByVideoID(videoID uint64) ([]model.MediaSubtitle, error) {
	subtitles, err := s.repo.FindSubtitlesByVideoID(videoID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to fetch subtitles")
	}
	return subtitles, nil
}

func (s *AssetService) AddSubtitle(videoID uint64, req AddSubtitleRequest) (*model.MediaSubtitle, error) {
	if req.Language == "" {
		return nil, errors.ErrValidation.WithMessage("language is required")
	}
	if req.Label == "" {
		return nil, errors.ErrValidation.WithMessage("label is required")
	}
	if req.URL == "" {
		return nil, errors.ErrValidation.WithMessage("url is required")
	}

	subtitle := &model.MediaSubtitle{
		VideoID:  videoID,
		Language: req.Language,
		Label:    req.Label,
		URL:      req.URL,
		IsAuto:   req.IsAuto,
	}

	if err := s.repo.CreateSubtitle(subtitle); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create subtitle")
	}

	return subtitle, nil
}

func (s *AssetService) DeleteSubtitle(id uint64) error {
	if _, err := s.repo.FindSubtitleByID(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("subtitle not found")
		}
		return errors.ErrInternal
	}

	if err := s.repo.DeleteSubtitle(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete subtitle")
	}
	return nil
}
