package domain

import (
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" bson:"_id,omitempty"`
	Email     string    `json:"email" bson:"email" binding:"required"`
	Username  string    `json:"username" bson:"username" binding:"required"`
	Photo     string    `json:"photo" bson:"photo"`
	CreatedAt int64     `json:"created_at" bson:"created_at" binding:"required"`
}

type UserUpdate struct {
	Username string `json:"username" bson:"username" binding:"required"`
	Photo    string `json:"photo" bson:"photo"`
}

type Session struct {
	ID           uuid.UUID `json:"id" bson:"_id,omitempty"`
	UserID       uuid.UUID `json:"user_id" bson:"user_id" binding:"required"`
	RefreshToken string    `json:"refresh_token" bson:"refresh_token" binding:"required"`
	UserAgent    string    `json:"user_agent" bson:"user_agent" binding:"required"`
	ClientIP     string    `json:"client_ip" bson:"client_ip" binding:"required"`
	IsBlocked    bool      `json:"is_blocked" bson:"is_blocked" binding:"required"`
	ExpiresAt    int64     `json:"expires_at" bson:"expires_at" binding:"required"`
}
