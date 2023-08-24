package service

import (
	"context"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/repository"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserSignInInput struct {
	Email      string
	SecretCode int32
}

type Tokens struct {
	RefreshToken          string
	RefreshTokenExpiresAt int64
	AccessToken           string
	AccessTokenExpiresAt  int64
}

type RefreshToken struct {
	AccessToken          string
	AccessTokenExpiresAt int64
}

type Users interface {
	SendCodeEmail(ctx context.Context, email string) error
	SignIn(ctx *gin.Context, inp UserSignInInput) (Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
	Get(ctx context.Context, identifier interface{}) (domain.User, error)
	Update(ctx context.Context, id primitive.ObjectID, user domain.UserUpdate) error
}

type Services struct {
	Users
}

type Deps struct {
	Repos        *repository.Repositories
	Hasher       hash.Hasher
	TokenManager auth.Manager
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
