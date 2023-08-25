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
	// os.Setenv("APP_ENV", "local")
	// cfg, err := config.InitConfig("../../../../configs")
	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }

	// postgresSource := fmt.Sprintf(
	// 	"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
	// 	cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName,
	// )

	postgresSource := "host=localhost port=5432 user=root dbname=service password=qwerty sslmode=disable"
	conn, err := sql.Open("postgres", postgresSource)
	if err != nil {
		logger.Errorf("cannot to connect to database: %s", err)
	}

	testQueries = New(conn)
	os.Exit(m.Run())
}
