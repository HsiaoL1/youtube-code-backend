package service

import (
	"time"

	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/video/internal/model"
	"youtube-code-backend/services/video/internal/repository"
)

type VideoService struct {
	videoRepo *repository.VideoRepository
	likeRepo  *repository.LikeRepository
	favRepo   *repository.FavoriteRepository
	watchRepo *repository.WatchRepository
}

func NewVideoService(
	videoRepo *repository.VideoRepository,
	likeRepo *repository.LikeRepository,
	favRepo *repository.FavoriteRepository,
	watchRepo *repository.WatchRepository,
) *VideoService {
	return &VideoService{
		videoRepo: videoRepo,
		likeRepo:  likeRepo,
		favRepo:   favRepo,
		watchRepo: watchRepo,
	}
}

// --- Request / Response DTOs ---

type CreateVideoRequest struct {
	ChannelID    uint64     `json:"channel_id"`
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Visibility   string     `json:"visibility"`
	ThumbnailURL string     `json:"thumbnail_url"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
}

type UpdateVideoRequest struct {
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Visibility   *string    `json:"visibility,omitempty"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
}

type UpdateProgressRequest struct {
	Progress int64 `json:"progress"`
	Duration int64 `json:"duration"`
}

// --- Video CRUD ---

func (s *VideoService) Create(req CreateVideoRequest) (*model.Video, error) {
	if req.Title == "" {
		return nil, errors.ErrValidation.WithMessage("title is required")
	}
	if req.ChannelID == 0 {
		return nil, errors.ErrValidation.WithMessage("channel_id is required")
	}

	videoType := model.VideoType(req.Type)
	if videoType != model.VideoTypeVideo && videoType != model.VideoTypeShort {
		videoType = model.VideoTypeVideo
	}

	visibility := model.VideoVisibility(req.Visibility)
	if visibility != model.VideoVisibilityPublic &&
		visibility != model.VideoVisibilityPrivate &&
		visibility != model.VideoVisibilityUnlisted {
		visibility = model.VideoVisibilityPrivate
	}

	video := &model.Video{
		ChannelID:    req.ChannelID,
		Type:         videoType,
		Title:        req.Title,
		Description:  req.Description,
		Status:       model.VideoStatusDraft,
		Visibility:   visibility,
		ThumbnailURL: req.ThumbnailURL,
		ScheduledAt:  req.ScheduledAt,
	}

	if err := s.videoRepo.Create(video); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create video")
	}
	return video, nil
}

func (s *VideoService) GetByID(id uint64) (*model.Video, error) {
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("video not found")
		}
		return nil, errors.ErrInternal
	}
	return video, nil
}

func (s *VideoService) Update(id uint64, ownerChannelID uint64, req UpdateVideoRequest) (*model.Video, error) {
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("video not found")
		}
		return nil, errors.ErrInternal
	}

	if video.ChannelID != ownerChannelID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this video")
	}

	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if req.Visibility != nil {
		vis := model.VideoVisibility(*req.Visibility)
		if vis == model.VideoVisibilityPublic ||
			vis == model.VideoVisibilityPrivate ||
			vis == model.VideoVisibilityUnlisted {
			video.Visibility = vis
		}
	}
	if req.ThumbnailURL != nil {
		video.ThumbnailURL = *req.ThumbnailURL
	}
	if req.ScheduledAt != nil {
		video.ScheduledAt = req.ScheduledAt
	}

	if err := s.videoRepo.Update(video); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update video")
	}
	return video, nil
}

func (s *VideoService) Delete(id uint64, ownerChannelID uint64) error {
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("video not found")
		}
		return errors.ErrInternal
	}

	if video.ChannelID != ownerChannelID {
		return errors.ErrForbidden.WithMessage("you do not own this video")
	}

	if err := s.videoRepo.SoftDelete(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete video")
	}
	return nil
}

func (s *VideoService) Publish(id uint64, ownerChannelID uint64) (*model.Video, error) {
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("video not found")
		}
		return nil, errors.ErrInternal
	}

	if video.ChannelID != ownerChannelID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this video")
	}

	// draft -> processing
	if !video.CanTransitionTo(model.VideoStatusProcessing) {
		return nil, errors.ErrBadRequest.WithMessage("video cannot be published from status: " + string(video.Status))
	}

	video.Status = model.VideoStatusProcessing
	if err := s.videoRepo.Update(video); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to publish video")
	}
	return video, nil
}

// --- Likes ---

