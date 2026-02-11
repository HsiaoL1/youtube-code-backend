package repository

import (
	"youtube-code-backend/services/media/internal/model"

	"gorm.io/gorm"
)

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// Session operations

func (r *UploadRepository) CreateSession(session *model.UploadSession) error {
	return r.db.Create(session).Error
}

func (r *UploadRepository) FindSessionByUUID(uuid string) (*model.UploadSession, error) {
	var session model.UploadSession
	if err := r.db.Where("session_uuid = ?", uuid).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UploadRepository) FindSessionByID(id uint64) (*model.UploadSession, error) {
	var session model.UploadSession
	if err := r.db.First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UploadRepository) UpdateSessionStatus(id uint64, status model.UploadStatus) error {
	return r.db.Model(&model.UploadSession{}).Where("id = ?", id).Update("status", status).Error
}

func (r *UploadRepository) UpdateSession(session *model.UploadSession) error {
	return r.db.Save(session).Error
}

func (r *UploadRepository) IncrementPartsUploaded(id uint64) error {
	return r.db.Model(&model.UploadSession{}).Where("id = ?", id).
		UpdateColumn("parts_uploaded", gorm.Expr("parts_uploaded + 1")).Error
}

// Part operations

func (r *UploadRepository) CreatePart(part *model.UploadPart) error {
	return r.db.Create(part).Error
}

func (r *UploadRepository) FindPartsBySessionID(sessionID uint64) ([]model.UploadPart, error) {
	var parts []model.UploadPart
	if err := r.db.Where("session_id = ?", sessionID).Order("part_number ASC").Find(&parts).Error; err != nil {
		return nil, err
	}
	return parts, nil
}

func (r *UploadRepository) UpdatePartStatus(id uint64, status model.PartStatus, etag string) error {
	return r.db.Model(&model.UploadPart{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "etag": etag}).Error
}
