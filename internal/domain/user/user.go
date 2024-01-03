package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}
