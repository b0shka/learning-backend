package domain

import "github.com/google/uuid"

type (
	SendCodeRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	SignInRequest struct {
		Email      string `json:"email" binding:"required,email"`
		SecretCode int32  `json:"secret_code" binding:"required,min=100000"`
	}

	SignInResponse struct {
		SessionID    uuid.UUID `json:"session_id"`
		RefreshToken string    `json:"refresh_token"`
		AccessToken  string    `json:"access_token"`
	}

	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	RefreshTokenResponse struct {
		AccessToken string `json:"access_token"`
	}
)
