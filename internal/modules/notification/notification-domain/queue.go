// internal/modules/notification/notification-domain/queue.go

package notificationdomain

import "context"

// TaskQueue is the outbound port for task queue operations
type TaskQueue interface {
	// Enqueue enqueues a task for processing
	Enqueue(ctx context.Context, taskType string, payload []byte, opts ...interface{}) error
}

// TaskEnqueuer is the inbound port for enqueuing tasks
// The service uses this to enqueue tasks
type TaskEnqueuer interface {
	// EnqueueVerificationOTP enqueues a verification OTP task
	EnqueueVerificationOTP(ctx context.Context, task VerificationOTPTask) error
	
	// EnqueueWelcomeIndividual enqueues an individual welcome task
	EnqueueWelcomeIndividual(ctx context.Context, task WelcomeIndividualTask) error
	
	// EnqueueWelcomeInstitution enqueues an institution welcome task
	EnqueueWelcomeInstitution(ctx context.Context, task WelcomeInstitutionTask) error
	
	// EnqueueTwoFactorOTP enqueues a 2FA OTP task
	EnqueueTwoFactorOTP(ctx context.Context, task TwoFactorOTPTask) error
	
	// EnqueuePasswordResetOTP enqueues a password reset OTP task
	EnqueuePasswordResetOTP(ctx context.Context, task PasswordResetOTPTask) error
	
	// EnqueuePasswordResetConfirm enqueues a password reset confirmation task
	EnqueuePasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error
	
	// EnqueueLoginNotification enqueues a login notification task
	EnqueueLoginNotification(ctx context.Context, task LoginNotificationTask) error
}