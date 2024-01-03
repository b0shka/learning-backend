package repository

import (
	"context"

	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	domain_user "github.com/b0shka/backend/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerifyEmails interface {
	Create(ctx context.Context, arg CreateVerifyEmailParams) (domain_auth.VerifyEmail, error)
	Get(ctx context.Context, arg GetVerifyEmailParams) (domain_auth.VerifyEmail, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
}

type Sessions interface {
	Create(ctx context.Context, arg CreateSessionParams) (domain_auth.Session, error)
	Get(ctx context.Context, id uuid.UUID) (domain_auth.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Users interface {
	Create(ctx context.Context, arg CreateUserParams) (domain_user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain_user.User, error)
	GetByEmail(ctx context.Context, email string) (domain_user.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Repositories struct {
	VerifyEmails VerifyEmails
	Sessions     Sessions
	Users        Users
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		VerifyEmails: NewVerifyEmailsRepo(db),
		Sessions:     NewSessionsRepo(db),
		Users:        NewUsersRepo(db),
	}
}
