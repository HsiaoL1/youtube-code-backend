package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/video/internal/model"
	"youtube-code-backend/services/video/internal/repository"
)

type PlaylistService struct {
	playlistRepo *repository.PlaylistRepository
	videoRepo    *repository.VideoRepository
}

func NewPlaylistService(
	playlistRepo *repository.PlaylistRepository,
	videoRepo *repository.VideoRepository,
) *PlaylistService {
	return &PlaylistService{
		playlistRepo: playlistRepo,
		videoRepo:    videoRepo,
	}
}

// --- Request DTOs ---

type CreatePlaylistRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type UpdatePlaylistRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

type AddPlaylistItemRequest struct {
	VideoID uint64 `json:"video_id"`
}

// --- Playlist CRUD ---

func (s *PlaylistService) Create(userID uint64, req CreatePlaylistRequest) (*model.Playlist, error) {
	if req.Title == "" {
		return nil, errors.ErrValidation.WithMessage("title is required")
	}

	visibility := req.Visibility
	if visibility != "public" && visibility != "private" && visibility != "unlisted" {
		visibility = "private"
	}

	playlist := &model.Playlist{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Visibility:  visibility,
	}

	if err := s.playlistRepo.Create(playlist); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create playlist")
	}
	return playlist, nil
}

func (s *PlaylistService) GetByID(id uint64) (*model.Playlist, error) {
	playlist, err := s.playlistRepo.FindByIDWithItems(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("playlist not found")
		}
		return nil, errors.ErrInternal
	}
	return playlist, nil
}

func (s *PlaylistService) Update(id, userID uint64, req UpdatePlaylistRequest) (*model.Playlist, error) {
	playlist, err := s.playlistRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("playlist not found")
		}
		return nil, errors.ErrInternal
	}

	if playlist.UserID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this playlist")
	}

	if req.Title != nil {
		playlist.Title = *req.Title
	}
	if req.Description != nil {
		playlist.Description = *req.Description
	}
	if req.Visibility != nil {
		vis := *req.Visibility
		if vis == "public" || vis == "private" || vis == "unlisted" {
			playlist.Visibility = vis
		}
	}

	if err := s.playlistRepo.Update(playlist); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update playlist")
	}
	return playlist, nil
}

func (s *PlaylistService) Delete(id, userID uint64) error {
	playlist, err := s.playlistRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("playlist not found")
		}
		return errors.ErrInternal
	}

	if playlist.UserID != userID {
		return errors.ErrForbidden.WithMessage("you do not own this playlist")
	}

	if err := s.playlistRepo.Delete(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete playlist")
	}
	return nil
}

func (s *PlaylistService) AddItem(playlistID, userID uint64, req AddPlaylistItemRequest) (*model.PlaylistItem, error) {
	playlist, err := s.playlistRepo.FindByID(playlistID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("playlist not found")
		}
		return nil, errors.ErrInternal
	}

	if playlist.UserID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this playlist")
	}

	// Verify video exists
	if _, err := s.videoRepo.FindByID(req.VideoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("video not found")
		}
		return nil, errors.ErrInternal
	}

	maxPos, err := s.playlistRepo.GetMaxPosition(playlistID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to determine position")
	}

	item := &model.PlaylistItem{
		PlaylistID: playlistID,
		VideoID:    req.VideoID,
		Position:   maxPos + 1,
	}

	if err := s.playlistRepo.AddItem(item); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to add item to playlist")
	}

	_ = s.playlistRepo.IncrementVideoCount(playlistID, 1)
	return item, nil
}

func (s *PlaylistService) RemoveItem(playlistID, videoID, userID uint64) error {
	playlist, err := s.playlistRepo.FindByID(playlistID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("playlist not found")
		}
		return errors.ErrInternal
	}

	if playlist.UserID != userID {
		return errors.ErrForbidden.WithMessage("you do not own this playlist")
	}

	if err := s.playlistRepo.RemoveItem(playlistID, videoID); err != nil {
		return errors.ErrInternal.WithMessage("failed to remove item from playlist")
	}

	_ = s.playlistRepo.IncrementVideoCount(playlistID, -1)
	return nil
}
