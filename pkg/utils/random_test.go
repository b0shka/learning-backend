package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUtils_RandomInt(t *testing.T) {
	min := 1
	max := 100
	random := RandomInt(int32(min), int32(max))

	require.NotEmpty(t, random)
	require.LessOrEqual(t, random, int32(max))
	require.GreaterOrEqual(t, random, int32(min))
}

func TestUtils_RandomString(t *testing.T) {
	length := 20
	random := RandomString(length)

	require.NotEmpty(t, random)
	require.Equal(t, len(random), length)
}
