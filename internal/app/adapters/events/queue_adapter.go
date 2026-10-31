package events

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// QueueAdapter bridges notificationdomain.TaskQueue to domain.TaskQueue.
//
// The two interfaces have identical method signatures, but Go treats
// them as distinct types. This adapter is a one-line forwarder.
type QueueAdapter struct {
	inner notificationdomain.TaskQueue
}

func NewQueueAdapter(inner notificationdomain.TaskQueue) domain.TaskQueue {
	return &QueueAdapter{inner: inner}
}

func (a *QueueAdapter) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	return a.inner.Enqueue(ctx, taskType, payload)
}