package http

import (
	"github.com/b0shka/backend/internal/domain/auth"
	"github.com/b0shka/backend/internal/domain/user"
)

func NewSendCodeEmailInput(req SendCodeRequest) auth.SendCodeEmailInput {
	return auth.NewSendCodeEmailInput(req.Email)
}

func NewSignInInput(req SignInRequest) auth.SignInInput {
	return auth.NewSignInInput(req.Email, req.SecretCode)
}

func NewSignInResponse(out auth.SignInOutput) SignInResponse {
	return SignInResponse{
		SessionID:    out.SessionID,
		RefreshToken: out.RefreshToken,
		AccessToken:  out.AccessToken,
	}
}

func NewRefreshTokenInput(req RefreshTokenRequest) auth.RefreshTokenInput {
	return auth.NewRefreshTokenInput(req.RefreshToken)
}

func NewRefreshTokenResponse(out auth.RefreshTokenOutput) RefreshTokenResponse {
	return RefreshTokenResponse{
		AccessToken: out.AccessToken,
	}
}

func NewGetUserResponse(out user.User) GetUserResponse {
	return GetUserResponse{
		ID:        out.ID,
		Email:     out.Email,
		CreatedAt: out.CreatedAt,
	}
}