func (s *VideoService) LikeVideo(videoID, userID uint64) error {
	if _, err := s.videoRepo.FindByID(videoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("video not found")
		}
		return errors.ErrInternal
	}

	existing, err := s.likeRepo.FindByVideoAndUser(videoID, userID)
	if err == nil {
		// Already has a record
		if existing.IsLike {
			return nil // already liked
		}
		// Was a dislike, switch to like
		existing.IsLike = true
		if err := s.likeRepo.Upsert(existing); err != nil {
			return errors.ErrInternal.WithMessage("failed to like video")
		}
		_ = s.videoRepo.IncrementCount(videoID, "like_count", 1)
		_ = s.videoRepo.IncrementCount(videoID, "dislike_count", -1)
		return nil
	}

	like := &model.VideoLike{
		VideoID: videoID,
		UserID:  userID,
		IsLike:  true,
	}
	if err := s.likeRepo.Upsert(like); err != nil {
		return errors.ErrInternal.WithMessage("failed to like video")
	}
	_ = s.videoRepo.IncrementCount(videoID, "like_count", 1)
	return nil
}

func (s *VideoService) DislikeVideo(videoID, userID uint64) error {
	if _, err := s.videoRepo.FindByID(videoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("video not found")
		}
		return errors.ErrInternal
	}

	existing, err := s.likeRepo.FindByVideoAndUser(videoID, userID)
	if err == nil {
		if !existing.IsLike {
			return nil // already disliked
		}
		existing.IsLike = false
		if err := s.likeRepo.Upsert(existing); err != nil {
			return errors.ErrInternal.WithMessage("failed to dislike video")
		}
		_ = s.videoRepo.IncrementCount(videoID, "dislike_count", 1)
		_ = s.videoRepo.IncrementCount(videoID, "like_count", -1)
		return nil
	}

	dislike := &model.VideoLike{
		VideoID: videoID,
		UserID:  userID,
		IsLike:  false,
	}
	if err := s.likeRepo.Upsert(dislike); err != nil {
		return errors.ErrInternal.WithMessage("failed to dislike video")
	}
	_ = s.videoRepo.IncrementCount(videoID, "dislike_count", 1)
	return nil
}

func (s *VideoService) RemoveLike(videoID, userID uint64) error {
	existing, err := s.likeRepo.FindByVideoAndUser(videoID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // nothing to remove
		}
		return errors.ErrInternal
	}

	if err := s.likeRepo.Delete(videoID, userID); err != nil {
		return errors.ErrInternal.WithMessage("failed to remove like")
	}

	if existing.IsLike {
		_ = s.videoRepo.IncrementCount(videoID, "like_count", -1)
	} else {
		_ = s.videoRepo.IncrementCount(videoID, "dislike_count", -1)
	}
	return nil
}

// --- Favorites ---

func (s *VideoService) AddFavorite(videoID, userID uint64) error {
	if _, err := s.videoRepo.FindByID(videoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("video not found")
		}
		return errors.ErrInternal
	}

	if _, err := s.favRepo.FindByVideoAndUser(videoID, userID); err == nil {
		return errors.ErrConflict.WithMessage("video already in favorites")
	}

	fav := &model.VideoFavorite{
		VideoID: videoID,
		UserID:  userID,
	}
	if err := s.favRepo.Create(fav); err != nil {
		return errors.ErrInternal.WithMessage("failed to add favorite")
	}
	_ = s.videoRepo.IncrementCount(videoID, "favorite_count", 1)
	return nil
}

func (s *VideoService) RemoveFavorite(videoID, userID uint64) error {
	if _, err := s.favRepo.FindByVideoAndUser(videoID, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("favorite not found")
		}
		return errors.ErrInternal
	}

	if err := s.favRepo.Delete(videoID, userID); err != nil {
		return errors.ErrInternal.WithMessage("failed to remove favorite")
	}
	_ = s.videoRepo.IncrementCount(videoID, "favorite_count", -1)
	return nil
}

// --- Watch Progress / History ---

func (s *VideoService) GetProgress(videoID, userID uint64) (*model.WatchHistory, error) {
	wh, err := s.watchRepo.FindByUserAndVideo(userID, videoID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("no watch progress found")
		}
		return nil, errors.ErrInternal
	}
	return wh, nil
}

func (s *VideoService) UpdateProgress(videoID, userID uint64, req UpdateProgressRequest) (*model.WatchHistory, error) {
	wh := &model.WatchHistory{
		UserID:   userID,
		VideoID:  videoID,
		Progress: req.Progress,
		Duration: req.Duration,
	}
	if err := s.watchRepo.Upsert(wh); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update progress")
	}
	return wh, nil
}

func (s *VideoService) GetWatchHistory(userID uint64, offset, limit int) ([]model.WatchHistory, int64, error) {
	items, total, err := s.watchRepo.FindByUser(userID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to get watch history")
	}
	return items, total, nil
}
