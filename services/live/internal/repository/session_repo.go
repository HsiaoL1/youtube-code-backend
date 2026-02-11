package repository

import (
	"youtube-code-backend/services/live/internal/model"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *model.LiveSession) error {
	return r.db.Create(session).Error
}

func (r *SessionRepository) Update(session *model.LiveSession) error {
	return r.db.Save(session).Error
}

func (r *SessionRepository) FindActiveByRoomID(roomID uint64) (*model.LiveSession, error) {
	var session model.LiveSession
	if err := r.db.Where("room_id = ? AND ended_at IS NULL", roomID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) FindByRoomID(roomID uint64, offset, limit int) ([]model.LiveSession, int64, error) {
	var sessions []model.LiveSession
	var total int64

	query := r.db.Model(&model.LiveSession{}).Where("room_id = ?", roomID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("started_at DESC").Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
