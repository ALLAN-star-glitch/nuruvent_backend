package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// notificationService implements notificationdomain.NotificationService
type notificationService struct {
	channels     map[notificationdomain.NotificationChannel]notificationdomain.Channel
	taskEnqueuer *TaskEnqueuer
	async        bool
}

// NewNotificationService creates a new notification service (synchronous)
func NewNotificationService(channels ...notificationdomain.Channel) notificationdomain.NotificationService {
	channelMap := make(map[notificationdomain.NotificationChannel]notificationdomain.Channel)
	for _, ch := range channels {
		channelMap[ch.GetChannel()] = ch
	}
	return &notificationService{
		channels: channelMap,
		async:    false,
	}
}

// NewNotificationServiceWithQueue creates a notification service with queue support
func NewNotificationServiceWithQueue(taskEnqueuer *TaskEnqueuer, channels ...notificationdomain.Channel) notificationdomain.NotificationService {
	channelMap := make(map[notificationdomain.NotificationChannel]notificationdomain.Channel)
	for _, ch := range channels {
		channelMap[ch.GetChannel()] = ch
	}
	return &notificationService{
		channels:     channelMap,
		taskEnqueuer: taskEnqueuer,
		async:        true,
	}
}

// getChannel returns the appropriate channel for sending
func (s *notificationService) getChannel(channel notificationdomain.NotificationChannel) (notificationdomain.Channel, error) {
	ch, ok := s.channels[channel]
	if !ok {
		return nil, fmt.Errorf("channel %s not configured", channel)
	}
	return ch, nil
}

// ============================================================
// VERIFICATION OTP
// ============================================================

func (s *notificationService) SendVerificationOTP(ctx context.Context, req notificationdomain.SendOTPRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.VerificationOTPTask{
			To:      req.To,
			Name:    req.Name,
			OTP:     req.OTP,
			Expires: req.Expires,
			Purpose: req.Purpose,
			Meta:    req.Meta,
		}
		if err := s.taskEnqueuer.EnqueueVerificationOTP(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue OTP task: %v, falling back to sync", err)
			return s.sendVerificationOTPSync(ctx, req)
		}
		return nil
	}
	return s.sendVerificationOTPSync(ctx, req)
}

func (s *notificationService) sendVerificationOTPSync(ctx context.Context, req notificationdomain.SendOTPRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	title, subtitle, description, message, extraInfo, warning := s.getVerificationContent(req)

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: title + " " + subtitle + " - Nuruvent",
		Type:    notificationdomain.TypeVerificationOTP,
		Meta: map[string]string{
			"title":       title,
			"subtitle":    subtitle,
			"description": description,
			"message":     message,
			"extra_info":  extraInfo,
			"warning":     warning,
			"otp":         req.OTP,
			"expires":     req.Expires,
			"name":        req.Name,
		},
	}

	return ch.Send(ctx, channelReq)
}

func (s *notificationService) getVerificationContent(req notificationdomain.SendOTPRequest) (title, subtitle, description, message, extraInfo, warning string) {
	switch req.Purpose {
	case notificationdomain.PurposeRegistration:
		title = "Verify Your"
		subtitle = "Account"
		description = "Complete your Nuruvent registration"
		message = "Thank you for joining Nuruvent. Please use the verification code below to complete your account setup."
		warning = "If you did not create an account on Nuruvent, please ignore this email."

	case notificationdomain.PurposeEmailChange:
		title = "Verify Your"
		subtitle = "Email Change"
		description = "Confirm your new email address"
		newEmail := req.Meta["new_email"]
		message = "You requested to change your email address to " + newEmail + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case notificationdomain.PurposePhoneChange:
		title = "Verify Your"
		subtitle = "Phone Change"
		description = "Confirm your new phone number"
		newPhone := req.Meta["new_phone"]
		message = "You requested to change your phone number to " + newPhone + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case notificationdomain.PurposePasswordReset:
		title = "Reset Your"
		subtitle = "Password"
		description = "Secure access to your account"
		message = "We received a request to reset your password for your Nuruvent account. Use the verification code below to continue."
		warning = "If you did not request a password reset, please ignore this email. Your account remains secure."

	case notificationdomain.PurposeTwoFactor:
		title = "Two-Factor"
		subtitle = "Authentication"
		description = "Secure your account access"
		message = "You requested a two-factor authentication code for your Nuruvent account. Please use the verification code below to complete your login."
		warning = "If you did not attempt to log in, please reset your password immediately."

	default:
		title = "Verify Your"
		subtitle = "Account"
		description = "Complete your verification"
		message = "Please use the verification code below to complete your verification."
	}
	return
}

// ============================================================
// WELCOME EMAILS
// ============================================================

func (s *notificationService) SendIndividualWelcome(ctx context.Context, req notificationdomain.SendWelcomeRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.WelcomeIndividualTask{
			To:   req.To,
			Name: req.Name,
		}
		if err := s.taskEnqueuer.EnqueueWelcomeIndividual(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue individual welcome task: %v, falling back to sync", err)
			return s.sendIndividualWelcomeSync(ctx, req)
		}
		return nil
	}
	return s.sendIndividualWelcomeSync(ctx, req)
}

