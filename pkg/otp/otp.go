package otp

import (
	"github.com/jltorresm/otpgo"
	"github.com/jltorresm/otpgo/config"
)

type Generator interface {
	RandomCode(length int) string
}

type TOTPGenerator struct{}

func NewTOTPGenerator() Generator {
	return &TOTPGenerator{}
}

func (g *TOTPGenerator) RandomCode(length int) string {
	t := otpgo.TOTP{
		Length: config.Length(length),
	}
	token, _ := t.Generate()

	return token
}
