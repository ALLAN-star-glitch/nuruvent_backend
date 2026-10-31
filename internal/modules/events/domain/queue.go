package domain

import "context"

// TaskQueue is the outbound port the events module uses to enqueue
// background work. Implementations are provided by the infrastructure
// layer (asynq in production; a fake in tests).
type TaskQueue interface {
	Enqueue(ctx context.Context, taskType string, payload []byte) error
}