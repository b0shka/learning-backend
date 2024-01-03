package repository

import (
	"context"
	"os"
	"testing"

	"github.com/b0shka/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testRepos *Repositories

func TestMain(m *testing.M) {
	postgresSource := "postgresql://root:qwerty@localhost:5432/service?sslmode=disable"

	postgreSQLClient, err := pgxpool.New(context.Background(), postgresSource)
	if err != nil {
		logger.Errorf("cannot to connect to database: %s", err)
	}

	// os.Setenv("APP_ENV", "local")
	// cfg, err := config.InitConfig("../../../../configs")
	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }

	testRepos = NewRepositories(postgreSQLClient)

	os.Exit(m.Run())
}
