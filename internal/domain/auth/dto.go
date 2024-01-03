package auth

import "github.com/google/uuid"

type SendCodeEmailInput struct {
	Email string `json:"email"`
}

func NewSendCodeEmailInput(email string) SendCodeEmailInput {
	return SendCodeEmailInput{
		Email: email,
	}
}

type SignInInput struct {
	Email      string `json:"email"`
	SecretCode string `json:"secret_code"`
}

func NewSignInInput(email, secretCode string) SignInInput {
	return SignInInput{
		Email:      email,
		SecretCode: secretCode,
	}
}

type SignInOutput struct {
	SessionID    uuid.UUID `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
}

func NewSignInOutput(sessionID uuid.UUID, refreshToken string, accessToken string) SignInOutput {
	return SignInOutput{
		SessionID:    sessionID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshTokenInput(refreshToken string) RefreshTokenInput {
	return RefreshTokenInput{
		RefreshToken: refreshToken,
	}
}

type RefreshTokenOutput struct {
	AccessToken string `json:"access_token"`
}

func NewRefreshTokenOutput(accessToken string) RefreshTokenOutput {
	return RefreshTokenOutput{
		AccessToken: accessToken,
	}
}
