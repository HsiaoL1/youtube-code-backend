package repository

import (
	"youtube-code-backend/services/gateway/internal/convert"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByID returns a single user's full details.
func (r *UserRepo) FindByID(id uint64) (*convert.UserRow, error) {
	var row convert.UserRow
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.username, u.email, u.role, u.status, u.password_hash,
		       COALESCE(p.nickname, '') AS nickname,
		       COALESCE(p.avatar, '') AS avatar,
		       COALESCE(c.handle, '') AS channel_handle,
		       COALESCE(c.subscriber_count, 0) AS subscriber_count
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN channels c ON c.user_id = u.id
		WHERE u.id = ? AND u.deleted_at IS NULL
	`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.UserID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// FindByUsername finds a user by username.
func (r *UserRepo) FindByUsername(username string) (*convert.UserRow, error) {
	var row convert.UserRow
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.username, u.email, u.role, u.status, u.password_hash,
		       COALESCE(p.nickname, '') AS nickname,
		       COALESCE(p.avatar, '') AS avatar,
		       COALESCE(c.handle, '') AS channel_handle,
		       COALESCE(c.subscriber_count, 0) AS subscriber_count
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN channels c ON c.user_id = u.id
		WHERE u.username = ? AND u.deleted_at IS NULL
	`, username).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.UserID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// FindByEmail finds a user by email.
func (r *UserRepo) FindByEmail(email string) (*convert.UserRow, error) {
	var row convert.UserRow
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.username, u.email, u.role, u.status, u.password_hash,
		       COALESCE(p.nickname, '') AS nickname,
		       COALESCE(p.avatar, '') AS avatar,
		       COALESCE(c.handle, '') AS channel_handle,
		       COALESCE(c.subscriber_count, 0) AS subscriber_count
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN channels c ON c.user_id = u.id
		WHERE u.email = ? AND u.deleted_at IS NULL
	`, email).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.UserID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// BatchFindUserDetails loads multiple users by IDs.
func (r *UserRepo) BatchFindUserDetails(ids []uint64) (map[uint64]convert.UserRow, error) {
	if len(ids) == 0 {
		return map[uint64]convert.UserRow{}, nil
	}
	var rows []convert.UserRow
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.username, u.email, u.role, u.status, u.password_hash,
		       COALESCE(p.nickname, '') AS nickname,
		       COALESCE(p.avatar, '') AS avatar,
		       COALESCE(c.handle, '') AS channel_handle,
		       COALESCE(c.subscriber_count, 0) AS subscriber_count
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN channels c ON c.user_id = u.id
		WHERE u.id IN ? AND u.deleted_at IS NULL
	`, ids).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint64]convert.UserRow, len(rows))
	for _, row := range rows {
		m[row.UserID] = row
	}
	return m, nil
}

// ListAll returns all users.
func (r *UserRepo) ListAll() ([]convert.UserRow, error) {
	var rows []convert.UserRow
	err := r.db.Raw(`
		SELECT u.id AS user_id, u.username, u.email, u.role, u.status, u.password_hash,
		       COALESCE(p.nickname, '') AS nickname,
		       COALESCE(p.avatar, '') AS avatar,
		       COALESCE(c.handle, '') AS channel_handle,
		       COALESCE(c.subscriber_count, 0) AS subscriber_count
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN channels c ON c.user_id = u.id
		WHERE u.deleted_at IS NULL
		ORDER BY u.id
	`).Scan(&rows).Error
	return rows, err
}

// CreateUser creates a user, profile, and channel in a transaction.
func (r *UserRepo) CreateUser(username, email, passwordHash, nickname, handle, avatar string) (uint64, error) {
	var userID uint64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`INSERT INTO users (username, email, password_hash, role, status, created_at, updated_at)
			VALUES (?, ?, ?, 'user', 'active', NOW(), NOW())`, username, email, passwordHash).Error; err != nil {
			return err
		}
		if err := tx.Raw(`SELECT id FROM users WHERE username = ? AND deleted_at IS NULL`, username).Scan(&userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO user_profiles (user_id, nickname, avatar, account_status, created_at, updated_at)
			VALUES (?, ?, ?, 'active', NOW(), NOW())`, userID, nickname, avatar).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO channels (user_id, handle, name, subscriber_count, video_count, created_at, updated_at)
			VALUES (?, ?, ?, 0, 0, NOW(), NOW())`, userID, handle, nickname).Error; err != nil {
			return err
		}
		return nil
	})
	return userID, err
}

// UpdateUserStatus updates a user's status.
func (r *UserRepo) UpdateUserStatus(userID uint64, status string) error {
	return r.db.Exec(`UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?`, status, userID).Error
}

// FindChannelByUserID returns the channel ID for a user.
func (r *UserRepo) FindChannelByUserID(userID uint64) (uint64, error) {
	var channelID uint64
	err := r.db.Raw(`SELECT id FROM channels WHERE user_id = ? AND deleted_at IS NULL LIMIT 1`, userID).Scan(&channelID).Error
	return channelID, err
}

// ChannelUserMapping returns a map of channel_id -> user_id for the given channel IDs.
func (r *UserRepo) ChannelUserMapping(channelIDs []uint64) (map[uint64]uint64, error) {
	if len(channelIDs) == 0 {
		return map[uint64]uint64{}, nil
	}
	type pair struct {
		ID     uint64 `gorm:"column:id"`
		UserID uint64 `gorm:"column:user_id"`
	}
	var pairs []pair
	err := r.db.Raw(`SELECT id, user_id FROM channels WHERE id IN ? AND deleted_at IS NULL`, channelIDs).Scan(&pairs).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint64]uint64, len(pairs))
	for _, p := range pairs {
		m[p.ID] = p.UserID
	}
	return m, nil
}
