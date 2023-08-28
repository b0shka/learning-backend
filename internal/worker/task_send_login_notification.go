package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/b0shka/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

const TaskSendLoginNotification = "task:send_login_notification"

type PayloadSendLoginNotification struct {
	Email     string `json:"email"`
	UserAgent string `json:"user_agent"`
	ClientIp  string `json:"client_ip"`
	Time      string `json:"time"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendLoginNotification(
	ctx context.Context,
	payload *PayloadSendLoginNotification,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendLoginNotification, jsonPayload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	logger.Infof("enqueued task: type - %s, payload - %v, queue - %s, max_retry - %d",
		task.Type(), payload, info.Queue, info.MaxRetry)
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendLoginNotification(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendLoginNotification
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	err := processor.emailService.SendEmailMessage(
		payload.Email,
		processor.emailConfig.Templates.LoginNotification,
		processor.emailConfig.Subjects.LoginNotification,
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	logger.Infof("processed task: type - %s, payload - %v, email - %s",
		task.Type(), payload, payload.Email)
	return nil
}
