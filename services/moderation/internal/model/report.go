package model

import "youtube-code-backend/pkg/common/types"

type ReportStatus string

const (
	ReportStatusOpen      ReportStatus = "open"
	ReportStatusReviewed  ReportStatus = "reviewed"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

type Report struct {
	types.BaseModel
	ReporterID  uint64 `gorm:"index" json:"reporter_id"`
	ContentType string `gorm:"not null" json:"content_type"`
	ContentID   uint64 `gorm:"not null" json:"content_id"`
	Reason      string `gorm:"not null" json:"reason"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"default:open;index" json:"status"`
}

func (Report) TableName() string { return "reports" }

// ValidReportStatuses is the set of allowed report status values.
var ValidReportStatuses = map[string]bool{
	string(ReportStatusOpen):      true,
	string(ReportStatusReviewed):  true,
	string(ReportStatusResolved):  true,
	string(ReportStatusDismissed): true,
}
