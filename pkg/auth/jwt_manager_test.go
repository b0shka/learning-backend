package auth

import (
	"testing"
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthJWT_NewJWTManager(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		shouldErr bool
	}{
		{
			name:      "ok",
			key:       utils.RandomString(32),
			shouldErr: false,
		},
		{
			name:      "invalid key length",
			key:       utils.RandomString(31),
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
			manager, err := NewJWTManager(testCase.key)

			if testCase.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, &JWTManager{}, manager)
			}
		})
	}

}

func TestAuthJWT_CreateTokenAndVerify(t *testing.T) {
	manager, err := NewJWTManager(utils.RandomString(32))
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

	jwtTokenWithNoneSigning := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	tokenWithNoneSigning, err := jwtTokenWithNoneSigning.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

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
			name:          "invalid token",
			payload:       payload,
			token:         tokenWithNoneSigning,
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
