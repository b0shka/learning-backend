package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandomInt(min, max int32) int32 {
	if min >= max {
		return min
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		panic(err)
	}

	return min + int32(n.Int64())
}

func RandomString(length int) string {
	sb := make([]byte, length)
	k := big.NewInt(int64(len(alphabet)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, k)
		if err != nil {
			panic(err)
		}

		sb[i] = alphabet[int(n.Int64())]
	}

	return string(sb)
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
