package auth

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/b0shka/backend/internal/domain"
	"github.com/google/uuid"
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

func (m *PasetoManager) CreateToken(userID uuid.UUID, ducation time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userID, ducation)
	if err != nil {
		return "", nil, err
	}

	token, err := m.paseto.Encrypt(m.symmetricKey, payload, nil)

	return token, payload, err
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
