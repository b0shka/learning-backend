package auth

import "time"

type Manager interface {
	CreateToken(userId string, tokenTTL time.Duration) (string, error)
	VerifyToken(accessToken string) (*Payload, error)
}
