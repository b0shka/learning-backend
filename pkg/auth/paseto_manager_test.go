package auth

import (
	"testing"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthPaseto_NewPasetoManager(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		shouldErr bool
	}{
		{
			name:      "ok",
			key:       utils.RandomString(chacha20poly1305.KeySize),
			shouldErr: false,
		},
		{
			name:      "invalid key length",
			key:       utils.RandomString(chacha20poly1305.KeySize - 1),
			shouldErr: true,
		},
		{
			name:      "invalid key length",
			key:       "",
			shouldErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			manager, err := NewPasetoManager(testCase.key)

			if testCase.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, &PasetoManager{}, manager)
			}
		})
	}

}

func TestAuthPaseto_CreateTokenAndVerify(t *testing.T) {
	manager, err := NewPasetoManager(utils.RandomString(chacha20poly1305.KeySize))
	require.NoError(t, err)

	userId := primitive.NewObjectID()
	duration := time.Minute
	payload, err := NewPayload(userId, duration)
	require.NoError(t, err)

	token, err := manager.CreateToken(userId, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	tokenExpired, err := manager.CreateToken(userId, -duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	tests := []struct {
		name          string
		payload       *Payload
		token         string
		shouldErr     bool
		expectedError error
	}{
		{
			name:      "ok",
			payload:   payload,
			token:     token,
			shouldErr: false,
		},
		{
			name:          "invalid token",
			payload:       payload,
			token:         "",
			shouldErr:     true,
			expectedError: domain.ErrInvalidToken,
		},
		{
			name:          "invalid token",
			payload:       payload,
			token:         "token",
			shouldErr:     true,
			expectedError: domain.ErrInvalidToken,
		},
		{
			name:          "expired token",
			payload:       payload,
			token:         tokenExpired,
			shouldErr:     true,
			expectedError: domain.ErrExpiredToken,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			payload, err = manager.VerifyToken(testCase.token)

			if testCase.shouldErr {
				require.Error(t, err)
				require.Equal(t, err, testCase.expectedError)
				require.Nil(t, payload)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				require.NotZero(t, payload.ID)
				require.Equal(t, testCase.payload.UserID, payload.UserID)
				require.WithinDuration(t, testCase.payload.IssuedAt, payload.IssuedAt, time.Second)
				require.WithinDuration(t, testCase.payload.ExpiredAt, payload.ExpiredAt, time.Second)
			}
		})
	}
}
