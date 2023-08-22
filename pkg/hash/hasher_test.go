package hash

import (
	"strconv"
	"testing"

	"github.com/b0shka/backend/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestHash_HashCode(t *testing.T) {
	code := utils.RandomInt(100000, 999999)
	hasher := NewSHA256Hasher("salt")

	hashCode, err := hasher.HashCode(strconv.Itoa(int(code)))
	require.NoError(t, err)
	require.NotEmpty(t, hashCode)
}
