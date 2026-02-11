package repository

import (
	"youtube-code-backend/services/moderation/internal/model"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(report *model.Report) error {
	return r.db.Create(report).Error
}

func (r *ReportRepository) FindByID(id uint64) (*model.Report, error) {
	var report model.Report
	if err := r.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) Update(report *model.Report) error {
	return r.db.Save(report).Error
}

func (r *ReportRepository) FindAll(status string, offset, limit int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Model(&model.Report{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}
	return reports, total, nil
}
