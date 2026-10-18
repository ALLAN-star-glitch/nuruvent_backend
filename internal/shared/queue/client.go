package queue

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// NewAsynqClient is the raw client — the single source of truth.
func NewAsynqClient(cfg *config.Config) (*asynq.Client, error) {
	opt, err := asynq.ParseRedisURI(cfg.GetRedisURL())
	if err != nil {
		return nil, err
	}
	return asynq.NewClient(opt), nil
}

// TaskQueueWrapper wraps *asynq.Client as the notification TaskQueue.
type TaskQueueWrapper struct {
	client *asynq.Client
}

func NewTaskQueue(client *asynq.Client) notificationdomain.TaskQueue {
	return &TaskQueueWrapper{client: client}
}

func (w *TaskQueueWrapper) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	task := asynq.NewTask(taskType, payload)
	info, err := w.client.Enqueue(task)
	if err != nil {
		log.Printf("Failed to enqueue task %s: %v", taskType, err)
		return err
	}
	log.Printf("Task enqueued: ID=%s, Type=%s", info.ID, info.Type)
	return nil
}

func (w *TaskQueueWrapper) EnqueueDelayed(ctx context.Context, taskType string, payload []byte, delaySeconds int) error {
	task := asynq.NewTask(taskType, payload)
	info, err := w.client.Enqueue(task, asynq.ProcessIn(time.Duration(delaySeconds)*time.Second))
	if err != nil {
		log.Printf("Failed to enqueue delayed task %s: %v", taskType, err)
		return err
	}
	log.Printf("Delayed task enqueued: ID=%s, Type=%s, Delay=%ds", info.ID, info.Type, delaySeconds)
	return nil
}




func (w *TaskQueueWrapper) Close() error {
	return w.client.Close()
}