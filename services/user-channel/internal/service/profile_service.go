package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/user-channel/internal/model"
	"youtube-code-backend/services/user-channel/internal/repository"
)

type ProfileService struct {
	repo *repository.ProfileRepository
}

func NewProfileService(repo *repository.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

// UpdateProfileRequest is the payload for updating a user profile.
type UpdateProfileRequest struct {
	Nickname      *string            `json:"nickname"`
	Avatar        *string            `json:"avatar"`
	Bio           *string            `json:"bio"`
	Region        *string            `json:"region"`
	Gender        *string            `json:"gender"`
	Links         *model.StringSlice `json:"links"`
	AccountStatus *string            `json:"account_status"`
}

func (s *ProfileService) GetByUserID(userID uint64) (*model.UserProfile, error) {
	profile, err := s.repo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("profile not found")
		}
		return nil, errors.ErrInternal
	}
	return profile, nil
}

func (s *ProfileService) UpdateProfile(userID uint64, req UpdateProfileRequest) (*model.UserProfile, error) {
	// Fetch existing or start fresh
	profile, err := s.repo.FindByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.ErrInternal
	}
	if profile == nil {
		profile = &model.UserProfile{
			UserID:        userID,
			AccountStatus: "active",
		}
	}

	// Apply partial updates
	if req.Nickname != nil {
		profile.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		profile.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		profile.Bio = *req.Bio
	}
	if req.Region != nil {
		profile.Region = *req.Region
	}
	if req.Gender != nil {
		profile.Gender = *req.Gender
	}
	if req.Links != nil {
		profile.Links = *req.Links
	}
	if req.AccountStatus != nil {
		profile.AccountStatus = *req.AccountStatus
	}

	if err := s.repo.Upsert(profile); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update profile")
	}
	return profile, nil
}
