package service

import (
	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/search/internal/model"
	"youtube-code-backend/services/search/internal/repository"

	"gorm.io/gorm"
)

// SearchService handles search business logic.
type SearchService struct {
	videoRepo   *repository.VideoRepository
	channelRepo *repository.ChannelRepository
	liveRepo    *repository.LiveRepository
}

// NewSearchService creates a new SearchService.
func NewSearchService(
	videoRepo *repository.VideoRepository,
	channelRepo *repository.ChannelRepository,
	liveRepo *repository.LiveRepository,
) *SearchService {
	return &SearchService{
		videoRepo:   videoRepo,
		channelRepo: channelRepo,
		liveRepo:    liveRepo,
	}
}

// UpsertVideoRequest is the payload for upserting a video index entry.
type UpsertVideoRequest struct {
	VideoID     uint64 `json:"video_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	ChannelName string `json:"channel_name"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
}

// UpsertChannelRequest is the payload for upserting a channel index entry.
type UpsertChannelRequest struct {
	ChannelID   uint64 `json:"channel_id" validate:"required"`
	Handle      string `json:"handle"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// UpsertLiveRequest is the payload for upserting a live room index entry.
type UpsertLiveRequest struct {
	RoomID      uint64 `json:"room_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	ChannelName string `json:"channel_name"`
	Category    string `json:"category"`
	IsLive      bool   `json:"is_live"`
}

// CombinedSearchResult holds results from a combined search across all types.
type CombinedSearchResult struct {
	Videos   []model.SearchVideo   `json:"videos"`
	Channels []model.SearchChannel `json:"channels"`
	Live     []model.SearchLive    `json:"live"`
}

// SearchVideos performs a paginated full-text search on videos.
func (s *SearchService) SearchVideos(query, category, sort string, pg types.PaginationRequest) ([]model.SearchVideo, types.PaginationMeta, error) {
	pg.Normalize()

	videos, total, err := s.videoRepo.Search(query, category, sort, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to search videos")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// SearchChannels performs a paginated full-text search on channels.
func (s *SearchService) SearchChannels(query string, pg types.PaginationRequest) ([]model.SearchChannel, types.PaginationMeta, error) {
	pg.Normalize()

	channels, total, err := s.channelRepo.Search(query, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to search channels")
	}

	meta := types.NewPaginationMeta(pg, total)
	return channels, meta, nil
}

// SearchLive performs a paginated full-text search on live rooms.
func (s *SearchService) SearchLive(query string, pg types.PaginationRequest) ([]model.SearchLive, types.PaginationMeta, error) {
	pg.Normalize()

	rooms, total, err := s.liveRepo.Search(query, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to search live rooms")
	}

	meta := types.NewPaginationMeta(pg, total)
	return rooms, meta, nil
}

// SearchAll performs a combined search across videos, channels, and live rooms returning top 5 each.
func (s *SearchService) SearchAll(query string) (*CombinedSearchResult, error) {
	if query == "" {
		return nil, errors.ErrBadRequest.WithMessage("query parameter 'q' is required")
	}

	videos, err := s.videoRepo.SearchTop(query, 5)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to search videos")
	}

	channels, err := s.channelRepo.SearchTop(query, 5)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to search channels")
	}

	live, err := s.liveRepo.SearchTop(query, 5)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to search live rooms")
	}

	return &CombinedSearchResult{
		Videos:   videos,
		Channels: channels,
		Live:     live,
	}, nil
}

// UpsertVideo upserts a video index entry.
func (s *SearchService) UpsertVideo(req UpsertVideoRequest) (*model.SearchVideo, error) {
	if req.VideoID == 0 {
		return nil, errors.ErrBadRequest.WithMessage("video_id is required")
	}
	if req.Title == "" {
		return nil, errors.ErrBadRequest.WithMessage("title is required")
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = "public"
	}
	status := req.Status
	if status == "" {
		status = "ready"
	}

	v := &model.SearchVideo{
		VideoID:     req.VideoID,
		Title:       req.Title,
		Description: req.Description,
		ChannelName: req.ChannelName,
		Category:    req.Category,
		Tags:        req.Tags,
		Visibility:  visibility,
		Status:      status,
	}

	if err := s.videoRepo.Upsert(v); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to upsert video index")
	}

	return v, nil
}

// DeleteVideo removes a video from the search index.
func (s *SearchService) DeleteVideo(videoID uint64) error {
	if err := s.videoRepo.DeleteByVideoID(videoID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("video not found in search index")
		}
		return errors.ErrInternal.WithMessage("failed to delete video from index")
	}
	return nil
}

// UpsertChannel upserts a channel index entry.
func (s *SearchService) UpsertChannel(req UpsertChannelRequest) (*model.SearchChannel, error) {
	if req.ChannelID == 0 {
		return nil, errors.ErrBadRequest.WithMessage("channel_id is required")
	}
	if req.Name == "" {
		return nil, errors.ErrBadRequest.WithMessage("name is required")
	}

	ch := &model.SearchChannel{
		ChannelID:   req.ChannelID,
		Handle:      req.Handle,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.channelRepo.Upsert(ch); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to upsert channel index")
	}

	return ch, nil
}

// UpsertLive upserts a live room index entry.
func (s *SearchService) UpsertLive(req UpsertLiveRequest) (*model.SearchLive, error) {
	if req.RoomID == 0 {
		return nil, errors.ErrBadRequest.WithMessage("room_id is required")
	}
	if req.Title == "" {
		return nil, errors.ErrBadRequest.WithMessage("title is required")
	}

	live := &model.SearchLive{
		RoomID:      req.RoomID,
		Title:       req.Title,
		ChannelName: req.ChannelName,
		Category:    req.Category,
		IsLive:      req.IsLive,
	}

	if err := s.liveRepo.Upsert(live); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to upsert live room index")
	}

	return live, nil
}
