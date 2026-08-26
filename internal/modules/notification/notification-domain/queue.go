// internal/modules/notification/notification-domain/queue.go

package notificationdomain

import "context"

// ============================================================
// OUTBOUND PORT: TaskQueue Interface
// ============================================================

// TaskQueue is the outbound port for task queue operations
type TaskQueue interface {
	// Enqueue enqueues a task for processing
	Enqueue(ctx context.Context, taskType string, payload []byte) error

	// EnqueueDelayed enqueues a task with a delay
	EnqueueDelayed(ctx context.Context, taskType string, payload []byte, delaySeconds int) error
}

// ============================================================
// INBOUND PORT: TaskEnqueuer Interface
// ============================================================

// TaskEnqueuer is the inbound port for enqueuing tasks
type TaskEnqueuer interface {
	// EnqueueVerificationOTP enqueues a verification OTP task for any purpose
	EnqueueVerificationOTP(ctx context.Context, task VerificationOTPTask) error

	// EnqueueWelcomeIndividual enqueues an individual welcome task
	EnqueueWelcomeIndividual(ctx context.Context, task WelcomeIndividualTask) error

	// EnqueueWelcomeInstitution enqueues an institution welcome task
	EnqueueWelcomeInstitution(ctx context.Context, task WelcomeInstitutionTask) error

	// EnqueueWelcomeInstitutionKYC enqueues an institution KYC welcome task
	EnqueueWelcomeInstitutionKYC(ctx context.Context, task WelcomeInstitutionKYCTask) error

	EnqueueNewInstitutionAccountRegistration(ctx context.Context, task NewInstitutionAccountRegistrationNotice)	error

	EnqueueNewPersonalAccountRegistration(ctx context.Context, task NewPersonalAccountRegistrationTask) error

	// EnqueuePasswordResetConfirm enqueues a password reset confirmation task
	EnqueuePasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error

	// EnqueueLoginNotification enqueues a login notification task
	EnqueueLoginNotification(ctx context.Context, task LoginNotificationTask) error


}