func (s *notificationService) sendIndividualWelcomeSync(ctx context.Context, req notificationdomain.SendWelcomeRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: "Welcome to Nuruvent - Your Professional Account is Ready!",
		Type:    notificationdomain.TypeWelcome,
		Meta: map[string]string{
			"name":         req.Name,
			"account_type": "individual",
		},
	}

	return ch.Send(ctx, channelReq)
}

func (s *notificationService) SendInstitutionWelcome(ctx context.Context, req notificationdomain.SendInstitutionWelcomeRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.WelcomeInstitutionTask{
			To:              req.To,
			AdminName:       req.AdminName,
			InstitutionName: req.InstitutionName,
		}
		if err := s.taskEnqueuer.EnqueueWelcomeInstitution(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue institution welcome task: %v, falling back to sync", err)
			return s.sendInstitutionWelcomeSync(ctx, req)
		}
		return nil
	}
	return s.sendInstitutionWelcomeSync(ctx, req)
}

func (s *notificationService) sendInstitutionWelcomeSync(ctx context.Context, req notificationdomain.SendInstitutionWelcomeRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: "Welcome to Nuruvent - Your Institution is Live!",
		Type:    notificationdomain.TypeWelcome,
		Meta: map[string]string{
			"admin_name":        req.AdminName,
			"institution_name":  req.InstitutionName,
			"account_type":      "institution",
		},
	}

	return ch.Send(ctx, channelReq)
}

func (s *notificationService) SendWelcome(ctx context.Context, req notificationdomain.SendWelcomeRequest) error {
	return s.SendIndividualWelcome(ctx, req)
}

// ============================================================
// TWO FACTOR AUTHENTICATION
// ============================================================

func (s *notificationService) SendTwoFactorOTP(ctx context.Context, req notificationdomain.SendTwoFactorRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.TwoFactorOTPTask{
			To:        req.To,
			Name:      req.Name,
			OTP:       req.OTP,
			Expires:   req.Expires,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		if err := s.taskEnqueuer.EnqueueTwoFactorOTP(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue 2FA OTP task: %v, falling back to sync", err)
			return s.sendTwoFactorOTPSync(ctx, req)
		}
		return nil
	}
	return s.sendTwoFactorOTPSync(ctx, req)
}

func (s *notificationService) sendTwoFactorOTPSync(ctx context.Context, req notificationdomain.SendTwoFactorRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: "Two-Factor Authentication - Nuruvent",
		Type:    notificationdomain.TypeTwoFactor,
		Meta: map[string]string{
			"name":       req.Name,
			"otp":        req.OTP,
			"expires":    req.Expires,
			"ip_address": req.IPAddress,
			"user_agent": req.UserAgent,
		},
	}

	return ch.Send(ctx, channelReq)
}

// ============================================================
// PASSWORD RESET
// ============================================================

func (s *notificationService) SendPasswordResetOTP(ctx context.Context, req notificationdomain.SendOTPRequest) error {
	req.Purpose = notificationdomain.PurposePasswordReset
	return s.SendVerificationOTP(ctx, req)
}

func (s *notificationService) SendPasswordResetConfirm(ctx context.Context, req notificationdomain.SendPasswordResetConfirmRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.PasswordResetConfirmTask{
			To:   req.To,
			Name: req.Name,
		}
		if err := s.taskEnqueuer.EnqueuePasswordResetConfirm(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue password reset confirm task: %v, falling back to sync", err)
			return s.sendPasswordResetConfirmSync(ctx, req)
		}
		return nil
	}
	return s.sendPasswordResetConfirmSync(ctx, req)
}

func (s *notificationService) sendPasswordResetConfirmSync(ctx context.Context, req notificationdomain.SendPasswordResetConfirmRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: "Password Reset Confirmation - Nuruvent",
		Type:    notificationdomain.TypePasswordResetConfirm,
		Meta: map[string]string{
			"name": req.Name,
		},
	}

	return ch.Send(ctx, channelReq)
}

// ============================================================
// LOGIN NOTIFICATION
// ============================================================

func (s *notificationService) SendLoginNotification(ctx context.Context, req notificationdomain.SendLoginNotificationRequest) error {
	// Validate channel exists
	_, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if s.async && s.taskEnqueuer != nil {
		task := notificationdomain.LoginNotificationTask{
			To:        req.To,
			Name:      req.Name,
			Time:      req.Time,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		if err := s.taskEnqueuer.EnqueueLoginNotification(ctx, task); err != nil {
			log.Printf("[NotificationService] Failed to enqueue login notification task: %v, falling back to sync", err)
			return s.sendLoginNotificationSync(ctx, req)
		}
		return nil
	}
	return s.sendLoginNotificationSync(ctx, req)
}

func (s *notificationService) sendLoginNotificationSync(ctx context.Context, req notificationdomain.SendLoginNotificationRequest) error {
	ch, err := s.getChannel(notificationdomain.ChannelEmail)
	if err != nil {
		return err
	}

	if req.Time == "" {
		req.Time = time.Now().Format("January 2, 2006 at 3:04 PM")
	}

	channelReq := notificationdomain.ChannelRequest{
		To:      req.To,
		Subject: "New Login Notification - Nuruvent",
		Type:    notificationdomain.TypeLoginNotification,
		Meta: map[string]string{
			"name":       req.Name,
			"time":       req.Time,
			"ip_address": req.IPAddress,
			"user_agent": req.UserAgent,
		},
	}

	return ch.Send(ctx, channelReq)
}