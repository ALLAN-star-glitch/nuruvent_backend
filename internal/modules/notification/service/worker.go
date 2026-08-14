package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationWorker handles async notification tasks
type NotificationWorker struct {
	service notificationdomain.NotificationService
}

func NewNotificationWorker(svc notificationdomain.NotificationService) *NotificationWorker {
	return &NotificationWorker{
		service: svc,
	}
}

// ============================================================
// VERIFICATION OTP HANDLER
// ============================================================

// HandleVerificationOTP handles verification OTP tasks
func (w *NotificationWorker) HandleVerificationOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.VerificationOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse verification OTP task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing verification OTP for %s", data.To)

	req := notificationdomain.SendOTPRequest{
		To:      data.To,
		Name:    data.Name,
		OTP:     data.OTP,
		Expires: data.Expires,
		Purpose: data.Purpose,
		Meta:    data.Meta,
	}

	if err := w.service.SendVerificationOTP(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send verification OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Verification OTP sent to %s", data.To)
	return nil
}

// ============================================================
// WELCOME HANDLERS
// ============================================================

// HandleWelcomeIndividual handles individual welcome tasks
func (w *NotificationWorker) HandleWelcomeIndividual(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeIndividualTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse individual welcome task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing individual welcome for %s", data.To)

	req := notificationdomain.SendWelcomeRequest{
		To:   data.To,
		Name: data.Name,
	}

	if err := w.service.SendIndividualWelcome(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send individual welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Individual welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeInstitution handles institution welcome tasks
func (w *NotificationWorker) HandleWelcomeInstitution(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeInstitutionTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse institution welcome task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing institution welcome for %s - Institution: %s", data.To, data.InstitutionName)

	req := notificationdomain.SendInstitutionWelcomeRequest{
		To:              data.To,
		AdminName:       data.AdminName,
		InstitutionName: data.InstitutionName,
	}

	if err := w.service.SendInstitutionWelcome(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send institution welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Institution welcome sent to %s", data.To)
	return nil
}

// ============================================================
// TWO FACTOR OTP HANDLER
// ============================================================

// HandleTwoFactorOTP handles 2FA OTP tasks
func (w *NotificationWorker) HandleTwoFactorOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TwoFactorOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse 2FA OTP task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing 2FA OTP for %s", data.To)

	req := notificationdomain.SendTwoFactorRequest{
		To:        data.To,
		Name:      data.Name,
		OTP:       data.OTP,
		Expires:   data.Expires,
		IPAddress: data.IPAddress,
		UserAgent: data.UserAgent,
	}

	if err := w.service.SendTwoFactorOTP(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send 2FA OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] 2FA OTP sent to %s", data.To)
	return nil
}

// ============================================================
// PASSWORD RESET HANDLERS
// ============================================================

// HandlePasswordResetOTP handles password reset OTP tasks
func (w *NotificationWorker) HandlePasswordResetOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.PasswordResetOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse password reset OTP task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing password reset OTP for %s", data.To)

	req := notificationdomain.SendOTPRequest{
		To:      data.To,
		Name:    data.Name,
		OTP:     data.OTP,
		Expires: data.Expires,
		Purpose: notificationdomain.PurposePasswordReset,
		Meta:    nil,
	}

	if err := w.service.SendPasswordResetOTP(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send password reset OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Password reset OTP sent to %s", data.To)
	return nil
}

// HandlePasswordResetConfirm handles password reset confirmation tasks
func (w *NotificationWorker) HandlePasswordResetConfirm(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.PasswordResetConfirmTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse password reset confirm task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing password reset confirmation for %s", data.To)

	req := notificationdomain.SendPasswordResetConfirmRequest{
		To:   data.To,
		Name: data.Name,
	}

	if err := w.service.SendPasswordResetConfirm(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send password reset confirmation to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Password reset confirmation sent to %s", data.To)
	return nil
}

// ============================================================
// LOGIN NOTIFICATION HANDLER
// ============================================================

// HandleLoginNotification handles login notification tasks
func (w *NotificationWorker) HandleLoginNotification(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.LoginNotificationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse login notification task: %v", err)
		return err
	}

	log.Printf("[NotificationWorker] Processing login notification for %s", data.To)

	req := notificationdomain.SendLoginNotificationRequest{
		To:        data.To,
		Name:      data.Name,
		Time:      data.Time,
		IPAddress: data.IPAddress,
		UserAgent: data.UserAgent,
	}

	if err := w.service.SendLoginNotification(ctx, req); err != nil {
		log.Printf("[NotificationWorker] Failed to send login notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Login notification sent to %s", data.To)
	return nil
}