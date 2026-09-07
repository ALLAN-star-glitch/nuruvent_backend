// internal/app/adapters/auth/notification_adapter.go

package auth

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationAdapter adapts the notification module's service to the auth domain's NotificationService
type NotificationAdapter struct {
	notifSvc notificationDomain.NotificationService
}

// NewNotificationAdapter creates a new notification adapter for auth
func NewNotificationAdapter(notifSvc notificationDomain.NotificationService) authDomain.NotificationService {
	return &NotificationAdapter{
		notifSvc: notifSvc,
	}
}

// ============================================================
// UNIFIED OTP
// ============================================================

// SendOTP sends a verification OTP for any purpose
func (a *NotificationAdapter) SendOTP(ctx context.Context, req authDomain.SendOTPRequest) error {
	// Convert string purpose to notification domain VerificationPurpose
	purpose := notificationDomain.VerificationPurpose(req.Purpose)

	// Convert auth domain request to notification domain request
	notifReq := notificationDomain.SendOTPRequest{
		To:      req.To,
		Name:    req.Name,
		OTP:     req.OTP,
		Expires: req.Expires,
		Purpose: purpose,
		Meta:    req.Meta,
	}
	return a.notifSvc.SendOTP(ctx, notifReq)
}

// ============================================================
// WELCOME EMAILS
// ============================================================

// SendIndividualWelcome sends a welcome email to an individual user
func (a *NotificationAdapter) SendIndividualWelcome(ctx context.Context, req authDomain.SendWelcomeRequest) error {
	notifReq := notificationDomain.SendWelcomeRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendIndividualWelcome(ctx, notifReq)
}

// SendInstitutionWelcome sends a welcome email to an institution admin
func (a *NotificationAdapter) SendInstitutionWelcome(ctx context.Context, req authDomain.SendInstitutionWelcomeRequest) error {
	notifReq := notificationDomain.SendInstitutionWelcomeRequest{
		To:               req.To,
		AdminName:        req.AdminName,
		InstitutionName:  req.InstitutionName,
		InstitutionEmail: req.InstitutionEmail,
	}
	return a.notifSvc.SendInstitutionWelcome(ctx, notifReq)
}

// SendInstitutionKYCWelcome sends KYC welcome email for institutions
func (a *NotificationAdapter) SendInstitutionKYCWelcome(ctx context.Context, req authDomain.SendInstitutionKYCWelcomeRequest) error {
	notifReq := notificationDomain.SendInstitutionKYCWelcomeRequest{
		To:              req.To,
		AdminName:       req.AdminName,
		InstitutionName: req.InstitutionName,
		InstitutionType: req.InstitutionType,
	}
	return a.notifSvc.SendInstitutionKYCWelcome(ctx, notifReq)
}

// ============================================================
// PASSWORD RESET CONFIRM
// ============================================================

// SendPasswordResetConfirm sends a password reset confirmation email
func (a *NotificationAdapter) SendPasswordResetConfirm(ctx context.Context, req authDomain.SendPasswordResetConfirmRequest) error {
	notifReq := notificationDomain.SendPasswordResetConfirmRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendPasswordResetConfirm(ctx, notifReq)
}

// ============================================================
// SECURITY NOTIFICATIONS
// ============================================================

// SendLoginNotification sends a login notification email
func (a *NotificationAdapter) SendLoginNotification(ctx context.Context, req authDomain.SendLoginNotificationRequest) error {
	notifReq := notificationDomain.SendLoginNotificationRequest{
		To:        req.To,
		Name:      req.Name,
		Time:      req.Time,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	return a.notifSvc.SendLoginNotification(ctx, notifReq)
}

// SendNewInstitutionAccountNotification sends notification for new institution account registration
func (a *NotificationAdapter) SendNewInstitutionAccountNotification(ctx context.Context, req authDomain.SendNewInstitutionAccountRegistrationRequest) error {
	notifReq := notificationDomain.SendNewInstitutionAccountRegistrationRequest{
		To:                  req.To,
		NewAccountAdminName: req.NewAccountAdminName,
		InstitutionName:     req.InstitutionName,
		InstitutionType:     req.InstitutionType,
	}
	return a.notifSvc.SendNewInstitutionAccountNotification(ctx, notifReq)
}

// SendNewPersonalAccountNotification sends notification for new personal account registration
func (a *NotificationAdapter) SendNewPersonalAccountNotification(ctx context.Context, req authDomain.SendNewPersonalAccountRegistrationRequest) error {
	notifReq := notificationDomain.SendNewPersonalAccountRegistrationRequest{
		To:                  req.To,
		NewAccountAdminName: req.NewAccountAdminName,
	}
	return a.notifSvc.SendNewPersonalAccountNotification(ctx, notifReq)
}