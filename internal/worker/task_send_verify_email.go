package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/b0shka/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Email      string `json:"email"`
	SecretCode int32  `json:"secret_code"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	logger.Infof("enqueued task: type - %s, payload - %v, queue - %s, max_retry - %d",
		task.Type(), payload, info.Queue, info.MaxRetry)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	err := processor.emailService.SendEmailMessage(
		payload.Email,
		processor.emailConfig.Templates.VerifyEmail,
		processor.emailConfig.Subjects.VerifyEmail,
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	logger.Infof("processed task: type - %s, payload - %v, email - %s",
		task.Type(), payload, payload.Email)
	return nil
}
