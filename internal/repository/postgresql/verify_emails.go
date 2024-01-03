package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/b0shka/backend/internal/domain"
	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerifyEmailsRepo struct {
	db *pgxpool.Pool
}

func NewVerifyEmailsRepo(db *pgxpool.Pool) *VerifyEmailsRepo {
	return &VerifyEmailsRepo{
		db: db,
	}
}

type CreateVerifyEmailParams struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	SecretCode string    `json:"secret_code"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (r *VerifyEmailsRepo) Create(ctx context.Context, arg CreateVerifyEmailParams) (domain_auth.VerifyEmail, error) {
	q := `
		INSERT INTO verify_emails 
		    (id, email, secret_code, expires_at)
		VALUES 
			($1, $2, $3, $4)
		RETURNING id, email, secret_code, expires_at
	`

	var verifyEmail domain_auth.VerifyEmail
	if err := r.db.
		QueryRow(
			ctx,
			q,
			arg.ID,
			arg.Email,
			arg.SecretCode,
			arg.ExpiresAt,
		).
		Scan(
			&verifyEmail.ID,
			&verifyEmail.Email,
			&verifyEmail.SecretCode,
			&verifyEmail.ExpiresAt,
		); err != nil {
		var pgErr *pgconn.PgError

		if ok := errors.As(err, &pgErr); ok {
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

			return domain_auth.VerifyEmail{}, newErr
		}

		return domain_auth.VerifyEmail{}, err
	}

	return verifyEmail, nil
}

type GetVerifyEmailParams struct {
	Email      string `json:"email"`
	SecretCode string `json:"secret_code"`
}

func (r *VerifyEmailsRepo) Get(ctx context.Context, arg GetVerifyEmailParams) (domain_auth.VerifyEmail, error) {
	q := `
		SELECT id, email, secret_code, expires_at FROM verify_emails WHERE email = $1 AND secret_code = $2
	`

	var verifyEmail domain_auth.VerifyEmail
	if err := r.db.
		QueryRow(ctx, q, arg.Email, arg.SecretCode).
		Scan(
			&verifyEmail.ID,
			&verifyEmail.Email,
			&verifyEmail.SecretCode,
			&verifyEmail.ExpiresAt,
		); err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return domain_auth.VerifyEmail{}, domain.ErrSecretCodeInvalid
		}

		return domain_auth.VerifyEmail{}, err
	}

	return verifyEmail, nil
}

func (r *VerifyEmailsRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	q := `
		DELETE FROM verify_emails WHERE id = $1
	`

	_, err := r.db.Exec(ctx, q, id)

	return err
}

func (r *VerifyEmailsRepo) DeleteByEmail(ctx context.Context, email string) error {
	q := `
		DELETE FROM verify_emails WHERE email = $1
	`

	_, err := r.db.Exec(ctx, q, email)

	return err
}
