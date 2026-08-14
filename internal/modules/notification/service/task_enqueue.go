package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/queue"
)

// TaskEnqueuer handles encoding and enqueueing tasks
type TaskEnqueuer struct {
	queueClient *queue.Client
}

func NewTaskEnqueuer(queueClient *queue.Client) *TaskEnqueuer {
	return &TaskEnqueuer{queueClient: queueClient}
}

// ============================================================
// ENQUEUE METHODS
// ============================================================

// EnqueueVerificationOTP enqueues a verification OTP task
func (e *TaskEnqueuer) EnqueueVerificationOTP(ctx context.Context, task notificationdomain.VerificationOTPTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskVerificationOTP, payload)
}

// EnqueueWelcomeIndividual enqueues an individual welcome task
func (e *TaskEnqueuer) EnqueueWelcomeIndividual(ctx context.Context, task notificationdomain.WelcomeIndividualTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskWelcomeIndividual, payload)
}

// EnqueueWelcomeInstitution enqueues an institution welcome task
func (e *TaskEnqueuer) EnqueueWelcomeInstitution(ctx context.Context, task notificationdomain.WelcomeInstitutionTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskWelcomeInstitution, payload)
}

// EnqueueTwoFactorOTP enqueues a 2FA OTP task
func (e *TaskEnqueuer) EnqueueTwoFactorOTP(ctx context.Context, task notificationdomain.TwoFactorOTPTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskTwoFactorOTP, payload)
}

// EnqueuePasswordResetOTP enqueues a password reset OTP task
func (e *TaskEnqueuer) EnqueuePasswordResetOTP(ctx context.Context, task notificationdomain.PasswordResetOTPTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskPasswordResetOTP, payload)
}

// EnqueuePasswordResetConfirm enqueues a password reset confirmation task
func (e *TaskEnqueuer) EnqueuePasswordResetConfirm(ctx context.Context, task notificationdomain.PasswordResetConfirmTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskPasswordResetConfirm, payload)
}

// EnqueueLoginNotification enqueues a login notification task
func (e *TaskEnqueuer) EnqueueLoginNotification(ctx context.Context, task notificationdomain.LoginNotificationTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	return e.queueClient.Enqueue(notificationdomain.TaskLoginNotification, payload)
}