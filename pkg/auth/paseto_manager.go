package auth

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/b0shka/backend/internal/domain"
	"github.com/o1egl/paseto"
)

type PasetoManager struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoManager(symmetricKey string) (Manager, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key length: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	return &PasetoManager{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}

func (m *PasetoManager) CreateToken(userId string, ducation time.Duration) (string, error) {
	payload, err := NewPayload(userId, ducation)
	if err != nil {
		return "", err
	}

	return m.paseto.Encrypt(m.symmetricKey, payload, nil)
}

func (m *PasetoManager) VerifyToken(accessToken string) (*Payload, error) {
	payload := &Payload{}

	err := m.paseto.Decrypt(accessToken, m.symmetricKey, payload, nil)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
