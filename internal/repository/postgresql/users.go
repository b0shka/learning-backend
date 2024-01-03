package repository

import (
	"context"
	"errors"
	"fmt"

	domain_user "github.com/b0shka/backend/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	uniqueViolation = "23505"
)

type UsersRepo struct {
	db *pgxpool.Pool
}

func NewUsersRepo(db *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{
		db: db,
	}
}

type CreateUserParams struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (r *UsersRepo) Create(ctx context.Context, arg CreateUserParams) (domain_user.User, error) {
	q := `
		INSERT INTO users 
		    (id, email) 
		VALUES 
			($1, $2)
		RETURNING id, email, created_at
	`

	var user domain_user.User
	if err := r.db.
		QueryRow(ctx, q, arg.ID, arg.Email).
		Scan(
			&user.ID,
			&user.Email,
			&user.CreatedAt,
		); err != nil {
		var pgErr *pgconn.PgError

		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == uniqueViolation {
				return domain_user.User{}, nil
			}

			newErr := fmt.Errorf(
				fmt.Sprintf(
					"SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
					pgErr.Message,
					pgErr.Detail,
					pgErr.Where,
					pgErr.Code,
					pgErr.SQLState(),
				),
			)

			return domain_user.User{}, newErr
		}

		return domain_user.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) GetByID(ctx context.Context, id uuid.UUID) (domain_user.User, error) {
	q := `
		SELECT id, email, created_at FROM users WHERE id = $1
	`

	var user domain_user.User
	if err := r.db.
		QueryRow(ctx, q, id).
		Scan(
			&user.ID,
			&user.Email,
			&user.CreatedAt,
		); err != nil {
		return domain_user.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (domain_user.User, error) {
	q := `
		SELECT id, email, created_at FROM users WHERE email = $1
	`

	var user domain_user.User
	if err := r.db.
		QueryRow(ctx, q, email).
		Scan(
			&user.ID,
			&user.Email,
			&user.CreatedAt,
		); err != nil {
		return domain_user.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `
		DELETE FROM users WHERE id = $1
	`

	_, err := r.db.Exec(ctx, q, id)

	return err
}
