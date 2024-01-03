package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepo struct {
	db *pgxpool.Pool
}

func NewSessionsRepo(db *pgxpool.Pool) *SessionsRepo {
	return &SessionsRepo{
		db: db,
	}
}

type CreateSessionParams struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (r *SessionsRepo) Create(ctx context.Context, arg CreateSessionParams) (domain_auth.Session, error) {
	q := `
		INSERT INTO sessions 
		    (id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at
	`

	var session domain_auth.Session
	if err := r.db.
		QueryRow(
			ctx,
			q,
			arg.ID,
			arg.UserID,
			arg.RefreshToken,
			arg.UserAgent,
			arg.ClientIP,
			arg.IsBlocked,
			arg.ExpiresAt,
		).
		Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.ClientIP,
			&session.IsBlocked,
			&session.ExpiresAt,
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

			return domain_auth.Session{}, newErr
		}

		return domain_auth.Session{}, err
	}

	return session, nil
}

func (r *SessionsRepo) Get(ctx context.Context, id uuid.UUID) (domain_auth.Session, error) {
	q := `
		SELECT 
			id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at 
		FROM sessions 
		WHERE id = $1
	`

	var session domain_auth.Session

	err := r.db.
		QueryRow(ctx, q, id).
		Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.ClientIP,
			&session.IsBlocked,
			&session.ExpiresAt,
		)
	if err != nil {
		return domain_auth.Session{}, err
	}

	return session, nil
}

func (r *SessionsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `
		DELETE FROM sessions WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, q, id)

	return err
}
