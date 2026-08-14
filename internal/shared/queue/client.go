// internal/shared/queue/client.go

package queue

import (
	"context"
	"log"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// Client implements notificationdomain.TaskQueue (outbound port)
type Client struct {
	client *asynq.Client
}

// NewClient creates a new queue client
func NewClient(redisAddr string) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Client{client: client}
}

// Enqueue implements notificationdomain.TaskQueue
func (c *Client) Enqueue(ctx context.Context, taskType string, payload []byte, opts ...interface{}) error {
	task := asynq.NewTask(taskType, payload)
	info, err := c.client.Enqueue(task)
	if err != nil {
		log.Printf("Failed to enqueue task %s: %v", taskType, err)
		return err
	}
	log.Printf("Task enqueued: ID=%s, Type=%s", info.ID, info.Type)
	return nil
}

// Close closes the queue client
func (c *Client) Close() error {
	return c.client.Close()
}

// Ensure Client implements notificationdomain.TaskQueue
var _ notificationdomain.TaskQueue = (*Client)(nil)