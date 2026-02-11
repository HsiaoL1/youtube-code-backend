package errors

import "fmt"

// AppError represents a structured application error.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new AppError.
func New(code, message string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode}
}

// Standard auth errors
var (
	ErrInvalidCredentials = New("AUTH_001", "invalid credentials", 401)
	ErrTokenExpired       = New("AUTH_002", "token has expired", 401)
	ErrTokenInvalid       = New("AUTH_003", "invalid token", 401)
	ErrUnauthorized       = New("AUTH_004", "unauthorized", 401)
	ErrForbidden          = New("AUTH_005", "forbidden", 403)
	ErrTokenRevoked       = New("AUTH_006", "token has been revoked", 401)
)

// Standard validation errors
var (
	ErrValidation     = New("VAL_001", "validation failed", 400)
	ErrBadRequest     = New("VAL_002", "bad request", 400)
	ErrInvalidPayload = New("VAL_003", "invalid request payload", 400)
)

// Standard resource errors
var (
	ErrNotFound  = New("RES_001", "resource not found", 404)
	ErrConflict  = New("RES_002", "resource already exists", 409)
	ErrGone      = New("RES_003", "resource is no longer available", 410)
)

// Standard server errors
var (
	ErrInternal = New("SRV_001", "internal server error", 500)
	ErrTimeout  = New("SRV_002", "request timeout", 504)
)

// Rate limiting
var (
	ErrRateLimited = New("RATE_001", "too many requests", 429)
)

// WithMessage creates a copy of the error with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg, StatusCode: e.StatusCode}
}
