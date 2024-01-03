package postgresql

import (
	"context"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/pkg/logger"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func NewClient(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool

	pgxCfg, parseConfigErr := pgxpool.ParseConfig(cfg.URL)
	if parseConfigErr != nil {
		logger.Errorf("Unable to parse config: %v\n", parseConfigErr)

		return nil, parseConfigErr
	}

	pool, parseConfigErr = pgxpool.NewWithConfig(ctx, pgxCfg)
	if parseConfigErr != nil {
		logger.Errorf("Failed to parse PostgreSQL configuration due to error: %v\n", parseConfigErr)

		return nil, parseConfigErr
	}

	err := DoWithAttempts(func() error {
		pingErr := pool.Ping(ctx)
		if pingErr != nil {
			logger.Errorf("Failed to connect to postgres due to error %v... Going to do the next attempt\n", pingErr)

			return pingErr
		}

		return nil
	}, cfg.MaxAttempts, cfg.MaxDelay)
	if err != nil {
		return nil, domain.ErrConnectPostgreSQL
	}

	return pool, nil
}

func DoWithAttempts(fn func() error, maxAttempts int, delay time.Duration) error {
	var err error

	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			maxAttempts--

			continue
		}

		return nil
	}

	return err
}
