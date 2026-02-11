package viewmodel

type UserVM struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Handle         string `json:"handle"`
	AvatarURL      string `json:"avatarUrl"`
	Role           string `json:"role"`
	FollowersCount int64  `json:"followersCount"`
}

type VideoVM struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverURL    string   `json:"coverUrl"`
	Duration    int64    `json:"duration"`
	Views       int64    `json:"views"`
	Likes       int64    `json:"likes"`
	CreatedAt   string   `json:"createdAt"`
	Author      UserVM   `json:"author"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category"`
	Visibility  string   `json:"visibility"`
	Status      string   `json:"status"`
	HlsURL      string   `json:"hlsUrl,omitempty"`
}

type CommentVM struct {
	ID         string  `json:"id"`
	EntityType string  `json:"entityType"`
	EntityID   string  `json:"entityId"`
	Author     UserVM  `json:"author"`
	Content    string  `json:"content"`
	CreatedAt  string  `json:"createdAt"`
	LikeCount  int64   `json:"likeCount"`
	ParentID   *string `json:"parentId,omitempty"`
}

type LiveRoomVM struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl"`
	Author    UserVM `json:"author"`
	Viewers   int64  `json:"viewers"`
	Category  string `json:"category"`
	Status    string `json:"status"`
	HlsURL    string `json:"hlsUrl"`
	StartedAt string `json:"startedAt"`
}

type ChapterVM struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Time  int    `json:"time"`
}

type PlaylistVM struct {
	ID    string    `json:"id"`
	Title string    `json:"title"`
	Items []VideoVM `json:"items"`
}
