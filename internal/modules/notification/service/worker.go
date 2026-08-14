// internal/modules/notification/service/worker.go

package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationWorker handles async notification tasks
// ✅ Implements notificationdomain.TaskProcessor (inbound port)
// ✅ Depends on notificationdomain.Channel (outbound port)
type NotificationWorker struct {
	emailChannel notificationdomain.Channel // ← Outbound Port (interface)
}

// NewNotificationWorker creates a new notification worker
// ✅ Injects the outbound port dependency
func NewNotificationWorker(emailChannel notificationdomain.Channel) *NotificationWorker {
	return &NotificationWorker{
		emailChannel: emailChannel,
	}
}

// ============================================================
// VERIFICATION OTP HANDLER
// ============================================================

// ProcessVerificationOTP implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessVerificationOTP(ctx context.Context, data notificationdomain.VerificationOTPTask) error {
	log.Printf("[NotificationWorker] Processing verification OTP for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Verify Your Account - Nuruvent",
		Type:    notificationdomain.TypeVerificationOTP,
		Meta: map[string]string{
			"name":        data.Name,
			"otp":         data.OTP,
			"expires":     data.Expires,
			"title":       "Verify Your",
			"subtitle":    "Account",
			"description": "Complete your Nuruvent registration",
			"message":     "Thank you for joining Nuruvent. Please use the verification code below to complete your account setup.",
			"warning":     "If you did not create an account on Nuruvent, please ignore this email.",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send verification OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Verification OTP sent to %s", data.To)
	return nil
}

// HandleVerificationOTP is the asynq task handler
func (w *NotificationWorker) HandleVerificationOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.VerificationOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse verification OTP task: %v", err)
		return err
	}
	return w.ProcessVerificationOTP(ctx, data)
}

// ============================================================
// WELCOME HANDLERS
// ============================================================

// ProcessWelcomeIndividual implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessWelcomeIndividual(ctx context.Context, data notificationdomain.WelcomeIndividualTask) error {
	log.Printf("[NotificationWorker] Processing individual welcome for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Welcome to Nuruvent - Your Professional Account is Ready!",
		Type:    notificationdomain.TypeWelcome,
		Meta: map[string]string{
			"name":         data.Name,
			"account_type": "individual",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send individual welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Individual welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeIndividual is the asynq task handler
func (w *NotificationWorker) HandleWelcomeIndividual(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeIndividualTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse individual welcome task: %v", err)
		return err
	}
	return w.ProcessWelcomeIndividual(ctx, data)
}

// ProcessWelcomeInstitution implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessWelcomeInstitution(ctx context.Context, data notificationdomain.WelcomeInstitutionTask) error {
	log.Printf("[NotificationWorker] Processing institution welcome for %s - Institution: %s", data.To, data.InstitutionName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Welcome to Nuruvent - Your Institution is Live!",
		Type:    notificationdomain.TypeWelcome,
		Meta: map[string]string{
			"admin_name":        data.AdminName,
			"institution_name":  data.InstitutionName,
			"account_type":      "institution",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send institution welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Institution welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeInstitution is the asynq task handler
func (w *NotificationWorker) HandleWelcomeInstitution(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeInstitutionTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse institution welcome task: %v", err)
		return err
	}
	return w.ProcessWelcomeInstitution(ctx, data)
}

// ============================================================
// TWO FACTOR OTP HANDLER
// ============================================================

// ProcessTwoFactorOTP implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessTwoFactorOTP(ctx context.Context, data notificationdomain.TwoFactorOTPTask) error {
	log.Printf("[NotificationWorker] Processing 2FA OTP for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Two-Factor Authentication - Nuruvent",
		Type:    notificationdomain.TypeTwoFactor,
		Meta: map[string]string{
			"name":       data.Name,
			"otp":        data.OTP,
			"expires":    data.Expires,
			"ip_address": data.IPAddress,
			"user_agent": data.UserAgent,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send 2FA OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] 2FA OTP sent to %s", data.To)
	return nil
}

// HandleTwoFactorOTP is the asynq task handler
func (w *NotificationWorker) HandleTwoFactorOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TwoFactorOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse 2FA OTP task: %v", err)
		return err
	}
	return w.ProcessTwoFactorOTP(ctx, data)
}

// ============================================================
// PASSWORD RESET HANDLERS
// ============================================================

// ProcessPasswordResetOTP implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessPasswordResetOTP(ctx context.Context, data notificationdomain.PasswordResetOTPTask) error {
	log.Printf("[NotificationWorker] Processing password reset OTP for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Reset Your Password - Nuruvent",
		Type:    notificationdomain.TypeVerificationOTP,
		Meta: map[string]string{
			"name":        data.Name,
			"otp":         data.OTP,
			"expires":     data.Expires,
			"title":       "Reset Your",
			"subtitle":    "Password",
			"description": "Secure access to your account",
			"message":     "We received a request to reset your password for your Nuruvent account. Use the verification code below to continue.",
			"warning":     "If you did not request a password reset, please ignore this email. Your account remains secure.",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send password reset OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Password reset OTP sent to %s", data.To)
	return nil
}

// HandlePasswordResetOTP is the asynq task handler
func (w *NotificationWorker) HandlePasswordResetOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.PasswordResetOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse password reset OTP task: %v", err)
		return err
	}
	return w.ProcessPasswordResetOTP(ctx, data)
}

// ProcessPasswordResetConfirm implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessPasswordResetConfirm(ctx context.Context, data notificationdomain.PasswordResetConfirmTask) error {
	log.Printf("[NotificationWorker] Processing password reset confirmation for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Password Reset Confirmation - Nuruvent",
		Type:    notificationdomain.TypePasswordResetConfirm,
		Meta: map[string]string{
			"name": data.Name,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send password reset confirmation to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Password reset confirmation sent to %s", data.To)
	return nil
}

// HandlePasswordResetConfirm is the asynq task handler
func (w *NotificationWorker) HandlePasswordResetConfirm(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.PasswordResetConfirmTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse password reset confirm task: %v", err)
		return err
	}
	return w.ProcessPasswordResetConfirm(ctx, data)
}

// ============================================================
// LOGIN NOTIFICATION HANDLER
// ============================================================

// ProcessLoginNotification implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessLoginNotification(ctx context.Context, data notificationdomain.LoginNotificationTask) error {
	log.Printf("[NotificationWorker] Processing login notification for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "New Login Notification - Nuruvent",
		Type:    notificationdomain.TypeLoginNotification,
		Meta: map[string]string{
			"name":       data.Name,
			"time":       data.Time,
			"ip_address": data.IPAddress,
			"user_agent": data.UserAgent,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send login notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Login notification sent to %s", data.To)
	return nil
}

// HandleLoginNotification is the asynq task handler
func (w *NotificationWorker) HandleLoginNotification(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.LoginNotificationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse login notification task: %v", err)
		return err
	}
	return w.ProcessLoginNotification(ctx, data)
}

// Ensure NotificationWorker implements notificationdomain.TaskProcessor
var _ notificationdomain.TaskProcessor = (*NotificationWorker)(nil)