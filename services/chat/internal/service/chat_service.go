package service

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/chat/internal/model"
	"youtube-code-backend/services/chat/internal/repository"
)

// ChatService handles chat business logic.
type ChatService struct {
	messageRepo *repository.MessageRepository
	configRepo  *repository.RoomConfigRepository
	managerRepo *repository.ManagerRepository
	muteRepo    *repository.MuteRepository
	filterRepo  *repository.KeywordFilterRepository
}

// NewChatService creates a new ChatService.
func NewChatService(
	messageRepo *repository.MessageRepository,
	configRepo *repository.RoomConfigRepository,
	managerRepo *repository.ManagerRepository,
	muteRepo *repository.MuteRepository,
	filterRepo *repository.KeywordFilterRepository,
) *ChatService {
	return &ChatService{
		messageRepo: messageRepo,
		configRepo:  configRepo,
		managerRepo: managerRepo,
		muteRepo:    muteRepo,
		filterRepo:  filterRepo,
	}
}

// --- Request / Response DTOs ---

// SendMessageRequest is the payload for sending a chat message.
type SendMessageRequest struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}

// UpdateRoomConfigRequest is the payload for updating room config.
type UpdateRoomConfigRequest struct {
	SlowModeSeconds  *int  `json:"slow_mode_seconds,omitempty"`
	SubscriberOnly   *bool `json:"subscriber_only,omitempty"`
	MaxMessageLength *int  `json:"max_message_length,omitempty"`
}

// AddManagerRequest is the payload for adding a chat manager.
type AddManagerRequest struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
}

// MuteUserRequest is the payload for muting a user.
type MuteUserRequest struct {
	UserID          uint64 `json:"user_id"`
	Reason          string `json:"reason"`
	DurationMinutes int    `json:"duration_minutes"`
}

// AddKeywordFilterRequest is the payload for adding a keyword filter.
type AddKeywordFilterRequest struct {
	Keyword string `json:"keyword"`
	Action  string `json:"action"`
}

// --- Messages ---

// GetMessages returns paginated message history for a room.
func (s *ChatService) GetMessages(roomID uint64, pg types.PaginationRequest) ([]model.ChatMessage, types.PaginationMeta, error) {
	pg.Normalize()

	messages, total, err := s.messageRepo.FindByRoomID(roomID, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to get messages")
	}

	meta := types.NewPaginationMeta(pg, total)
	return messages, meta, nil
}

// SendMessage stores a new chat message in a room.
func (s *ChatService) SendMessage(roomID, userID uint64, req SendMessageRequest) (*model.ChatMessage, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.ErrValidation.WithMessage("content is required")
	}

	// Check if user is muted
	_, err := s.muteRepo.FindActiveByRoomAndUser(roomID, userID)
	if err == nil {
		return nil, errors.ErrForbidden.WithMessage("you are muted in this room")
	}

	// Validate message type
	msgType := model.MessageType(req.Type)
	if msgType != model.MessageTypeText && msgType != model.MessageTypeSystem && msgType != model.MessageTypeDonation {
		msgType = model.MessageTypeText
	}

	// Check keyword filters
	filters, _ := s.filterRepo.ListByRoomID(roomID)
	contentLower := strings.ToLower(req.Content)
	for _, f := range filters {
		if strings.Contains(contentLower, strings.ToLower(f.Keyword)) {
			if f.Action == "block" {
				return nil, errors.ErrForbidden.WithMessage("message contains blocked keyword")
			}
		}
	}

	// Check room config for max message length
	cfg, err := s.configRepo.FindByRoomID(roomID)
	if err == nil && cfg.MaxMessageLength > 0 {
		if len(req.Content) > cfg.MaxMessageLength {
			return nil, errors.ErrValidation.WithMessage("message exceeds maximum length")
		}
	}

	msg := &model.ChatMessage{
		RoomID:  roomID,
		UserID:  userID,
		Type:    msgType,
		Content: req.Content,
		Status:  model.MessageStatusActive,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to send message")
	}

	return msg, nil
}

// --- Room Config ---

// GetRoomConfig returns the config for a chat room.
func (s *ChatService) GetRoomConfig(roomID uint64) (*model.ChatRoomConfig, error) {
	cfg, err := s.configRepo.FindByRoomID(roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config
			return &model.ChatRoomConfig{
				RoomID:           roomID,
				SlowModeSeconds:  0,
				SubscriberOnly:   false,
				MaxMessageLength: 200,
			}, nil
		}
		return nil, errors.ErrInternal.WithMessage("failed to get room config")
	}
	return cfg, nil
}

// UpdateRoomConfig updates the config for a chat room.
func (s *ChatService) UpdateRoomConfig(roomID, userID uint64, req UpdateRoomConfigRequest) (*model.ChatRoomConfig, error) {
	// Check if user is a manager of the room
	if !s.isManager(roomID, userID) {
		return nil, errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	cfg, err := s.configRepo.FindByRoomID(roomID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			cfg = &model.ChatRoomConfig{
				RoomID:           roomID,
				SlowModeSeconds:  0,
				SubscriberOnly:   false,
				MaxMessageLength: 200,
			}
		} else {
			return nil, errors.ErrInternal.WithMessage("failed to get room config")
		}
	}

	if req.SlowModeSeconds != nil {
		cfg.SlowModeSeconds = *req.SlowModeSeconds
	}
	if req.SubscriberOnly != nil {
		cfg.SubscriberOnly = *req.SubscriberOnly
	}
	if req.MaxMessageLength != nil {
		cfg.MaxMessageLength = *req.MaxMessageLength
	}

	if err := s.configRepo.Upsert(cfg); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update room config")
	}

	return cfg, nil
}

