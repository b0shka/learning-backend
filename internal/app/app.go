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
	repository "github.com/b0shka/backend/internal/repository/postgresql"
	"github.com/b0shka/backend/internal/server"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/database/postgresql"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/identity"
	"github.com/b0shka/backend/pkg/logger"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // for connect to postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/streadway/amqp"
)

//	@title			Service API
//	@version		1.0
//	@description	REST API for Service App

//	@host		localhost:8080
//	@BasePath	/api/v1/

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

	otpGenerator := otp.NewTOTPGenerator()
	idGenerator := identity.NewIDGenerator()

	rabbitMQClient, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Errorf("Cannot connect to RabbitMQ: %s", err)

		return
	}

	logger.Info("Success connect to RabbitMQ")

	postgreSQLClient, err := postgresql.NewClient(context.Background(), cfg.Postgres)
	if err != nil {
		logger.Errorf("Cannot connect to database: %s", err)

		return
	}

	logger.Info("Success connect to database")

	err = runDBMigration(cfg.Postgres.MigrationURL, cfg.Postgres.URL)
	if err != nil {
		logger.Error(err)

		return
	}

	logger.Info("DB migrated successfully")

	repos := repository.NewRepositories(postgreSQLClient)

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.Redis.Address,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	go runTaskProcessor(redisOpt, repos, hasher, idGenerator, cfg)

	services := service.NewServices(service.Deps{
		Repos:           repos,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		OTPGenerator:    otpGenerator,
		IDGenerator:     idGenerator,
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
	gracefulShutdown(srv, postgreSQLClient, rabbitMQClient)
}

func gracefulShutdown(srv *server.Server, postgreSQLClient *pgxpool.Pool, rabbitMQClient *amqp.Connection) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)

	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("Failed to stop server: %v", err)
	}

	logger.Info("Server stopped")

	postgreSQLClient.Close()
	logger.Info("Database disconnected")

	rabbitMQClient.Close()
	logger.Info("RabbitMQ disconnected")
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
	repos *repository.Repositories,
	hasher hash.Hasher,
	idGenerator identity.Generator,
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
		repos,
		hasher,
		idGenerator,
		emailService,
		cfg.Email,
		cfg.Auth,
	)

	logger.Info("Start task processor")

	if err := taskProcessor.Start(); err != nil {
		logger.Error("Failed to start task processor")
	}
}
