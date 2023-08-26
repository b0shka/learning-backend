package auth

import (
	"time"

	"github.com/google/uuid"
)

type Manager interface {
	CreateToken(userId uuid.UUID, tokenTTL time.Duration) (string, *Payload, error)
	VerifyToken(accessToken string) (*Payload, error)
}
