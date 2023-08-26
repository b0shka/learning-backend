package service

import (
	"context"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserSignInInput struct {
	Email      string
	SecretCode int32
}

type Tokens struct {
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"accesss_token_expires_at"`
}

type RefreshToken struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"accesss_token_expires_at"`
}

type Users interface {
	SendCodeEmail(ctx context.Context, email string) error
	SignIn(ctx *gin.Context, inp UserSignInInput) (Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
	GetById(ctx context.Context, id uuid.UUID) (repository.User, error)
	Update(ctx context.Context, id uuid.UUID, user domain.UserUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Services struct {
	Users
}

type Deps struct {
	Repos        repository.Store
	Hasher       hash.Hasher
	TokenManager auth.Manager
	EmailService email.EmailService
	EmailConfig  config.EmailConfig
	AuthConfig   config.AuthConfig
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users: NewUsersService(
			deps.Repos,
			deps.Hasher,
			deps.TokenManager,
			deps.EmailService,
			deps.EmailConfig,
			deps.AuthConfig,
		),
	}
}
