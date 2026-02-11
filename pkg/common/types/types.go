package types

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel is the common model for all database tables.
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// PaginationRequest holds pagination parameters from query strings.
type PaginationRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

// Normalize sets defaults and caps for pagination parameters.
func (p *PaginationRequest) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset returns the SQL offset for the current page.
func (p *PaginationRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PaginationMeta holds pagination metadata for responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationMeta creates pagination metadata from request and total count.
func NewPaginationMeta(p PaginationRequest, total int64) PaginationMeta {
	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize > 0 {
		totalPages++
	}
	return PaginationMeta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
