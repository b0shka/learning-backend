package auth

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type VerifyEmail struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	SecretCode string    `json:"secret_code"`
	ExpiresAt  time.Time `json:"expires_at"`
}
