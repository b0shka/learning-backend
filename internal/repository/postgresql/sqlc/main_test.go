package repository

import (
	"context"
	"os"
	"testing"

	"github.com/b0shka/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore Store

func TestMain(m *testing.M) {
	// os.Setenv("APP_ENV", "local")
	// cfg, err := config.InitConfig("../../../../configs")
	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }

	postgresSource := "postgresql://root:qwerty@localhost:5432/service?sslmode=disable"
	connPool, err := pgxpool.New(context.Background(), postgresSource)
	if err != nil {
		logger.Errorf("cannot to connect to database: %s", err)
	}

	testStore = NewStore(connPool)
	os.Exit(m.Run())
}
