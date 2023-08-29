package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b0shka/backend/internal/config"
	handler "github.com/b0shka/backend/internal/handler/http"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/server"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // for connect to postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

//	@title			Service API
//	@version		1.0
//	@description	REST API for Service App

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	UsersAuth
//	@in							header
//	@name						Authorization

func Run(configPath string) { //nolint: funlen
	cfg, err := config.InitConfig(configPath)
	if err != nil {
		logger.Error(err)

		return
	}

	hasher, err := hash.NewSHA256Hasher(cfg.Auth.CodeSalt)
	if err != nil {
		logger.Error(err)

		return
	}

	tokenManager, err := auth.NewPasetoManager(cfg.Auth.SecretKey)
	if err != nil {
		logger.Error(err)

		return
	}

	connPool, err := pgxpool.New(context.Background(), cfg.Postgres.URL)
	if err != nil {
		logger.Errorf("cannot connect to database: %s", err)
	}

	logger.Info("Success connect to database")

	err = runDBMigration(cfg.Postgres.MigrationURL, cfg.Postgres.URL)
	if err != nil {
		logger.Error(err)

		return
	}

	logger.Info("db migrated successfully")

	repos := repository.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.Redis.Address,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runTaskProcessor(redisOpt, repos, hasher, cfg)

	services := service.NewServices(service.Deps{
		Repos:           repos,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AuthConfig:      cfg.Auth,
		TaskDistributor: taskDistributor,
	})

	handlers := handler.NewHandler(services, tokenManager)
	routes := handlers.InitRoutes(cfg)
	srv := server.NewServer(cfg, routes)

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()

	logger.Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)

	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err)
	}

	logger.Info("Server stoped")

	connPool.Close()
	logger.Info("Database disconnected")
}

func runDBMigration(migrationURL, dbSource string) error {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		return err
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func runTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store repository.Store,
	hasher hash.Hasher,
	cfg *config.Config,
) {
	emailService := email.NewEmailService(
		cfg.Email.ServiceName,
		cfg.Email.ServiceAddress,
		cfg.Email.ServicePassword,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
	)

	taskProcessor := worker.NewRedisTaskProcessor(
		redisOpt,
		store,
		hasher,
		emailService,
		cfg.Email,
		cfg.Auth,
	)

	logger.Info("start task processor")

	if err := taskProcessor.Start(); err != nil {
		logger.Error("failed to start task processor")
	}
}
