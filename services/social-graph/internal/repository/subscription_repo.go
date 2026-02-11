package repository

import (
	"youtube-code-backend/services/social-graph/internal/model"

	"gorm.io/gorm"
)

// SubscriptionRepository handles database operations for subscriptions.
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create inserts a new subscription record.
func (r *SubscriptionRepository) Create(sub *model.Subscription) error {
	return r.db.Create(sub).Error
}

// Delete removes a subscription by subscriber ID and channel ID (soft delete).
func (r *SubscriptionRepository) Delete(subscriberID, channelID uint64) error {
	result := r.db.Where("subscriber_id = ? AND channel_id = ?", subscriberID, channelID).Delete(&model.Subscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindBySubscriberAndChannel finds a subscription by the composite key.
func (r *SubscriptionRepository) FindBySubscriberAndChannel(subscriberID, channelID uint64) (*model.Subscription, error) {
	var sub model.Subscription
	if err := r.db.Where("subscriber_id = ? AND channel_id = ?", subscriberID, channelID).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListBySubscriber returns paginated subscriptions for a given subscriber.
func (r *SubscriptionRepository) ListBySubscriber(subscriberID uint64, offset, limit int) ([]model.Subscription, int64, error) {
	var subs []model.Subscription
	var total int64

	query := r.db.Model(&model.Subscription{}).Where("subscriber_id = ?", subscriberID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

// ListByChannel returns paginated subscribers for a given channel.
func (r *SubscriptionRepository) ListByChannel(channelID uint64, offset, limit int) ([]model.Subscription, int64, error) {
	var subs []model.Subscription
	var total int64

	query := r.db.Model(&model.Subscription{}).Where("channel_id = ?", channelID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

// CountByChannel returns the total subscriber count for a channel.
func (r *SubscriptionRepository) CountByChannel(channelID uint64) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Subscription{}).Where("channel_id = ?", channelID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Update saves changes to an existing subscription.
func (r *SubscriptionRepository) Update(sub *model.Subscription) error {
	return r.db.Save(sub).Error
}

// GetFollowerIDs returns all subscriber IDs for a given channel.
func (r *SubscriptionRepository) GetFollowerIDs(channelID uint64) ([]uint64, error) {
	var ids []uint64
	if err := r.db.Model(&model.Subscription{}).
		Where("channel_id = ?", channelID).
		Pluck("subscriber_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
