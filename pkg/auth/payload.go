package auth

import (
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/pkg/identity"
	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewPayload(userID uuid.UUID, duration time.Duration) (*Payload, error) {
	idGenerator := identity.NewIDGenerator()

	payload := &Payload{
		ID:        idGenerator.GenerateUUID(),
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiresAt) {
		return domain.ErrExpiredToken
	}

	return nil
}
