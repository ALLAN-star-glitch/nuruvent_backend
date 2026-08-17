// internal/modules/notification/service/task_enqueuer.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// ✅ Unexported concrete type
type taskEnqueuer struct {
	queue notificationdomain.TaskQueue // ← Outbound Port (interface)
}

// ✅ Constructor returns the interface
func NewTaskEnqueuer(queue notificationdomain.TaskQueue) notificationdomain.TaskEnqueuer {
	return &taskEnqueuer{
		queue: queue,
	}
}

// ============================================================
// ENQUEUE METHODS
// ============================================================

// EnqueueVerificationOTP enqueues a verification OTP task for any purpose
// The purpose field determines the OTP type:
//   - PurposeRegistration: New user registration
//   - PurposeTwoFactor: Login 2FA
//   - PurposePasswordReset: Password reset
//   - PurposeEmailChange: Email change verification
//   - PurposePhoneChange: Phone change verification
func (e *taskEnqueuer) EnqueueVerificationOTP(ctx context.Context, task notificationdomain.VerificationOTPTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing verification OTP for %s (purpose: %s)", task.To, task.Purpose)

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := e.queue.Enqueue(ctx, notificationdomain.TaskVerificationOTP, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue verification OTP: %v", err)
		return err
	}

	log.Printf("✅ [TaskEnqueuer] Verification OTP enqueued for %s (purpose: %s)", task.To, task.Purpose)
	return nil
}

// EnqueueWelcomeIndividual enqueues an individual welcome task
func (e *taskEnqueuer) EnqueueWelcomeIndividual(ctx context.Context, task notificationdomain.WelcomeIndividualTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing individual welcome for %s", task.To)

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := e.queue.Enqueue(ctx, notificationdomain.TaskWelcomeIndividual, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue individual welcome: %v", err)
		return err
	}

	log.Printf("✅ [TaskEnqueuer] Individual welcome enqueued for %s", task.To)
	return nil
}

// EnqueueWelcomeInstitution enqueues an institution welcome task
func (e *taskEnqueuer) EnqueueWelcomeInstitution(ctx context.Context, task notificationdomain.WelcomeInstitutionTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing institution welcome for %s", task.To)

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := e.queue.Enqueue(ctx, notificationdomain.TaskWelcomeInstitution, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue institution welcome: %v", err)
		return err
	}

	log.Printf("✅ [TaskEnqueuer] Institution welcome enqueued for %s", task.To)
	return nil
}

// EnqueuePasswordResetConfirm enqueues a password reset confirmation task
func (e *taskEnqueuer) EnqueuePasswordResetConfirm(ctx context.Context, task notificationdomain.PasswordResetConfirmTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing password reset confirmation for %s", task.To)

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := e.queue.Enqueue(ctx, notificationdomain.TaskPasswordResetConfirm, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue password reset confirmation: %v", err)
		return err
	}

	log.Printf("✅ [TaskEnqueuer] Password reset confirmation enqueued for %s", task.To)
	return nil
}

// EnqueueLoginNotification enqueues a login notification task
func (e *taskEnqueuer) EnqueueLoginNotification(ctx context.Context, task notificationdomain.LoginNotificationTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing login notification for %s", task.To)

	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := e.queue.Enqueue(ctx, notificationdomain.TaskLoginNotification, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue login notification: %v", err)
		return err
	}

	log.Printf("✅ [TaskEnqueuer] Login notification enqueued for %s", task.To)
	return nil
}

// Ensure taskEnqueuer implements notificationdomain.TaskEnqueuer
var _ notificationdomain.TaskEnqueuer = (*taskEnqueuer)(nil)