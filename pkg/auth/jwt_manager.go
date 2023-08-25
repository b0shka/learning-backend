package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var minSecretKeyLength = 32

type JWTManager struct {
	secretKey string
}

func NewJWTManager(secretKey string) (Manager, error) {
	if len(secretKey) < minSecretKeyLength {
		return nil, fmt.Errorf("invalid key length: must be at least %d characters", minSecretKeyLength)
	}

	return &JWTManager{secretKey: secretKey}, nil
}

func (m *JWTManager) CreateToken(userId uuid.UUID, ducation time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userId, ducation)
	if err != nil {
		return "", nil, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(m.secretKey))
	return token, payload, err
}

func (m *JWTManager) VerifyToken(accessToken string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, domain.ErrInvalidToken
		}

		return []byte(m.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(accessToken, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, domain.ErrExpiredToken) {
			return nil, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return payload, nil
}
