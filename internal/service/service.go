package service

import (
	"context"

	"github.com/b0shka/backend/internal/config"
	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	domain_user "github.com/b0shka/backend/internal/domain/user"
	repository "github.com/b0shka/backend/internal/repository/postgresql"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/identity"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Auth interface {
	SendCodeEmail(ctx context.Context, inp domain_auth.SendCodeEmailInput) error
	SignIn(ctx *gin.Context, inp domain_auth.SignInInput) (domain_auth.SignInOutput, error)
	RefreshToken(ctx context.Context, inp domain_auth.RefreshTokenInput) (domain_auth.RefreshTokenOutput, error)
}

type Users interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain_user.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Services struct {
	Auth
	Users
}

type Deps struct {
	Repos           *repository.Repositories
	Hasher          hash.Hasher
	TokenManager    auth.Manager
	OTPGenerator    otp.Generator
	IDGenerator     identity.Generator
	AuthConfig      config.AuthConfig
	TaskDistributor worker.TaskDistributor
}

func NewServices(deps Deps) *Services {
	return &Services{
		Auth: NewAuthService(
			deps.Repos.Users,
			deps.Repos.Sessions,
			deps.Repos.VerifyEmails,
			deps.Hasher,
			deps.TokenManager,
			deps.OTPGenerator,
			deps.IDGenerator,
			deps.AuthConfig,
			deps.TaskDistributor,
		),
		Users: NewUsersService(
			deps.Repos.Users,
			deps.Repos.Sessions,
			deps.Repos.VerifyEmails,
		),
	}
}
