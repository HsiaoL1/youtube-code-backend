package repository

import (
	"youtube-code-backend/services/media/internal/model"

	"gorm.io/gorm"
)

type AssetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// MediaAsset operations

func (r *AssetRepository) CreateAsset(asset *model.MediaAsset) error {
	return r.db.Create(asset).Error
}

func (r *AssetRepository) FindAssetByID(id uint64) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := r.db.First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *AssetRepository) FindAssetsByVideoID(videoID uint64) ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if err := r.db.Where("video_id = ?", videoID).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *AssetRepository) UpdateAssetStatus(id uint64, status model.AssetStatus) error {
	return r.db.Model(&model.MediaAsset{}).Where("id = ?", id).Update("status", status).Error
}

// MediaVariant operations

func (r *AssetRepository) CreateVariant(variant *model.MediaVariant) error {
	return r.db.Create(variant).Error
}

func (r *AssetRepository) FindVariantsByAssetID(assetID uint64) ([]model.MediaVariant, error) {
	var variants []model.MediaVariant
	if err := r.db.Where("asset_id = ?", assetID).Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

// MediaThumbnail operations

func (r *AssetRepository) CreateThumbnail(thumbnail *model.MediaThumbnail) error {
	return r.db.Create(thumbnail).Error
}

func (r *AssetRepository) FindThumbnailsByVideoID(videoID uint64) ([]model.MediaThumbnail, error) {
	var thumbnails []model.MediaThumbnail
	if err := r.db.Where("video_id = ?", videoID).Find(&thumbnails).Error; err != nil {
		return nil, err
	}
	return thumbnails, nil
}

func (r *AssetRepository) SetDefaultThumbnail(videoID uint64, thumbnailID uint64) error {
	tx := r.db.Begin()
	if err := tx.Model(&model.MediaThumbnail{}).Where("video_id = ?", videoID).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&model.MediaThumbnail{}).Where("id = ? AND video_id = ?", thumbnailID, videoID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// MediaSubtitle operations

func (r *AssetRepository) CreateSubtitle(subtitle *model.MediaSubtitle) error {
	return r.db.Create(subtitle).Error
}

func (r *AssetRepository) FindSubtitlesByVideoID(videoID uint64) ([]model.MediaSubtitle, error) {
	var subtitles []model.MediaSubtitle
	if err := r.db.Where("video_id = ?", videoID).Find(&subtitles).Error; err != nil {
		return nil, err
	}
	return subtitles, nil
}

func (r *AssetRepository) FindSubtitleByID(id uint64) (*model.MediaSubtitle, error) {
	var subtitle model.MediaSubtitle
	if err := r.db.First(&subtitle, id).Error; err != nil {
		return nil, err
	}
	return &subtitle, nil
}

func (r *AssetRepository) DeleteSubtitle(id uint64) error {
	return r.db.Delete(&model.MediaSubtitle{}, id).Error
}
