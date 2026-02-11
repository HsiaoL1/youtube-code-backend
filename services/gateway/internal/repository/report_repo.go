package repository

import (
	"time"

	"gorm.io/gorm"
)

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

type ReportDBRow struct {
	ID          uint64    `gorm:"column:id"`
	ReporterID  uint64    `gorm:"column:reporter_id"`
	ContentType string    `gorm:"column:content_type"`
	ContentID   uint64    `gorm:"column:content_id"`
	Reason      string    `gorm:"column:reason"`
	Description string    `gorm:"column:description"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// FindAll returns all reports.
func (r *ReportRepo) FindAll() ([]ReportDBRow, error) {
	var rows []ReportDBRow
	err := r.db.Raw(`
		SELECT id, reporter_id, content_type, content_id, reason, COALESCE(description, '') AS description, status, created_at
		FROM reports WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`).Scan(&rows).Error
	return rows, err
}

// UpdateStatus updates a report's status.
func (r *ReportRepo) UpdateStatus(id uint64, status string) error {
	return r.db.Exec(`UPDATE reports SET status = ?, updated_at = NOW() WHERE id = ?`, status, id).Error
}

// CountOpen counts open reports.
func (r *ReportRepo) CountOpen() (int64, error) {
	var count int64
	err := r.db.Raw(`SELECT COUNT(*) FROM reports WHERE status = 'open' AND deleted_at IS NULL`).Scan(&count).Error
	return count, err
}
