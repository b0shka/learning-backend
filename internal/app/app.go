package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

//	@title			Service API
//	@version		1.0
//	@description	REST API for Service App

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.apikey	UsersAuth
//	@in							header
//	@name						Authorization

func Run(configPath string) {
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

	// mongoClient, err := mongodb.NewClient(cfg.Mongo.URI)
	// if err != nil {
	// 	logger.Errorf("cannot connect to database: %s", err)
	// 	return
	// }
	// db := mongoClient.Database(cfg.Mongo.DBName)
	// repos := repository.NewRepositories(db)

	db, err := sql.Open("postgres", cfg.Postgres.URL)
	if err != nil {
		logger.Errorf("cannot connect to database: %s", err)
		return
	}
	logger.Info("Success connect to database")

	err = runDBMigration(cfg.Postgres.MigrationURL, cfg.Postgres.URL)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("db migrated successfully")

	repos := repository.NewStore(db)

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.Redis.Address,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runTaskProcessor(redisOpt, repos, cfg)

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

	// if err := mongoClient.Disconnect(context.Background()); err != nil {
	// 	logger.Error(err.Error())
	// }

	if err := db.Close(); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Database disconnected")
}

func runDBMigration(migrationURL, dbSource string) error {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		return fmt.Errorf("connot create new migarte instance: %s", err.Error())
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %s", err.Error())
	}

	return nil
}

func runTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store repository.Store,
	cfg *config.Config,
) {
	emailService := email.NewEmailService(
		cfg.Email.ServiceName,
		cfg.Email.ServiceAddress,
		cfg.Email.ServicePassword,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
	)

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, emailService, cfg.Email)
	logger.Info("start task processor")

	err := taskProcessor.Start()
	if err != nil {
		logger.Error("failed to start task processor")
	}
}
