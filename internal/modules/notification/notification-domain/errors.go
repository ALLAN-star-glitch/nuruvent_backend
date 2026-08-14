package notificationdomain

import "errors"

// ============================================================
// DOMAIN ERRORS
// ============================================================

var (
	// Notification errors
	ErrNotificationNotFound       = errors.New("notification not found")
	ErrInvalidChannel             = errors.New("invalid notification channel")
	ErrInvalidType                = errors.New("invalid notification type")
	ErrInvalidStatus              = errors.New("invalid notification status")
	ErrNotificationFailed         = errors.New("notification failed to send")
	ErrTemplateNotFound           = errors.New("notification template not found")
	ErrInvalidVerificationPurpose = errors.New("invalid verification purpose")

	// Channel specific errors
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPhone    = errors.New("invalid phone number")
	ErrEmailSendFailed = errors.New("failed to send email")
	ErrSMSSendFailed   = errors.New("failed to send SMS")
)