package repository

import (
	"context"

	"github.com/b0shka/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Users interface {
	AddVerifyEmail(ctx context.Context, verifyEmail domain.VerifyEmail) error
	GetVerifyEmail(ctx context.Context, email, code string) (domain.VerifyEmail, error)
	RemoveVerifyEmail(ctx context.Context, id primitive.ObjectID) error
	CreateSession(ctx context.Context, session domain.Session) error
	GetSession(ctx context.Context, id primitive.ObjectID) (domain.Session, error)
	CreateUser(ctx context.Context, user domain.User) error
	GetUser(ctx context.Context, identifier interface{}) (domain.User, error)
	UpdateUser(ctx context.Context, id primitive.ObjectID, user domain.UserUpdate) error
}

type Repositories struct {
	Users
}

func NewRepositories(db *mongo.Database) *Repositories {
	return &Repositories{
		Users: NewUsersRepo(db),
	}
}
