package auth

import (
	"context"
	"encoding/json"
	"fmt"

	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// QueueAdapter adapts notificationdomain.TaskQueue to authDomain.QueueService
type QueueAdapter struct {
	queue notificationdomain.TaskQueue
}

func NewQueueAdapter(queue notificationdomain.TaskQueue) authDomain.QueueService {
	return &QueueAdapter{queue: queue}
}

func (a *QueueAdapter) Enqueue(ctx context.Context, task string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return a.queue.Enqueue(ctx, task, data)
}

func (a *QueueAdapter) EnqueueDelayed(ctx context.Context, task string, payload any, delaySeconds int) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return a.queue.EnqueueDelayed(ctx, task, data, delaySeconds)
}
