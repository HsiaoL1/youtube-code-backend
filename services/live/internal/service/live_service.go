package service

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/live/internal/model"
	"youtube-code-backend/services/live/internal/repository"
)

type LiveService struct {
	roomRepo    *repository.RoomRepository
	sessionRepo *repository.SessionRepository
}

func NewLiveService(roomRepo *repository.RoomRepository, sessionRepo *repository.SessionRepository) *LiveService {
	return &LiveService{roomRepo: roomRepo, sessionRepo: sessionRepo}
}

// ── Request types ───────────────────────────────────────────────────────────

type CreateRoomRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type UpdateRoomRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type StreamAuthRequest struct {
	StreamKey string `json:"stream_key"`
}

// ── Room operations ─────────────────────────────────────────────────────────

func (s *LiveService) CreateRoom(userID uint64, req CreateRoomRequest) (*model.LiveRoom, error) {
	if req.Title == "" {
		return nil, errors.ErrValidation.WithMessage("title is required")
	}

	room := &model.LiveRoom{
		ChannelID:   userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Status:      model.RoomStatusIdle,
		StreamKey:   uuid.New().String(),
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create live room")
	}
	return room, nil
}

func (s *LiveService) GetRoom(id uint64) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}
	return room, nil
}

func (s *LiveService) UpdateRoom(id, userID uint64, req UpdateRoomRequest) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}

	if room.ChannelID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this live room")
	}

	if req.Title != "" {
		room.Title = req.Title
	}
	if req.Description != "" {
		room.Description = req.Description
	}
	if req.Category != "" {
		room.Category = req.Category
	}
	if req.ThumbnailURL != "" {
		room.ThumbnailURL = req.ThumbnailURL
	}

	if err := s.roomRepo.Update(room); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to update live room")
	}
	return room, nil
}

func (s *LiveService) DeleteRoom(id, userID uint64) error {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("live room not found")
		}
		return errors.ErrInternal
	}

	if room.ChannelID != userID {
		return errors.ErrForbidden.WithMessage("you do not own this live room")
	}

	if room.Status == model.RoomStatusLive {
		return errors.ErrBadRequest.WithMessage("cannot delete a room that is currently live")
	}

	if err := s.roomRepo.SoftDelete(id); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete live room")
	}
	return nil
}

func (s *LiveService) RegenerateStreamKey(id, userID uint64) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}

	if room.ChannelID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this live room")
	}

	if room.Status == model.RoomStatusLive {
		return nil, errors.ErrBadRequest.WithMessage("cannot regenerate stream key while live")
	}

	room.StreamKey = uuid.New().String()
	if err := s.roomRepo.Update(room); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to regenerate stream key")
	}
	return room, nil
}

// ── Live stream lifecycle ───────────────────────────────────────────────────

func (s *LiveService) GoLive(id, userID uint64) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}

	if room.ChannelID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this live room")
	}

	if !room.CanTransitionTo(model.RoomStatusLive) {
		return nil, errors.ErrBadRequest.WithMessage("room cannot go live from current status")
	}

	room.Status = model.RoomStatusLive
	room.ViewerCount = 0
	room.PeakViewerCount = 0

	if err := s.roomRepo.Update(room); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to start live stream")
	}

	// Create a new session
	session := &model.LiveSession{
		RoomID:    room.ID,
		StartedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create live session")
	}

	return room, nil
}

func (s *LiveService) EndStream(id, userID uint64) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}

	if room.ChannelID != userID {
		return nil, errors.ErrForbidden.WithMessage("you do not own this live room")
	}

	if !room.CanTransitionTo(model.RoomStatusEnded) {
		return nil, errors.ErrBadRequest.WithMessage("room is not currently live")
	}

	room.Status = model.RoomStatusEnded

	if err := s.roomRepo.Update(room); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to end live stream")
	}

	// Close the active session
	session, err := s.sessionRepo.FindActiveByRoomID(room.ID)
	if err == nil {
		now := time.Now()
		session.EndedAt = &now
		session.Duration = int64(now.Sub(session.StartedAt).Seconds())
		session.PeakViewers = room.PeakViewerCount
		_ = s.sessionRepo.Update(session)
	}

	return room, nil
}

// ── Playback & queries ──────────────────────────────────────────────────────

func (s *LiveService) GetPlaybackInfo(id uint64) (*model.LiveRoom, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live room not found")
		}
		return nil, errors.ErrInternal
	}
	return room, nil
}

func (s *LiveService) ListLiveRooms(offset, limit int) ([]model.LiveRoom, int64, error) {
	rooms, total, err := s.roomRepo.FindLiveRooms(offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list live rooms")
	}
	return rooms, total, nil
}

func (s *LiveService) ListChannelRooms(channelID uint64, offset, limit int) ([]model.LiveRoom, int64, error) {
	rooms, total, err := s.roomRepo.FindByChannelID(channelID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list channel rooms")
	}
	return rooms, total, nil
}

func (s *LiveService) StreamAuth(req StreamAuthRequest) (*model.LiveRoom, error) {
	if req.StreamKey == "" {
		return nil, errors.ErrValidation.WithMessage("stream_key is required")
	}

	room, err := s.roomRepo.FindByStreamKey(req.StreamKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrForbidden.WithMessage("invalid stream key")
		}
		return nil, errors.ErrInternal
	}
	return room, nil
}

func (s *LiveService) ListSessions(roomID uint64, offset, limit int) ([]model.LiveSession, int64, error) {
	sessions, total, err := s.sessionRepo.FindByRoomID(roomID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list sessions")
	}
	return sessions, total, nil
}
