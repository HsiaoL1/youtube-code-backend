package repository

import (
	"gorm.io/gorm"
)

type PlaylistRepo struct {
	db *gorm.DB
}

func NewPlaylistRepo(db *gorm.DB) *PlaylistRepo {
	return &PlaylistRepo{db: db}
}

type PlaylistDBRow struct {
	ID    uint64 `gorm:"column:id"`
	Title string `gorm:"column:title"`
}

type PlaylistItemDBRow struct {
	PlaylistID uint64 `gorm:"column:playlist_id"`
	VideoID    uint64 `gorm:"column:video_id"`
	Position   int    `gorm:"column:position"`
}

// FindByID returns a playlist.
func (r *PlaylistRepo) FindByID(id uint64) (*PlaylistDBRow, error) {
	var row PlaylistDBRow
	err := r.db.Raw(`SELECT id, title FROM playlists WHERE id = ? AND deleted_at IS NULL`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// FindItems returns playlist items ordered by position.
func (r *PlaylistRepo) FindItems(playlistID uint64) ([]PlaylistItemDBRow, error) {
	var rows []PlaylistItemDBRow
	err := r.db.Raw(`
		SELECT playlist_id, video_id, position FROM playlist_items
		WHERE playlist_id = ? AND deleted_at IS NULL
		ORDER BY position
	`, playlistID).Scan(&rows).Error
	return rows, err
}
