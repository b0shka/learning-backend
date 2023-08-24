package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/b0shka/backend/pkg/logger"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	// cfg, err := config.InitConfig("../../../../configs")
	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }

	// databseUrl := fmt.Sprintf(
	// 	"postgres://%s:%s@%s:%s/%s?sslmode=disable",
	// 	cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName,
	// )

	conn, err := sql.Open("postgres", "postgresql://root:qwerty@localhost:5432/service?sslmode=disable")
	// conn, err := sql.Open("postgres", databseUrl)
	if err != nil {
		logger.Errorf("cannot to connect to database: %s", err)
	}

	testQueries = New(conn)
	os.Exit(m.Run())
}