// --- Managers ---

// ListManagers returns all managers for a room.
func (s *ChatService) ListManagers(roomID uint64) ([]model.ChatManager, error) {
	managers, err := s.managerRepo.ListByRoomID(roomID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to list managers")
	}
	return managers, nil
}

// AddManager adds a new manager to a room.
func (s *ChatService) AddManager(roomID, userID uint64, req AddManagerRequest) (*model.ChatManager, error) {
	if req.UserID == 0 {
		return nil, errors.ErrValidation.WithMessage("user_id is required")
	}

	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return nil, errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	// Check if target user is already a manager
	_, err := s.managerRepo.FindByRoomAndUser(roomID, req.UserID)
	if err == nil {
		return nil, errors.ErrConflict.WithMessage("user is already a manager")
	}

	role := req.Role
	if role == "" {
		role = "moderator"
	}

	mgr := &model.ChatManager{
		RoomID: roomID,
		UserID: req.UserID,
		Role:   role,
	}

	if err := s.managerRepo.Create(mgr); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to add manager")
	}

	return mgr, nil
}

// RemoveManager removes a manager from a room.
func (s *ChatService) RemoveManager(roomID, userID, targetUserID uint64) error {
	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	if err := s.managerRepo.Delete(roomID, targetUserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("manager not found")
		}
		return errors.ErrInternal.WithMessage("failed to remove manager")
	}

	return nil
}

// --- Mutes ---

// MuteUser mutes a user in a room.
func (s *ChatService) MuteUser(roomID, userID uint64, req MuteUserRequest) (*model.ChatMute, error) {
	if req.UserID == 0 {
		return nil, errors.ErrValidation.WithMessage("user_id is required")
	}

	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return nil, errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	// Check if target user is already muted
	existing, err := s.muteRepo.FindActiveByRoomAndUser(roomID, req.UserID)
	if err == nil && existing != nil {
		return nil, errors.ErrConflict.WithMessage("user is already muted")
	}

	mute := &model.ChatMute{
		RoomID: roomID,
		UserID: req.UserID,
		Reason: req.Reason,
	}

	if req.DurationMinutes > 0 {
		expires := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)
		mute.ExpiresAt = &expires
	}

	if err := s.muteRepo.Create(mute); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to mute user")
	}

	return mute, nil
}

// UnmuteUser removes a mute for a user in a room.
func (s *ChatService) UnmuteUser(roomID, userID, targetUserID uint64) error {
	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	if err := s.muteRepo.DeleteByRoomAndUser(roomID, targetUserID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("mute not found")
		}
		return errors.ErrInternal.WithMessage("failed to unmute user")
	}

	return nil
}

// --- Keyword Filters ---

// ListFilters returns all keyword filters for a room.
func (s *ChatService) ListFilters(roomID uint64) ([]model.ChatKeywordFilter, error) {
	filters, err := s.filterRepo.ListByRoomID(roomID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to list filters")
	}
	return filters, nil
}

// AddFilter adds a new keyword filter to a room.
func (s *ChatService) AddFilter(roomID, userID uint64, req AddKeywordFilterRequest) (*model.ChatKeywordFilter, error) {
	if strings.TrimSpace(req.Keyword) == "" {
		return nil, errors.ErrValidation.WithMessage("keyword is required")
	}

	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return nil, errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	action := req.Action
	if action == "" {
		action = "block"
	}

	filter := &model.ChatKeywordFilter{
		RoomID:  roomID,
		Keyword: req.Keyword,
		Action:  action,
	}

	if err := s.filterRepo.Create(filter); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to add filter")
	}

	return filter, nil
}

// RemoveFilter removes a keyword filter by ID.
func (s *ChatService) RemoveFilter(roomID, userID, filterID uint64) error {
	// Check if requester is a manager
	if !s.isManager(roomID, userID) {
		return errors.ErrForbidden.WithMessage("you are not a manager of this room")
	}

	// Verify the filter belongs to this room
	filter, err := s.filterRepo.FindByID(filterID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("filter not found")
		}
		return errors.ErrInternal.WithMessage("failed to find filter")
	}
	if filter.RoomID != roomID {
		return errors.ErrForbidden.WithMessage("filter does not belong to this room")
	}

	if err := s.filterRepo.Delete(filterID); err != nil {
		return errors.ErrInternal.WithMessage("failed to remove filter")
	}

	return nil
}

// --- Helpers ---

// isManager checks if a user is a manager of the given room.
func (s *ChatService) isManager(roomID, userID uint64) bool {
	_, err := s.managerRepo.FindByRoomAndUser(roomID, userID)
	return err == nil
}
