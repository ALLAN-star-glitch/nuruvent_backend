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

	// ProcessWelcomeInstitutionKYC processes an institution KYC welcome task
	ProcessWelcomeInstitutionKYC(ctx context.Context, task WelcomeInstitutionKYCTask) error

	// ProcessPasswordResetConfirm processes a password reset confirmation task
	ProcessPasswordResetConfirm(ctx context.Context, task PasswordResetConfirmTask) error

	// ProcessLoginNotification processes a login notification task
	ProcessLoginNotification(ctx context.Context, task LoginNotificationTask) error

	// ProcessNewInstitutionAccountRegistration processes new institution account registration notice
	ProcessNewInstitutionAccountRegistration(ctx context.Context, task NewInstitutionAccountRegistrationNotice) error

	// ProcessNewPersonalAccountRegistration processes new personal account registration notice
	ProcessNewPersonalAccountRegistration(ctx context.Context, task NewPersonalAccountRegistrationTask) error

	// ============================================================
	// ✅ TEAM INVITATION TASKS (UPDATED)
	// ============================================================

	// ProcessTeamInviteExistingUser processes a team invitation for existing users
	// Sent when: User already has a Nuruvent account
	// ✅ NO ROLE - Roles are inherited from account level
	// ✅ NO OTP - User clicks accept link to join
	ProcessTeamInviteExistingUser(ctx context.Context, task TeamInviteExistingUserTask) error

	// ProcessTeamInviteRegistration processes a team invitation for new users
	// Sent when: User does NOT have a Nuruvent account
	// ✅ NO ROLE - Roles are inherited from account level
	// ✅ NO OTP - User clicks registration link with token embedded
	ProcessTeamInviteRegistration(ctx context.Context, task TeamInviteRegistrationTask) error

	// ProcessTeamInviteAccepted processes a notification when invitation is accepted
	// Sent to: Admin who sent the invitation
	ProcessTeamInviteAccepted(ctx context.Context, task TeamInviteAcceptedTask) error

	// ProcessTeamInviteDeclined processes a notification when invitation is declined
	// Sent to: Admin who sent the invitation
	ProcessTeamInviteDeclined(ctx context.Context, task TeamInviteDeclinedTask) error
}