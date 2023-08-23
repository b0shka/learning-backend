package hash

import (
	"strconv"
	"testing"

	"github.com/b0shka/backend/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestHash_NewSHA256Hasher(t *testing.T) {
	tests := []struct {
		name      string
		salt      string
		shouldErr bool
	}{
		{
			name:      "ok",
			salt:      utils.RandomString(32),
			shouldErr: false,
		},
		{
			name:      "invalid salt length",
			salt:      utils.RandomString(31),
			shouldErr: true,
		},
		{
			name:      "invalid salt length",
			salt:      "",
			shouldErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hasher, err := NewSHA256Hasher(testCase.salt)

			if testCase.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, &SHA256Hasher{}, hasher)
			}
		})
	}
}

func TestHash_HashCode(t *testing.T) {
	code := utils.RandomInt(100000, 999999)

	hasher, err := NewSHA256Hasher(utils.RandomString(32))
	require.NoError(t, err)

	hashCode, err := hasher.HashCode(strconv.Itoa(int(code)))
	require.NoError(t, err)
	require.NotEmpty(t, hashCode)
}
