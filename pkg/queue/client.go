package queue

import (
    "log"

    "github.com/hibiken/asynq"
)

type Client struct {
    client *asynq.Client
}

func NewClient(redisAddr string) *Client {
    client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
    return &Client{client: client}
}

func (c *Client) Enqueue(taskType string, payload []byte, opts ...asynq.Option) error {
    task := asynq.NewTask(taskType, payload, opts...)
    info, err := c.client.Enqueue(task)
    if err != nil {
        log.Printf("Failed to enqueue task: %v", err)
        return err
    }
    log.Printf("Task enqueued: ID=%s, Type=%s", info.ID, info.Type)
    return nil
}

func (c *Client) Close() error {
    return c.client.Close()
}