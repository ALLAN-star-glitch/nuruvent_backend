// internal/modules/notification/notification-domain/worker.go

package notificationdomain

import "context"

// TaskProcessor is the inbound port for task processing
// The worker implements this interface to process tasks from the queue
type TaskProcessor interface {
	// ProcessVerificationOTP processes a verification OTP task for any purpose
	// The purpose field determines the OTP type:
	//   - PurposeRegistration: New user registration
	//   - PurposeTwoFactor: Login 2FA
	//   - PurposePasswordReset: Password reset
	//   - PurposeEmailChange: Email change verification
	//   - PurposePhoneChange: Phone change verification
	ProcessVerificationOTP(ctx context.Context, task VerificationOTPTask) error

	// ProcessWelcomeIndividual processes an individual welcome task
	ProcessWelcomeIndividual(ctx context.Context, task WelcomeIndividualTask) error

	// ProcessWelcomeInstitution processes an institution welcome task
	ProcessWelcomeInstitution(ctx context.Context, task WelcomeInstitutionTask) error

	// ProcessPasswordResetConfirm processes a password reset confirmation task
	ProcessPasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error

	// ProcessLoginNotification processes a login notification task
	ProcessLoginNotification(ctx context.Context, task LoginNotificationTask) error
}