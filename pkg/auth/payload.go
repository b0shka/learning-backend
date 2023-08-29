package auth

import (
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewPayload(userID uuid.UUID, duration time.Duration) (*Payload, error) {
	payloadID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        payloadID,
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
