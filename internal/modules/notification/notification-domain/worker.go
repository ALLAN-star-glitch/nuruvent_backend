// internal/modules/notification/notification-domain/worker.go

package notificationdomain

import "context"

// TaskProcessor is the inbound port for task processing
// The worker implements this interface to process tasks from the queue
type TaskProcessor interface {
	// ProcessVerificationOTP processes a verification OTP task
	ProcessVerificationOTP(ctx context.Context, task VerificationOTPTask) error
	
	// ProcessWelcomeIndividual processes an individual welcome task
	ProcessWelcomeIndividual(ctx context.Context, task WelcomeIndividualTask) error
	
	// ProcessWelcomeInstitution processes an institution welcome task
	ProcessWelcomeInstitution(ctx context.Context, task WelcomeInstitutionTask) error
	
	// ProcessTwoFactorOTP processes a 2FA OTP task
	ProcessTwoFactorOTP(ctx context.Context, task TwoFactorOTPTask) error
	
	// ProcessPasswordResetOTP processes a password reset OTP task
	ProcessPasswordResetOTP(ctx context.Context, task PasswordResetOTPTask) error
	
	// ProcessPasswordResetConfirm processes a password reset confirmation task
	ProcessPasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error
	
	// ProcessLoginNotification processes a login notification task
	ProcessLoginNotification(ctx context.Context, task LoginNotificationTask) error
}