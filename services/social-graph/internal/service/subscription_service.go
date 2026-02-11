package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/social-graph/internal/model"
	"youtube-code-backend/services/social-graph/internal/repository"
)

// SubscriptionService handles subscription business logic.
type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// SubscribeRequest is the payload for subscribing to a channel.
type SubscribeRequest struct {
	ChannelID        uint64 `json:"channel_id" validate:"required"`
	NotifyPreference string `json:"notify_preference"`
}

// UpdateNotifyRequest is the payload for updating notification preferences.
type UpdateNotifyRequest struct {
	NotifyPreference string `json:"notify_preference" validate:"required"`
}

// Subscribe creates a new subscription for the given user and channel.
func (s *SubscriptionService) Subscribe(subscriberID uint64, req SubscribeRequest) (*model.Subscription, error) {
	if req.ChannelID == 0 {
		return nil, errors.ErrBadRequest.WithMessage("channel_id is required")
	}

	if subscriberID == req.ChannelID {
		return nil, errors.ErrBadRequest.WithMessage("cannot subscribe to yourself")
	}

	// Check if already subscribed
	existing, err := s.repo.FindBySubscriberAndChannel(subscriberID, req.ChannelID)
	if err == nil && existing != nil {
		return nil, errors.ErrConflict.WithMessage("already subscribed to this channel")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.ErrInternal.WithMessage("failed to check subscription")
	}

	notifyPref := req.NotifyPreference
	if notifyPref == "" {
		notifyPref = "all"
	}

	sub := &model.Subscription{
		SubscriberID:     subscriberID,
		ChannelID:        req.ChannelID,
		NotifyPreference: notifyPref,
	}

	if err := s.repo.Create(sub); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create subscription")
	}

	return sub, nil
}

// Unsubscribe removes a subscription for the given user and channel.
func (s *SubscriptionService) Unsubscribe(subscriberID, channelID uint64) error {
	if err := s.repo.Delete(subscriberID, channelID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("subscription not found")
		}
		return errors.ErrInternal.WithMessage("failed to delete subscription")
	}
	return nil
}

// ListSubscriptions returns paginated subscriptions for a subscriber.
func (s *SubscriptionService) ListSubscriptions(subscriberID uint64, pg types.PaginationRequest) ([]model.Subscription, types.PaginationMeta, error) {
	pg.Normalize()

	subs, total, err := s.repo.ListBySubscriber(subscriberID, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to list subscriptions")
	}

	meta := types.NewPaginationMeta(pg, total)
	return subs, meta, nil
}

// ListSubscribers returns paginated subscribers for a channel.
func (s *SubscriptionService) ListSubscribers(channelID uint64, pg types.PaginationRequest) ([]model.Subscription, types.PaginationMeta, error) {
	pg.Normalize()

	subs, total, err := s.repo.ListByChannel(channelID, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to list subscribers")
	}

	meta := types.NewPaginationMeta(pg, total)
	return subs, meta, nil
}

// GetSubscriberCount returns the total subscriber count for a channel.
func (s *SubscriptionService) GetSubscriberCount(channelID uint64) (int64, error) {
	count, err := s.repo.CountByChannel(channelID)
	if err != nil {
		return 0, errors.ErrInternal.WithMessage("failed to count subscribers")
	}
	return count, nil
}

// CheckSubscription checks if a user is subscribed to a channel.
func (s *SubscriptionService) CheckSubscription(subscriberID, channelID uint64) (*model.Subscription, bool, error) {
	sub, err := s.repo.FindBySubscriberAndChannel(subscriberID, channelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, errors.ErrInternal.WithMessage("failed to check subscription")
	}
	return sub, true, nil
}

// UpdateNotifyPreference updates the notification preference for a subscription.
func (s *SubscriptionService) UpdateNotifyPreference(subscriberID, channelID uint64, req UpdateNotifyRequest) (*model.Subscription, error) {
	if req.NotifyPreference == "" {
		return nil, errors.ErrBadRequest.WithMessage("notify_preference is required")
	}

	sub, err := s.repo.FindBySubscriberAndChannel(subscriberID, channelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("subscription not found")
		}
		return nil, errors.ErrInternal.WithMessage("failed to find subscription")
	}

	sub.NotifyPreference = req.NotifyPreference
	if err := s.repo.Update(sub); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update notification preference")
	}

	return sub, nil
}

// GetFollowerIDs returns all subscriber IDs for a channel (internal endpoint).
func (s *SubscriptionService) GetFollowerIDs(channelID uint64) ([]uint64, error) {
	ids, err := s.repo.GetFollowerIDs(channelID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to get follower IDs")
	}
	return ids, nil
}
