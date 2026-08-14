// internal/modules/notification/service/task_enqueuer.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// TaskEnqueuer handles encoding and enqueueing tasks
// ✅ Implements notificationdomain.TaskEnqueuer (inbound port)
// ✅ Depends on notificationdomain.TaskQueue (outbound port)
type TaskEnqueuer struct {
	queue notificationdomain.TaskQueue // ← Outbound Port (interface)
}

// NewTaskEnqueuer creates a new task enqueuer
// ✅ Injects the outbound port dependency
func NewTaskEnqueuer(queue notificationdomain.TaskQueue) *TaskEnqueuer {
	return &TaskEnqueuer{
		queue: queue,
	}
}

// ============================================================
// ENQUEUE METHODS
// ============================================================

// EnqueueVerificationOTP enqueues a verification OTP task
func (e *TaskEnqueuer) EnqueueVerificationOTP(ctx context.Context, task notificationdomain.VerificationOTPTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing verification OTP for %s", task.To)
	
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	
	if err := e.queue.Enqueue(ctx, notificationdomain.TaskVerificationOTP, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue verification OTP: %v", err)
		return err
	}
	
	log.Printf("✅ [TaskEnqueuer] Verification OTP enqueued for %s", task.To)
	return nil
}

// EnqueueWelcomeIndividual enqueues an individual welcome task
func (e *TaskEnqueuer) EnqueueWelcomeIndividual(ctx context.Context, task notificationdomain.WelcomeIndividualTask) error {
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
func (e *TaskEnqueuer) EnqueueWelcomeInstitution(ctx context.Context, task notificationdomain.WelcomeInstitutionTask) error {
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

// EnqueueTwoFactorOTP enqueues a 2FA OTP task
func (e *TaskEnqueuer) EnqueueTwoFactorOTP(ctx context.Context, task notificationdomain.TwoFactorOTPTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing 2FA OTP for %s", task.To)
	
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	
	if err := e.queue.Enqueue(ctx, notificationdomain.TaskTwoFactorOTP, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue 2FA OTP: %v", err)
		return err
	}
	
	log.Printf("✅ [TaskEnqueuer] 2FA OTP enqueued for %s", task.To)
	return nil
}

// EnqueuePasswordResetOTP enqueues a password reset OTP task
func (e *TaskEnqueuer) EnqueuePasswordResetOTP(ctx context.Context, task notificationdomain.PasswordResetOTPTask) error {
	log.Printf("📧 [TaskEnqueuer] Enqueuing password reset OTP for %s", task.To)
	
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	
	if err := e.queue.Enqueue(ctx, notificationdomain.TaskPasswordResetOTP, payload); err != nil {
		log.Printf("❌ [TaskEnqueuer] Failed to enqueue password reset OTP: %v", err)
		return err
	}
	
	log.Printf("✅ [TaskEnqueuer] Password reset OTP enqueued for %s", task.To)
	return nil
}

// EnqueuePasswordResetConfirm enqueues a password reset confirmation task
func (e *TaskEnqueuer) EnqueuePasswordResetConfirm(ctx context.Context, task notificationdomain.PasswordResetConfirmTask) error {
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
func (e *TaskEnqueuer) EnqueueLoginNotification(ctx context.Context, task notificationdomain.LoginNotificationTask) error {
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

// Ensure TaskEnqueuer implements notificationdomain.TaskEnqueuer
var _ notificationdomain.TaskEnqueuer = (*TaskEnqueuer)(nil)