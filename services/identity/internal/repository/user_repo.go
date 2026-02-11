package repository

import (
	"youtube-code-backend/services/identity/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateRole(id uint64, role model.UserRole) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *UserRepository) UpdateStatus(id uint64, status model.UserStatus) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

// Refresh token operations

func (r *UserRepository) CreateRefreshToken(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *UserRepository) FindRefreshTokenByHash(hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := r.db.Where("token_hash = ? AND revoked = false", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UserRepository) RevokeRefreshToken(hash string) error {
	return r.db.Model(&model.RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *UserRepository) RevokeAllUserTokens(userID uint64) error {
	return r.db.Model(&model.RefreshToken{}).Where("user_id = ? AND revoked = false", userID).Update("revoked", true).Error
}

// Verification code operations

func (r *UserRepository) CreateVerificationCode(code *model.VerificationCode) error {
	return r.db.Create(code).Error
}

func (r *UserRepository) FindVerificationCode(userID uint64, code, codeType string) (*model.VerificationCode, error) {
	var vc model.VerificationCode
	if err := r.db.Where("user_id = ? AND code = ? AND type = ? AND used = false", userID, code, codeType).
		First(&vc).Error; err != nil {
		return nil, err
	}
	return &vc, nil
}

func (r *UserRepository) MarkVerificationCodeUsed(id uint64) error {
	return r.db.Model(&model.VerificationCode{}).Where("id = ?", id).Update("used", true).Error
}
