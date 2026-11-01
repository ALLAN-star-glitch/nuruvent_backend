// internal/app/adapters/events/video_queue_adapter.go

package events

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// VideoQueueAdapter is a dedicated Asynq client for events-module
// background tasks. Every task it emits is routed to the isolated
// "video" queue, where the video provisioning worker (and only that
// worker) listens.
//
// This replaces the earlier approach of reusing the notification
// module's TaskQueue. That approach broke silently: the notification
// adapter has no concept of queue routing, so tasks landed in the
// "default" queue, where the notification worker pulled them, found
// no handler for "video:provision_meetings", and dropped them.
type VideoQueueAdapter struct {
	client *asynq.Client
}

func NewVideoQueueAdapter(client *asynq.Client) domain.TaskQueue {
	return &VideoQueueAdapter{client: client}
}

func (a *VideoQueueAdapter) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	task := asynq.NewTask(taskType, payload)
	_, err := a.client.EnqueueContext(ctx, task, asynq.Queue("video"))
	return err
}