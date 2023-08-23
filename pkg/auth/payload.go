package auth

import (
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payload struct {
	ID        uuid.UUID          `json:"id"`
	UserID    primitive.ObjectID `json:"user_id"`
	IssuedAt  time.Time          `json:"issued_at"`
	ExpiredAt time.Time          `json:"expired_at"`
}

func NewPayload(userId primitive.ObjectID, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		UserID:    userId,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return domain.ErrExpiredToken
	}
	return nil
}
