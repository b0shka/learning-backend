package hash

import (
	"crypto/sha256"
	"fmt"
)

var minSaltLength = 32

type Hasher interface {
	HashCode(code string) (string, error)
}

type SHA256Hasher struct {
	salt string
}

func NewSHA256Hasher(salt string) (*SHA256Hasher, error) {
	if len(salt) < minSaltLength {
		return nil, fmt.Errorf("invalid salt length: must be at least %d characters", minSaltLength)
	}
	return &SHA256Hasher{salt: salt}, nil
}

func (h *SHA256Hasher) HashCode(code string) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(code)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum([]byte(h.salt))), nil
}
