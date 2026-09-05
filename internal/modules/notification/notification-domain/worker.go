// internal/modules/notification/notification-domain/worker.go

package notificationdomain

import "context"

// TaskProcessor is the inbound port for task processing
type TaskProcessor interface {
	// ProcessVerificationOTP processes a verification OTP task for any purpose
	ProcessVerificationOTP(ctx context.Context, task VerificationOTPTask) error

	// ProcessWelcomeIndividual processes an individual welcome task
	ProcessWelcomeIndividual(ctx context.Context, task WelcomeIndividualTask) error

	// ProcessWelcomeInstitution processes an institution welcome task
	ProcessWelcomeInstitution(ctx context.Context, task WelcomeInstitutionTask) error

	// NEW: ProcessWelcomeInstitutionKYC processes an institution KYC welcome task
	ProcessWelcomeInstitutionKYC(ctx context.Context, task WelcomeInstitutionKYCTask) error

	// ProcessPasswordResetConfirm processes a password reset confirmation task
	ProcessPasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error

	// ProcessLoginNotification processes a login notification task
	ProcessLoginNotification(ctx context.Context, task LoginNotificationTask) error

	// ============================================================
	//  TEAM INVITATION TASKS
	// ============================================================

	// ProcessTeamInvite processes a team invitation email task
	ProcessTeamInvite(ctx context.Context, task TeamInviteTask) error

	// ProcessTeamInviteRegistration processes a team invite registration OTP task
	ProcessTeamInviteRegistration(ctx context.Context, task TeamInviteRegistrationTask) error

	// ProcessTeamInviteAccepted processes a team invite accepted notification task
	ProcessTeamInviteAccepted(ctx context.Context, task TeamInviteAcceptedTask) error

	// ProcessTeamInviteDeclined processes a team invite declined notification task
	ProcessTeamInviteDeclined(ctx context.Context, task TeamInviteDeclinedTask) error
}