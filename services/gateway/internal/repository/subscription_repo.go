package repository

import (
	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// FindChannelIDsBySubscriber returns channel IDs that a user is subscribed to.
func (r *SubscriptionRepo) FindChannelIDsBySubscriber(subscriberID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Raw(`
		SELECT channel_id FROM subscriptions
		WHERE subscriber_id = ? AND deleted_at IS NULL
	`, subscriberID).Scan(&ids).Error
	return ids, err
}

// GetSubscriberCount returns the subscriber count for a channel.
func (r *SubscriptionRepo) GetSubscriberCount(channelID uint64) (int64, error) {
	var count int64
	err := r.db.Raw(`SELECT COALESCE(subscriber_count, 0) FROM channels WHERE id = ? AND deleted_at IS NULL`, channelID).Scan(&count).Error
	return count, err
}
