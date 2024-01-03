package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for i := 0; i < 1_000; i++ {
		str, err := RandomString(i)
		assert.NoError(t, err)
		assert.Equal(t, i, len(str))
	}
}

func BenchmarkRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := RandomString(100)
		assert.NoError(b, err)
	}
}

func TestRandomInt(t *testing.T) {
	var i int64
	for i = 0; i < 1000; i++ {
		_, err := RandomInt(0, i)
		assert.NoError(t, err)
	}
}

func BenchmarkRandomInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := RandomInt(0, 100)
		assert.NoError(b, err)
	}
}
