package auth

import (
	"time"

	"github.com/b0shka/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payload struct {
	ID        primitive.ObjectID `json:"id"`
	UserID    primitive.ObjectID `json:"user_id"`
	IssuedAt  int64              `json:"issued_at"`
	ExpiresAt int64              `json:"expires_at"`
}

func NewPayload(userId primitive.ObjectID, duration time.Duration) (*Payload, error) {
	payloadID := primitive.NewObjectID()

	payload := &Payload{
		ID:        payloadID,
		UserID:    userId,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(duration).Unix(),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().Unix() > payload.ExpiresAt {
		return domain.ErrExpiredToken
	}
	return nil
}
