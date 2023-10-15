package domain

import (
	"time"

	"github.com/google/uuid"
)

type (
	GetUserResponse struct {
		ID        uuid.UUID `json:"id" binding:"required"`
		Email     string    `json:"email" binding:"required"`
		Username  string    `json:"username" binding:"required"`
		Photo     string    `json:"photo"`
		CreatedAt time.Time `json:"created_at" binding:"required"`
	}

	UpdateUserRequest struct {
		Username string `json:"username" binding:"required"`
		Photo    string `json:"photo"`
	}
)
