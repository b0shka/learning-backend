package worker

import (
	"context"
	"encoding/json"

	"github.com/b0shka/backend/internal/config"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/logger"
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
	hasher       hash.Hasher
	emailService *email.EmailService
	emailConfig  config.EmailConfig
	authConfig   config.AuthConfig
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store repository.Store,
	hasher hash.Hasher,
	emailService *email.EmailService,
	emailConfig config.EmailConfig,
	authConfig config.AuthConfig,
) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				var data map[string]interface{}
				verr := json.Unmarshal(task.Payload(), &data)
				if verr != nil {
					logger.Errorf("Error decode payload: %s", verr.Error())

					return
				}

				logger.Errorf("process task failed: type - %s, payload - %v, err - %s", task.Type(), data, err.Error())
			}),
			// Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server:       server,
		store:        store,
		hasher:       hasher,
		emailService: emailService,
		emailConfig:  emailConfig,
		authConfig:   authConfig,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendLoginNotification, processor.ProcessTaskSendLoginNotification)

	return processor.server.Start(mux)
}
