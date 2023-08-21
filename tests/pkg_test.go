package tests

import (
	"strconv"
	"testing"

	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuth_NewManager(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		shouldErr bool
	}{
		{
			name:      "ok",
			key:       "key",
			shouldErr: false,
		},
		{
			name:      "empty key",
			key:       "",
			shouldErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			manager, err := auth.NewManager(testCase.key)

			if testCase.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, &auth.Manager{}, manager)
			}
		})
	}

}

func TestAuth_NewJWTAndParse(t *testing.T) {
	userId := primitive.NewObjectID().Hex()
	manager, err := auth.NewManager("key")
	require.NoError(t, err)

	token, err := manager.NewJWT(userId, 10)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	tests := []struct {
		name      string
		userId    string
		token     string
		shouldErr bool
	}{
		{
			name:      "ok",
			userId:    userId,
			token:     token,
			shouldErr: false,
		},
		{
			name:      "empty token",
			userId:    userId,
			token:     "",
			shouldErr: true,
		},
		{
			name:      "invalid token",
			userId:    userId,
			token:     "token",
			shouldErr: true,
		},
		{
			name:      "unexpected signing method",
			userId:    userId,
			token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			shouldErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			res, err := manager.Parse(testCase.token)

			if testCase.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.Equal(t, testCase.userId, res)
			}
		})
	}
}

func TestHash_HashCode(t *testing.T) {
	code := utils.RandomInt(100000, 999999)
	hasher := hash.NewSHA256Hasher("salt")

	hashCode, err := hasher.HashCode(strconv.Itoa(int(code)))
	require.NoError(t, err)
	require.NotEmpty(t, hashCode)
}

func TestUtils_RandomInt(t *testing.T) {
	min := 1
	max := 100
	random := utils.RandomInt(int32(min), int32(max))

	require.NotEmpty(t, random)
	require.LessOrEqual(t, random, int32(max))
	require.GreaterOrEqual(t, random, int32(min))
}
