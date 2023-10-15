package service

import (
	"context"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Tokens struct {
	SessionID    uuid.UUID `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
}

type RefreshToken struct {
	AccessToken string `json:"access_token"`
}

type Auth interface {
	SendCodeEmail(ctx context.Context, email string) error
	SignIn(ctx *gin.Context, inp domain.SignInRequest) (Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
}

type Users interface {
	GetByID(ctx context.Context, id uuid.UUID) (repository.User, error)
	Update(ctx context.Context, id uuid.UUID, user domain.UpdateUserRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Services struct {
	Auth
	Users
}

type Deps struct {
	Repos           repository.Store
	Hasher          hash.Hasher
	TokenManager    auth.Manager
	OTPGenerator    otp.Generator
	AuthConfig      config.AuthConfig
	TaskDistributor worker.TaskDistributor
}

func NewServices(deps Deps) *Services {
	return &Services{
		Auth: NewAuthService(
			deps.Repos,
			deps.Hasher,
			deps.TokenManager,
			deps.OTPGenerator,
			deps.AuthConfig,
			deps.TaskDistributor,
		),
		Users: NewUsersService(
			deps.Repos,
		),
	}
}
