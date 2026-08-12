package email

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

type EmailWorker struct {
	emailService *EmailService
}

func NewEmailWorker(emailService *EmailService) *EmailWorker {
	return &EmailWorker{
		emailService: emailService,
	}
}

// ================================================
// VERIFICATION OTP HANDLER
// ================================================

// HandleVerificationOTP handles all verification OTP emails
func (w *EmailWorker) HandleVerificationOTP(ctx context.Context, task *asynq.Task) error {
	var data VerificationOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse verification OTP task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing %s verification OTP for %s", data.Purpose, data.To)

	if err := w.emailService.SendVerificationOTP(
		data.To,
		data.Name,
		data.OTP,
		data.Expires,
		string(data.Purpose),
		data.Meta,
	); err != nil {
		log.Printf("[EmailWorker] Failed to send verification OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Verification OTP sent to %s for purpose: %s", data.To, data.Purpose)
	return nil
}

// ================================================
// INDIVIDUAL WELCOME HANDLER
// ================================================

// HandleWelcomeIndividual handles individual professional welcome email
func (w *EmailWorker) HandleWelcomeIndividual(ctx context.Context, task *asynq.Task) error {
	var data WelcomeIndividualTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse individual welcome task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing individual welcome email for %s", data.To)

	if err := w.emailService.SendIndividualWelcome(data.To, data.Name); err != nil {
		log.Printf("[EmailWorker] Failed to send individual welcome email to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Individual welcome email sent to %s", data.To)
	return nil
}

// ================================================
// INSTITUTION WELCOME HANDLER
// ================================================

// HandleWelcomeInstitution handles institution welcome email
func (w *EmailWorker) HandleWelcomeInstitution(ctx context.Context, task *asynq.Task) error {
	var data WelcomeInstitutionTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse institution welcome task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing institution welcome email for %s - Institution: %s", data.To, data.InstitutionName)

	if err := w.emailService.SendInstitutionWelcome(data.To, data.AdminName, data.InstitutionName); err != nil {
		log.Printf("[EmailWorker] Failed to send institution welcome email to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Institution welcome email sent to %s", data.To)
	return nil
}

// ================================================
// TWO-FACTOR OTP HANDLER
// ================================================

// HandleTwoFactorOTP handles 2FA OTP email
func (w *EmailWorker) HandleTwoFactorOTP(ctx context.Context, task *asynq.Task) error {
	var data TwoFactorOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse 2FA OTP task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing 2FA OTP for %s", data.To)

	if err := w.emailService.SendTwoFactorOTP(data.To, data.Name, data.OTP, data.Expires); err != nil {
		log.Printf("[EmailWorker] Failed to send 2FA OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] 2FA OTP sent to %s", data.To)
	return nil
}

// ================================================
// PASSWORD RESET OTP HANDLER
// ================================================

// HandlePasswordResetOTP handles password reset OTP
func (w *EmailWorker) HandlePasswordResetOTP(ctx context.Context, task *asynq.Task) error {
	var data PasswordResetOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse password reset OTP task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing password reset OTP for %s", data.To)

	if err := w.emailService.SendPasswordResetOTP(data.To, data.Name, data.OTP, data.Expires); err != nil {
		log.Printf("[EmailWorker] Failed to send password reset OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Password reset OTP sent to %s", data.To)
	return nil
}

// ================================================
// PASSWORD RESET CONFIRM HANDLER
// ================================================

// HandlePasswordResetConfirm handles password reset confirmation
func (w *EmailWorker) HandlePasswordResetConfirm(ctx context.Context, task *asynq.Task) error {
	var data PasswordResetConfirmTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse password reset confirm task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing password reset confirmation for %s", data.To)

	if err := w.emailService.SendPasswordResetConfirm(data.To, data.Name); err != nil {
		log.Printf("[EmailWorker] Failed to send password reset confirmation to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Password reset confirmation sent to %s", data.To)
	return nil
}

// ================================================
// LOGIN NOTIFICATION HANDLER
// ================================================

// HandleLoginNotification handles login notification
func (w *EmailWorker) HandleLoginNotification(ctx context.Context, task *asynq.Task) error {
	var data LoginNotificationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[EmailWorker] Failed to parse login notification task: %v", err)
		return err
	}

	log.Printf("[EmailWorker] Processing login notification for %s", data.To)

	if err := w.emailService.SendLoginNotification(
		data.To,
		data.Name,
		data.Time,
		data.IPAddress,
		data.UserAgent,
	); err != nil {
		log.Printf("[EmailWorker] Failed to send login notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[EmailWorker] Login notification sent to %s", data.To)
	return nil
}