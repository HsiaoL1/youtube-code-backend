package repository

import (
	"youtube-code-backend/services/user-channel/internal/model"

	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

func (r *ChannelRepository) FindByID(id uint64) (*model.Channel, error) {
	var channel model.Channel
	if err := r.db.Preload("Links").First(&channel, id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) FindByHandle(handle string) (*model.Channel, error) {
	var channel model.Channel
	if err := r.db.Preload("Links").Where("handle = ?", handle).First(&channel).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

func (r *ChannelRepository) ReplaceLinks(channelID uint64, links []model.ChannelLink) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("channel_id = ?", channelID).Delete(&model.ChannelLink{}).Error; err != nil {
			return err
		}
		if len(links) > 0 {
			for i := range links {
				links[i].ChannelID = channelID
			}
			return tx.Create(&links).Error
		}
		return nil
	})
}
