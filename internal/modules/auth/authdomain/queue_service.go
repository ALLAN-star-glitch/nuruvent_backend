package authdomain

import "context"

// ============================================================
// OUTBOUND PORT: QueueService
// ============================================================

type QueueService interface {
    Enqueue(ctx context.Context, task string, payload any) error
    EnqueueDelayed(ctx context.Context, task string, payload any, delaySeconds int) error
}








