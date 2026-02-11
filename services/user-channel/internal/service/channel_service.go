package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/user-channel/internal/model"
	"youtube-code-backend/services/user-channel/internal/repository"
)

type ChannelService struct {
	repo *repository.ChannelRepository
}

func NewChannelService(repo *repository.ChannelRepository) *ChannelService {
	return &ChannelService{repo: repo}
}

// CreateChannelRequest is the payload for creating a channel.
type CreateChannelRequest struct {
	Handle      string             `json:"handle"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Banner      string             `json:"banner"`
	Links       []ChannelLinkInput `json:"links"`
}

// UpdateChannelRequest is the payload for updating a channel.
type UpdateChannelRequest struct {
	Handle      *string             `json:"handle"`
	Name        *string             `json:"name"`
	Description *string             `json:"description"`
	Banner      *string             `json:"banner"`
	Links       *[]ChannelLinkInput `json:"links"`
}

// ChannelLinkInput represents a link in create/update requests.
type ChannelLinkInput struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ChannelStats holds channel statistics.
type ChannelStats struct {
	ChannelID       uint64 `json:"channel_id"`
	SubscriberCount int64  `json:"subscriber_count"`
	VideoCount      int64  `json:"video_count"`
}

func (s *ChannelService) Create(userID uint64, req CreateChannelRequest) (*model.Channel, error) {
	if req.Handle == "" || req.Name == "" {
		return nil, errors.ErrValidation.WithMessage("handle and name are required")
	}

	// Check handle uniqueness
	if _, err := s.repo.FindByHandle(req.Handle); err == nil {
		return nil, errors.ErrConflict.WithMessage("handle already taken")
	}

	channel := &model.Channel{
		UserID:      userID,
		Handle:      req.Handle,
		Name:        req.Name,
		Description: req.Description,
		Banner:      req.Banner,
	}

	if err := s.repo.Create(channel); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create channel")
	}

	// Create links if provided
	if len(req.Links) > 0 {
		links := make([]model.ChannelLink, len(req.Links))
		for i, l := range req.Links {
			links[i] = model.ChannelLink{
				ChannelID: channel.ID,
				Title:     l.Title,
				URL:       l.URL,
			}
		}
		if err := s.repo.ReplaceLinks(channel.ID, links); err != nil {
			return nil, errors.ErrInternal.WithMessage("failed to create channel links")
		}
	}

	// Re-fetch with links
	return s.repo.FindByID(channel.ID)
}

func (s *ChannelService) GetByID(id uint64) (*model.Channel, error) {
	channel, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("channel not found")
		}
		return nil, errors.ErrInternal
	}
	return channel, nil
}

func (s *ChannelService) GetByHandle(handle string) (*model.Channel, error) {
	channel, err := s.repo.FindByHandle(handle)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("channel not found")
		}
		return nil, errors.ErrInternal
	}
	return channel, nil
}

func (s *ChannelService) Update(userID, channelID uint64, req UpdateChannelRequest) (*model.Channel, error) {
	channel, err := s.repo.FindByID(channelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("channel not found")
		}
		return nil, errors.ErrInternal
	}

	// Ownership check
	if channel.UserID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this channel")
	}

	// Check handle uniqueness if changing
	if req.Handle != nil && *req.Handle != channel.Handle {
		if _, err := s.repo.FindByHandle(*req.Handle); err == nil {
			return nil, errors.ErrConflict.WithMessage("handle already taken")
		}
		channel.Handle = *req.Handle
	}

	if req.Name != nil {
		channel.Name = *req.Name
	}
	if req.Description != nil {
		channel.Description = *req.Description
	}
	if req.Banner != nil {
		channel.Banner = *req.Banner
	}

	if err := s.repo.Update(channel); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update channel")
	}

	// Replace links if provided
	if req.Links != nil {
		links := make([]model.ChannelLink, len(*req.Links))
		for i, l := range *req.Links {
			links[i] = model.ChannelLink{
				ChannelID: channel.ID,
				Title:     l.Title,
				URL:       l.URL,
			}
		}
		if err := s.repo.ReplaceLinks(channel.ID, links); err != nil {
			return nil, errors.ErrInternal.WithMessage("failed to update channel links")
		}
	}

	// Re-fetch with links
	return s.repo.FindByID(channel.ID)
}

func (s *ChannelService) GetStats(channelID uint64) (*ChannelStats, error) {
	channel, err := s.repo.FindByID(channelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("channel not found")
		}
		return nil, errors.ErrInternal
	}
	return &ChannelStats{
		ChannelID:       channel.ID,
		SubscriberCount: channel.SubscriberCount,
		VideoCount:      channel.VideoCount,
	}, nil
}
