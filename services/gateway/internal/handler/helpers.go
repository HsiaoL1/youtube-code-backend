package handler

import (
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

// enrichVideos converts VideoDBRows to VideoVMs by batch-loading authors.
func enrichVideos(rows []repository.VideoDBRow, userRepo *repository.UserRepo) ([]viewmodel.VideoVM, error) {
	if len(rows) == 0 {
		return []viewmodel.VideoVM{}, nil
	}

	// Collect unique channel IDs
	channelIDs := make([]uint64, 0, len(rows))
	seen := map[uint64]bool{}
	for _, r := range rows {
		if !seen[r.ChannelID] {
			channelIDs = append(channelIDs, r.ChannelID)
			seen[r.ChannelID] = true
		}
	}

	// Load channel_id -> user_id mapping
	channelUserMap, err := userRepo.ChannelUserMapping(channelIDs)
	if err != nil {
		return nil, err
	}

	// Collect unique user IDs
	userIDs := make([]uint64, 0, len(channelUserMap))
	seenUsers := map[uint64]bool{}
	for _, uid := range channelUserMap {
		if !seenUsers[uid] {
			userIDs = append(userIDs, uid)
			seenUsers[uid] = true
		}
	}

	users, err := userRepo.BatchFindUserDetails(userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]viewmodel.VideoVM, 0, len(rows))
	for _, r := range rows {
		userID := channelUserMap[r.ChannelID]
		author := viewmodel.UserVM{}
		if ur, ok := users[userID]; ok {
			author = convert.UserRowToVM(ur)
		}
		vr := convert.VideoRow{
			VideoID:      r.ID,
			ChannelID:    r.ChannelID,
			Type:         r.Type,
			Title:        r.Title,
			Description:  r.Description,
			Status:       r.Status,
			Visibility:   r.Visibility,
			Duration:     r.Duration,
			ViewCount:    r.ViewCount,
			LikeCount:    r.LikeCount,
			ThumbnailURL: r.ThumbnailURL,
			Tags:         r.Tags,
			Category:     r.Category,
			HlsURL:       r.HlsURL,
			CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		}
		result = append(result, convert.VideoRowToVM(vr, author))
	}
	return result, nil
}

// enrichComments converts CommentDBRows to CommentVMs.
func enrichComments(rows []repository.CommentDBRow, userRepo *repository.UserRepo, entityType string) ([]viewmodel.CommentVM, error) {
	if len(rows) == 0 {
		return []viewmodel.CommentVM{}, nil
	}

	userIDs := make([]uint64, 0, len(rows))
	seen := map[uint64]bool{}
	for _, r := range rows {
		if !seen[r.UserID] {
			userIDs = append(userIDs, r.UserID)
			seen[r.UserID] = true
		}
	}

	users, err := userRepo.BatchFindUserDetails(userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]viewmodel.CommentVM, 0, len(rows))
	for _, r := range rows {
		author := viewmodel.UserVM{}
		if ur, ok := users[r.UserID]; ok {
			author = convert.UserRowToVM(ur)
		}
		vm := viewmodel.CommentVM{
			ID:         convert.Uint64ToStr(r.ID),
			EntityType: entityType,
			EntityID:   convert.Uint64ToStr(r.VideoID),
			Author:     author,
			Content:    r.Content,
			CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
			LikeCount:  r.LikeCount,
		}
		if r.ParentID != nil {
			pid := convert.Uint64ToStr(*r.ParentID)
			vm.ParentID = &pid
		}
		result = append(result, vm)
	}
	return result, nil
}

// enrichLiveRooms converts LiveRoomDBRows to LiveRoomVMs.
func enrichLiveRooms(rows []repository.LiveRoomDBRow, userRepo *repository.UserRepo) ([]viewmodel.LiveRoomVM, error) {
	if len(rows) == 0 {
		return []viewmodel.LiveRoomVM{}, nil
	}

	channelIDs := make([]uint64, 0, len(rows))
	seen := map[uint64]bool{}
	for _, r := range rows {
		if !seen[r.ChannelID] {
			channelIDs = append(channelIDs, r.ChannelID)
			seen[r.ChannelID] = true
		}
	}

	channelUserMap, err := userRepo.ChannelUserMapping(channelIDs)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint64, 0, len(channelUserMap))
	seenUsers := map[uint64]bool{}
	for _, uid := range channelUserMap {
		if !seenUsers[uid] {
			userIDs = append(userIDs, uid)
			seenUsers[uid] = true
		}
	}

	users, err := userRepo.BatchFindUserDetails(userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]viewmodel.LiveRoomVM, 0, len(rows))
	for _, r := range rows {
		userID := channelUserMap[r.ChannelID]
		author := viewmodel.UserVM{}
		if ur, ok := users[userID]; ok {
			author = convert.UserRowToVM(ur)
		}
		result = append(result, viewmodel.LiveRoomVM{
			ID:        convert.Uint64ToStr(r.ID),
			Title:     r.Title,
			CoverURL:  r.ThumbnailURL,
			Author:    author,
			Viewers:   r.ViewerCount,
			Category:  r.Category,
			Status:    convert.LiveStatus(r.Status),
			HlsURL:    r.PlaybackURL,
			StartedAt: r.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		})
	}
	return result, nil
}
