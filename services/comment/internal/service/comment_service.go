package service

import (
	"strings"

	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/comment/internal/model"
	"youtube-code-backend/services/comment/internal/repository"
)

type CommentService struct {
	commentRepo       *repository.CommentRepository
	likeRepo          *repository.CommentLikeRepository
	sensitiveWordRepo *repository.SensitiveWordRepository
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	likeRepo *repository.CommentLikeRepository,
	sensitiveWordRepo *repository.SensitiveWordRepository,
) *CommentService {
	return &CommentService{
		commentRepo:       commentRepo,
		likeRepo:          likeRepo,
		sensitiveWordRepo: sensitiveWordRepo,
	}
}

// CreateCommentRequest is the payload for creating a comment.
type CreateCommentRequest struct {
	VideoID  uint64  `json:"video_id"`
	Content  string  `json:"content"`
	ParentID *uint64 `json:"parent_id,omitempty"`
}

// ReportCommentRequest is the payload for reporting a comment.
type ReportCommentRequest struct {
	Reason string `json:"reason"`
}

// CommentResponse is the response for a single comment, optionally with like status.
type CommentResponse struct {
	model.Comment
	IsLiked bool `json:"is_liked"`
}

func (s *CommentService) Create(userID uint64, req CreateCommentRequest) (*model.Comment, error) {
	if req.VideoID == 0 {
		return nil, errors.ErrValidation.WithMessage("video_id is required")
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.ErrValidation.WithMessage("content is required")
	}

	// Check sensitive words
	if err := s.checkSensitiveWords(req.Content); err != nil {
		return nil, err
	}

	// If replying, verify the parent comment exists and is active
	if req.ParentID != nil {
		parent, err := s.commentRepo.FindByID(*req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.ErrNotFound.WithMessage("parent comment not found")
			}
			return nil, errors.ErrInternal
		}
		if parent.Status != model.CommentStatusActive {
			return nil, errors.ErrNotFound.WithMessage("parent comment not found")
		}
		// Replies must be on the same video
		if parent.VideoID != req.VideoID {
			return nil, errors.ErrValidation.WithMessage("parent comment belongs to a different video")
		}
		// Prevent nested replies (only allow one level of nesting)
		if parent.ParentID != nil {
			return nil, errors.ErrValidation.WithMessage("cannot reply to a reply")
		}
	}

	comment := &model.Comment{
		VideoID:  req.VideoID,
		UserID:   userID,
		ParentID: req.ParentID,
		Content:  strings.TrimSpace(req.Content),
		Status:   model.CommentStatusActive,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to create comment")
	}

	// Increment parent reply count if this is a reply
	if req.ParentID != nil {
		_ = s.commentRepo.IncrementCount(*req.ParentID, "reply_count", 1)
	}

	return comment, nil
}

func (s *CommentService) ListByVideoID(videoID uint64, offset, limit int, userID uint64) ([]CommentResponse, int64, error) {
	comments, total, err := s.commentRepo.FindByVideoID(videoID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list comments")
	}

	responses := s.enrichComments(comments, userID)
	return responses, total, nil
}

func (s *CommentService) ListReplies(parentID uint64, offset, limit int, userID uint64) ([]CommentResponse, int64, error) {
	// Verify parent exists
	parent, err := s.commentRepo.FindByID(parentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, errors.ErrNotFound.WithMessage("comment not found")
		}
		return nil, 0, errors.ErrInternal
	}
	if parent.Status != model.CommentStatusActive {
		return nil, 0, errors.ErrNotFound.WithMessage("comment not found")
	}

	comments, total, err := s.commentRepo.FindReplies(parentID, offset, limit)
	if err != nil {
		return nil, 0, errors.ErrInternal.WithMessage("failed to list replies")
	}

	responses := s.enrichComments(comments, userID)
	return responses, total, nil
}

