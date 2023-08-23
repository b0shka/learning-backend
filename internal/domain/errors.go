package domain

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input body")
	ErrInvalidEmail = errors.New("invalid email")
	ErrIdentifier   = errors.New("invalid identifier type")

	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")

	ErrUserNotFound      = errors.New("user not found")
	ErrSecretCodeInvalid = errors.New("code is incorrect")
	ErrSecretCodeExpired = errors.New("code is expired")
)
