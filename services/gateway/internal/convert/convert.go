package convert

import (
	"encoding/json"
	"strconv"
	"strings"

	"youtube-code-backend/services/gateway/internal/viewmodel"
)

// Uint64ToStr converts uint64 to string.
func Uint64ToStr(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// StrToUint64 converts string to uint64, returns 0 on error.
func StrToUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// VideoStatus maps DB status to frontend status.
func VideoStatus(dbStatus string) string {
	switch dbStatus {
	case "draft":
		return "draft"
	case "processing":
		return "reviewing"
	case "ready":
		return "published"
	case "rejected":
		return "rejected"
	case "removed":
		return "draft"
	default:
		return dbStatus
	}
}

// VideoStatusToDB maps frontend status to DB status.
func VideoStatusToDB(feStatus string) string {
	switch feStatus {
	case "draft":
		return "draft"
	case "reviewing":
		return "processing"
	case "published":
		return "ready"
	case "rejected":
		return "rejected"
	default:
		return feStatus
	}
}

// VideoType maps DB type to frontend type.
func VideoType(dbType string) string {
	switch dbType {
	case "video":
		return "long"
	case "short":
		return "short"
	default:
		return dbType
	}
}

// VideoTypeToDB maps frontend type to DB type.
func VideoTypeToDB(feType string) string {
	switch feType {
	case "long":
		return "video"
	case "short":
		return "short"
	default:
		return feType
	}
}

// LiveStatus maps DB live_room status to frontend status.
func LiveStatus(dbStatus string) string {
	switch dbStatus {
	case "live":
		return "live"
	default:
		return "offline"
	}
}

// ParseTags parses a tags column that may be JSON array or comma-separated.
func ParseTags(raw string) []string {
	if raw == "" {
		return []string{}
	}
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "[") {
		var tags []string
		if err := json.Unmarshal([]byte(raw), &tags); err == nil {
			return tags
		}
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// UserRow represents the result of a user JOIN query.
type UserRow struct {
	UserID          uint64
	Username        string
	Email           string
	Role            string
	Status          string
	PasswordHash    string
	Nickname        string
	Avatar          string
	ChannelHandle   string
	SubscriberCount int64
}

// UserRowToVM converts a UserRow to a UserVM.
func UserRowToVM(r UserRow) viewmodel.UserVM {
	name := r.Nickname
	if name == "" {
		name = r.Username
	}
	handle := r.ChannelHandle
	if handle == "" {
		handle = "@" + r.Username
	} else if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	role := r.Role
	// DB has "user"|"moderator"|"admin"; frontend expects "user"|"creator"|"admin"
	// Treat "moderator" as "creator" for frontend display
	if role == "moderator" {
		role = "creator"
	}
	return viewmodel.UserVM{
		ID:             Uint64ToStr(r.UserID),
		Name:           name,
		Handle:         handle,
		AvatarURL:      r.Avatar,
		Role:           role,
		FollowersCount: r.SubscriberCount,
	}
}

// VideoRow represents a joined video query result.
type VideoRow struct {
	VideoID      uint64
	ChannelID    uint64
	Type         string
	Title        string
	Description  string
	Status       string
	Visibility   string
	Duration     int64
	ViewCount    int64
	LikeCount    int64
	ThumbnailURL string
	Tags         string
	Category     string
	HlsURL       string
	CreatedAt    string
}

// VideoRowToVM converts a VideoRow + author to VideoVM.
func VideoRowToVM(r VideoRow, author viewmodel.UserVM) viewmodel.VideoVM {
	tags := ParseTags(r.Tags)
	return viewmodel.VideoVM{
		ID:          Uint64ToStr(r.VideoID),
		Type:        VideoType(r.Type),
		Title:       r.Title,
		Description: r.Description,
		CoverURL:    r.ThumbnailURL,
		Duration:    r.Duration,
		Views:       r.ViewCount,
		Likes:       r.LikeCount,
		CreatedAt:   r.CreatedAt,
		Author:      author,
		Tags:        tags,
		Category:    r.Category,
		Visibility:  r.Visibility,
		Status:      VideoStatus(r.Status),
		HlsURL:      r.HlsURL,
	}
}
