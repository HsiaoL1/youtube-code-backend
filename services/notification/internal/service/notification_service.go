package service

import (
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/notification/internal/model"
	"youtube-code-backend/services/notification/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
	prefRepo  *repository.PreferenceRepository
}

func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	prefRepo *repository.PreferenceRepository,
) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		prefRepo:  prefRepo,
	}
}

// --- Request DTOs ---

type SendNotificationRequest struct {
	UserID       uint64 `json:"user_id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	ActorID      uint64 `json:"actor_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   uint64 `json:"resource_id"`
}

type BroadcastNotificationRequest struct {
	UserIDs      []uint64 `json:"user_ids"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	ActorID      uint64   `json:"actor_id"`
	ResourceType string   `json:"resource_type"`
	ResourceID   uint64   `json:"resource_id"`
}

type UpdatePreferencesRequest struct {
	NewVideo            *bool `json:"new_video,omitempty"`
	NewLive             *bool `json:"new_live,omitempty"`
	CommentReply        *bool `json:"comment_reply,omitempty"`
	Subscription        *bool `json:"subscription,omitempty"`
	Likes               *bool `json:"likes,omitempty"`
	SystemNotifications *bool `json:"system_notifications,omitempty"`
}

// --- Notification operations ---

func (s *NotificationService) List(userID uint64, offset, limit int) ([]model.Notification, int64, error) {
	notifications, total, err := s.notifRepo.FindByUserID(userID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list notifications")
	}
	return notifications, total, nil
}

func (s *NotificationService) UnreadCount(userID uint64) (int64, error) {
	count, err := s.notifRepo.CountUnread(userID)
	if err != nil {
		return 0, errors.ErrInternal.WithMessage("failed to count unread notifications")
	}
	return count, nil
}

func (s *NotificationService) MarkAsRead(id, userID uint64) error {
	n, err := s.notifRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("notification not found")
		}
		return errors.ErrInternal
	}

	if n.UserID != userID {
		return errors.ErrForbidden.WithMessage("you do not own this notification")
	}

	if err := s.notifRepo.MarkAsRead(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to mark notification as read")
	}
	return nil
}

func (s *NotificationService) MarkAllAsRead(userID uint64) error {
	if err := s.notifRepo.MarkAllAsRead(userID); err != nil {
		return errors.ErrInternal.WithMessage("failed to mark all notifications as read")
	}
	return nil
}

func (s *NotificationService) Delete(id, userID uint64) error {
	n, err := s.notifRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("notification not found")
		}
		return errors.ErrInternal
	}

	if n.UserID != userID {
		return errors.ErrForbidden.WithMessage("you do not own this notification")
	}

	if err := s.notifRepo.Delete(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete notification")
	}
	return nil
}

// --- Send / Broadcast ---

func (s *NotificationService) Send(req SendNotificationRequest) (*model.Notification, error) {
	if req.UserID == 0 {
		return nil, errors.ErrValidation.WithMessage("user_id is required")
	}
	if req.Title == "" {
		return nil, errors.ErrValidation.WithMessage("title is required")
	}
	if req.Type == "" {
		return nil, errors.ErrValidation.WithMessage("type is required")
	}

	n := &model.Notification{
		UserID:       req.UserID,
		Type:         model.NotificationType(req.Type),
		Title:        req.Title,
		Body:         req.Body,
		ActorID:      req.ActorID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}

	if err := s.notifRepo.Create(n); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to send notification")
	}
	return n, nil
}

func (s *NotificationService) Broadcast(req BroadcastNotificationRequest) (int, error) {
	if len(req.UserIDs) == 0 {
		return 0, errors.ErrValidation.WithMessage("user_ids is required")
	}
	if req.Title == "" {
		return 0, errors.ErrValidation.WithMessage("title is required")
	}
	if req.Type == "" {
		return 0, errors.ErrValidation.WithMessage("type is required")
	}

	notifications := make([]model.Notification, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		notifications = append(notifications, model.Notification{
			UserID:       uid,
			Type:         model.NotificationType(req.Type),
			Title:        req.Title,
			Body:         req.Body,
			ActorID:      req.ActorID,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
		})
	}

	if err := s.notifRepo.CreateBatch(notifications); err != nil {
		return 0, errors.ErrInternal.WithMessage("failed to broadcast notifications")
	}
	return len(notifications), nil
}

// --- Preferences ---

func (s *NotificationService) GetPreferences(userID uint64) (*model.NotificationPreference, error) {
	pref, err := s.prefRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default preferences for new users
			defaultPref := &model.NotificationPreference{
				UserID:              userID,
				NewVideo:            true,
				NewLive:             true,
				CommentReply:        true,
				Subscription:        true,
				Likes:               true,
				SystemNotifications: true,
			}
			if err := s.prefRepo.Upsert(defaultPref); err != nil {
				return nil, errors.ErrInternal.WithMessage("failed to create default preferences")
			}
			return defaultPref, nil
		}
		return nil, errors.ErrInternal.WithMessage("failed to get preferences")
	}
	return pref, nil
}

func (s *NotificationService) UpdatePreferences(userID uint64, req UpdatePreferencesRequest) (*model.NotificationPreference, error) {
	pref, err := s.prefRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new preferences with defaults, then apply updates
			pref = &model.NotificationPreference{
				UserID:              userID,
				NewVideo:            true,
				NewLive:             true,
				CommentReply:        true,
				Subscription:        true,
				Likes:               true,
				SystemNotifications: true,
			}
		} else {
			return nil, errors.ErrInternal.WithMessage("failed to get preferences")
		}
	}

	if req.NewVideo != nil {
		pref.NewVideo = *req.NewVideo
	}
	if req.NewLive != nil {
		pref.NewLive = *req.NewLive
	}
	if req.CommentReply != nil {
		pref.CommentReply = *req.CommentReply
	}
	if req.Subscription != nil {
		pref.Subscription = *req.Subscription
	}
	if req.Likes != nil {
		pref.Likes = *req.Likes
	}
	if req.SystemNotifications != nil {
		pref.SystemNotifications = *req.SystemNotifications
	}

	if err := s.prefRepo.Upsert(pref); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update preferences")
	}
	return pref, nil
}
