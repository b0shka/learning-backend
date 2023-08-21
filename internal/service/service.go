package service

import (
	"context"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/repository"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserSignInInput struct {
	Email      string
	SecretCode int32
}

type Tokens struct {
	AccessToken string
}

type Users interface {
	SendCodeEmail(ctx context.Context, email string) error
	SignIn(ctx context.Context, inp UserSignInInput) (Tokens, error)
	Get(ctx context.Context, identifier interface{}) (domain.User, error)
	Update(ctx context.Context, id primitive.ObjectID, user domain.UserUpdate) error
}

type Services struct {
	Users
}

type Deps struct {
	Repos        *repository.Repositories
	Hasher       hash.PasswordHasher
	TokenManager auth.TokenManager
	EmailService email.EmailService
	EmailConfig  config.EmailConfig
	AuthConfig   config.AuthConfig
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users: NewUsersService(
			deps.Repos.Users,
			deps.Hasher,
			deps.TokenManager,
			deps.EmailService,
			deps.EmailConfig,
			deps.AuthConfig,
		),
	}
}