func (s *CommentService) Delete(userID, commentID uint64) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}

	if comment.UserID != userID {
		return errors.ErrForbidden.WithMessage("you can only delete your own comments")
	}

	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	comment.Status = model.CommentStatusDeleted
	if err := s.commentRepo.Update(comment); err != nil {
		return errors.ErrInternal.WithMessage("failed to delete comment")
	}

	// Decrement parent reply count if this is a reply
	if comment.ParentID != nil {
		_ = s.commentRepo.IncrementCount(*comment.ParentID, "reply_count", -1)
	}

	return nil
}

func (s *CommentService) Like(userID, commentID uint64) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}
	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	// Check if already liked
	if _, err := s.likeRepo.FindByCommentAndUser(commentID, userID); err == nil {
		return errors.ErrConflict.WithMessage("already liked this comment")
	}

	like := &model.CommentLike{
		CommentID: commentID,
		UserID:    userID,
	}
	if err := s.likeRepo.Create(like); err != nil {
		return errors.ErrInternal.WithMessage("failed to like comment")
	}

	_ = s.commentRepo.IncrementCount(commentID, "like_count", 1)
	return nil
}

func (s *CommentService) Unlike(userID, commentID uint64) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}
	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	// Check if liked
	if _, err := s.likeRepo.FindByCommentAndUser(commentID, userID); err != nil {
		return errors.ErrNotFound.WithMessage("like not found")
	}

	if err := s.likeRepo.Delete(commentID, userID); err != nil {
		return errors.ErrInternal.WithMessage("failed to unlike comment")
	}

	_ = s.commentRepo.IncrementCount(commentID, "like_count", -1)
	return nil
}

func (s *CommentService) Pin(userID, commentID uint64) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}
	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	// Only top-level comments can be pinned
	if comment.ParentID != nil {
		return errors.ErrValidation.WithMessage("cannot pin a reply")
	}

	comment.IsPinned = !comment.IsPinned
	if err := s.commentRepo.Update(comment); err != nil {
		return errors.ErrInternal.WithMessage("failed to pin comment")
	}
	return nil
}

func (s *CommentService) Heart(userID, commentID uint64) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}
	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	comment.IsHearted = !comment.IsHearted
	if err := s.commentRepo.Update(comment); err != nil {
		return errors.ErrInternal.WithMessage("failed to heart comment")
	}
	return nil
}

func (s *CommentService) Report(userID, commentID uint64, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.ErrValidation.WithMessage("reason is required")
	}

	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrNotFound.WithMessage("comment not found")
		}
		return errors.ErrInternal
	}
	if comment.Status != model.CommentStatusActive {
		return errors.ErrNotFound.WithMessage("comment not found")
	}

	// In a full implementation, we'd store the report in a reports table.
	// For now, we acknowledge the report.
	return nil
}

// enrichComments adds the is_liked field for the current user.
func (s *CommentService) enrichComments(comments []model.Comment, userID uint64) []CommentResponse {
	responses := make([]CommentResponse, len(comments))

	var likedMap map[uint64]bool
	if userID > 0 && len(comments) > 0 {
		commentIDs := make([]uint64, len(comments))
		for i, c := range comments {
			commentIDs[i] = c.ID
		}
		likedMap, _ = s.likeRepo.FindUserLikedCommentIDs(userID, commentIDs)
	}

	for i, c := range comments {
		responses[i] = CommentResponse{
			Comment: c,
			IsLiked: likedMap[c.ID],
		}
	}
	return responses
}

// checkSensitiveWords checks if the content contains any active sensitive words.
func (s *CommentService) checkSensitiveWords(content string) error {
	words, err := s.sensitiveWordRepo.FindActive()
	if err != nil {
		// If we can't fetch words, allow the comment through
		return nil
	}

	lower := strings.ToLower(content)
	for _, w := range words {
		if strings.Contains(lower, strings.ToLower(w.Word)) {
			return errors.ErrValidation.WithMessage("comment contains inappropriate content")
		}
	}
	return nil
}
