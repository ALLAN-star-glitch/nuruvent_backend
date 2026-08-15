// internal/shared/queue/client.go

package queue

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

type client struct {
	client *asynq.Client
}

func NewClient(cfg *config.Config) notificationdomain.TaskQueue {
	c := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.URL})
	return &client{client: c}
}

// Enqueue implements notificationdomain.TaskQueue
func (c *client) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	task := asynq.NewTask(taskType, payload)
	info, err := c.client.Enqueue(task)
	if err != nil {
		log.Printf("Failed to enqueue task %s: %v", taskType, err)
		return err
	}
	log.Printf("Task enqueued: ID=%s, Type=%s", info.ID, info.Type)
	return nil
}

// EnqueueDelayed implements notificationdomain.TaskQueue
func (c *client) EnqueueDelayed(ctx context.Context, taskType string, payload []byte, delaySeconds int) error {
	task := asynq.NewTask(taskType, payload)
	info, err := c.client.Enqueue(task, asynq.ProcessIn(time.Duration(delaySeconds)*time.Second))
	if err != nil {
		log.Printf("Failed to enqueue delayed task %s: %v", taskType, err)
		return err
	}
	log.Printf("Delayed task enqueued: ID=%s, Type=%s, Delay=%ds", info.ID, info.Type, delaySeconds)
	return nil
}

func (c *client) Close() error {
	return c.client.Close()
}