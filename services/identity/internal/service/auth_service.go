package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/services/identity/internal/model"
	"youtube-code-backend/services/identity/internal/repository"
)

type AuthService struct {
	repo       *repository.UserRepository
	jwtManager *jwt.Manager
}

func NewAuthService(repo *repository.UserRepository, jm *jwt.Manager) *AuthService {
	return &AuthService{repo: repo, jwtManager: jm}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirm struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"new_password"`
}

type VerificationRequest struct {
	Type string `json:"type"` // email or phone
}

type VerificationConfirm struct {
	Code string `json:"code"`
	Type string `json:"type"`
}

func (s *AuthService) Register(req RegisterRequest) (*model.User, *jwt.TokenPair, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, nil, errors.ErrValidation.WithMessage("username, email, and password are required")
	}
	if len(req.Password) < 8 {
		return nil, nil, errors.ErrValidation.WithMessage("password must be at least 8 characters")
	}

	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, nil, errors.ErrConflict.WithMessage("email already registered")
	}
	if _, err := s.repo.FindByUsername(req.Username); err == nil {
		return nil, nil, errors.ErrConflict.WithMessage("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errors.ErrInternal.WithMessage("failed to hash password")
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Role:         model.UserRoleUser,
		Status:       model.UserStatusActive,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, nil, errors.ErrInternal.WithMessage("failed to create user")
	}

	tokens, err := s.generateAndStoreTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(req LoginRequest) (*model.User, *jwt.TokenPair, error) {
	if req.Email == "" || req.Password == "" {
		return nil, nil, errors.ErrValidation.WithMessage("email and password are required")
	}

	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	if user.Status != model.UserStatusActive {
		return nil, nil, errors.ErrForbidden.WithMessage("account is " + string(user.Status))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	tokens, err := s.generateAndStoreTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	hash := hashToken(refreshToken)
	stored, err := s.repo.FindRefreshTokenByHash(hash)
	if err != nil {
		return nil, errors.ErrTokenRevoked
	}

	if stored.Revoked || stored.ExpiresAt < time.Now().Unix() {
		return nil, errors.ErrTokenExpired
	}

	// Revoke old token
	_ = s.repo.RevokeRefreshToken(hash)

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.ErrNotFound.WithMessage("user not found")
	}

	return s.generateAndStoreTokens(user)
}

func (s *AuthService) Logout(refreshToken string) error {
	hash := hashToken(refreshToken)
	return s.repo.RevokeRefreshToken(hash)
}

func (s *AuthService) LogoutAll(userID uint64) error {
	return s.repo.RevokeAllUserTokens(userID)
}

func (s *AuthService) GetMe(userID uint64) (*model.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("user not found")
		}
		return nil, errors.ErrInternal
	}
	return user, nil
}

func (s *AuthService) ChangePassword(userID uint64, req ChangePasswordRequest) error {
	if req.OldPassword == "" || req.NewPassword == "" {
		return errors.ErrValidation.WithMessage("old and new passwords are required")
	}
	if len(req.NewPassword) < 8 {
		return errors.ErrValidation.WithMessage("new password must be at least 8 characters")
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.ErrNotFound.WithMessage("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.ErrInvalidCredentials.WithMessage("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrInternal
	}

	user.PasswordHash = string(hash)
	if err := s.repo.Update(user); err != nil {
		return errors.ErrInternal
	}

	// Revoke all existing tokens
	_ = s.repo.RevokeAllUserTokens(userID)
	return nil
}

func (s *AuthService) RequestPasswordReset(req PasswordResetRequest) error {
	if req.Email == "" {
		return errors.ErrValidation.WithMessage("email is required")
	}

	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		// Don't reveal whether email exists
		return nil
	}

	code := generateCode()
	vc := &model.VerificationCode{
		UserID:    user.ID,
		Code:      code,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}
	_ = s.repo.CreateVerificationCode(vc)
	// In production, send email/SMS here
	return nil
}

func (s *AuthService) ConfirmPasswordReset(req PasswordResetConfirm) error {
	if req.Email == "" || req.Code == "" || req.Password == "" {
		return errors.ErrValidation.WithMessage("email, code, and new password are required")
	}
	if len(req.Password) < 8 {
		return errors.ErrValidation.WithMessage("password must be at least 8 characters")
	}

	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return errors.ErrNotFound.WithMessage("user not found")
	}

	vc, err := s.repo.FindVerificationCode(user.ID, req.Code, "password_reset")
	if err != nil {
		return errors.ErrValidation.WithMessage("invalid or expired code")
	}

	if vc.ExpiresAt < time.Now().Unix() {
		return errors.ErrValidation.WithMessage("code has expired")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrInternal
	}

	user.PasswordHash = string(hash)
	if err := s.repo.Update(user); err != nil {
		return errors.ErrInternal
	}

	_ = s.repo.MarkVerificationCodeUsed(vc.ID)
	_ = s.repo.RevokeAllUserTokens(user.ID)
	return nil
}

func (s *AuthService) SendVerification(userID uint64, req VerificationRequest) error {
	code := generateCode()
	vc := &model.VerificationCode{
		UserID:    userID,
		Code:      code,
		Type:      req.Type,
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
	}
	return s.repo.CreateVerificationCode(vc)
}

func (s *AuthService) ConfirmVerification(userID uint64, req VerificationConfirm) error {
	vc, err := s.repo.FindVerificationCode(userID, req.Code, req.Type)
	if err != nil {
		return errors.ErrValidation.WithMessage("invalid or expired code")
	}
	if vc.ExpiresAt < time.Now().Unix() {
		return errors.ErrValidation.WithMessage("code has expired")
	}
	return s.repo.MarkVerificationCodeUsed(vc.ID)
}

func (s *AuthService) UpdateUserRole(userID uint64, role model.UserRole) error {
	return s.repo.UpdateRole(userID, role)
}

func (s *AuthService) UpdateUserStatus(userID uint64, status model.UserStatus) error {
	return s.repo.UpdateStatus(userID, status)
}

func (s *AuthService) generateAndStoreTokens(user *model.User) (*jwt.TokenPair, error) {
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to generate tokens")
	}

	hash := hashToken(tokens.RefreshToken)
	rt := &model.RefreshToken{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTokenExpiry()).Unix(),
	}
	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to store refresh token")
	}

	return tokens, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
