package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Manager interface {
	CreateToken(userId primitive.ObjectID, tokenTTL time.Duration) (string, error)
	VerifyToken(accessToken string) (*Payload, error)
}
