package worker

import (
	"context"

	"github.com/b0shka/backend/internal/config"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/pkg/email"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskSendLoginNotification(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server       *asynq.Server
	store        repository.Store
	emailService *email.EmailService
	emailConfig  config.EmailConfig
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store repository.Store,
	emailService *email.EmailService,
	emailConfig config.EmailConfig,
) TaskProcessor {
	// redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			// ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			// 	log.Error().Err(err).Str("type", task.Type()).
			// 		Bytes("payload", task.Payload()).Msg("process task failed")
			// }),
			// Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server:       server,
		store:        store,
		emailService: emailService,
		emailConfig:  emailConfig,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendLoginNotification, processor.ProcessTaskSendLoginNotification)

	return processor.server.Start(mux)
}
