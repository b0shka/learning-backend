package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Photo     string    `json:"photo"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}

type UserSignIn struct {
	Email      string `json:"email" binding:"required,email"`
	SecretCode int32  `json:"secret_code" binding:"required,min=100000"`
}

type UserUpdate struct {
	Username string `json:"username" binding:"required"`
	Photo    string `json:"photo"`
}
