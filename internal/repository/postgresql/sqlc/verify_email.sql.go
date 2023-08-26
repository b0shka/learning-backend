// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: verify_email.sql

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createVerifyEmail = `-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
    id,
    email,
    secret_code,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING id, email, secret_code, expires_at
`

type CreateVerifyEmailParams struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	SecretCode string    `json:"secret_code"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (q *Queries) CreateVerifyEmail(ctx context.Context, arg CreateVerifyEmailParams) (VerifyEmail, error) {
	row := q.db.QueryRowContext(ctx, createVerifyEmail,
		arg.ID,
		arg.Email,
		arg.SecretCode,
		arg.ExpiresAt,
	)
	var i VerifyEmail
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.SecretCode,
		&i.ExpiresAt,
	)
	return i, err
}

const deleteVerifyEmailByEmail = `-- name: DeleteVerifyEmailByEmail :exec
DELETE FROM verify_emails
WHERE email = $1
`

func (q *Queries) DeleteVerifyEmailByEmail(ctx context.Context, email string) error {
	_, err := q.db.ExecContext(ctx, deleteVerifyEmailByEmail, email)
	return err
}

const deleteVerifyEmailById = `-- name: DeleteVerifyEmailById :exec
DELETE FROM verify_emails
WHERE id = $1
`

func (q *Queries) DeleteVerifyEmailById(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteVerifyEmailById, id)
	return err
}

const getVerifyEmail = `-- name: GetVerifyEmail :one
SELECT id, email, secret_code, expires_at FROM verify_emails
WHERE email = $1 AND secret_code = $2 LIMIT 1
`

type GetVerifyEmailParams struct {
	Email      string `json:"email"`
	SecretCode string `json:"secret_code"`
}

func (q *Queries) GetVerifyEmail(ctx context.Context, arg GetVerifyEmailParams) (VerifyEmail, error) {
	row := q.db.QueryRowContext(ctx, getVerifyEmail, arg.Email, arg.SecretCode)
	var i VerifyEmail
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.SecretCode,
		&i.ExpiresAt,
	)
	return i, err
}
