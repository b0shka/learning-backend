package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandomInt(min, max int64) (int64, error) {
	if min >= max {
		return min, nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, err
	}

	return min + n.Int64(), nil
}

func RandomString(length int) (string, error) {
	sb := make([]byte, length)
	k := big.NewInt(int64(len(alphabet)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, k)
		if err != nil {
			return "", err
		}

		sb[i] = alphabet[int(n.Int64())]
	}

	return string(sb), nil
}

func RandomEmail() (string, error) {
	randStr, err := RandomString(6)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@email.com", randStr), nil
}
