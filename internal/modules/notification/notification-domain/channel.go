package notificationdomain

import "context"

// ============================================================
// OUTBOUND PORT: Notification Channel Interface
// ============================================================

// Channel is the interface that all notification channels must implement
type Channel interface {
	Send(ctx context.Context, req ChannelRequest) error
	GetChannel() NotificationChannel
	GetPriority() int
}

// ChannelRequest represents a request to send through a channel
type ChannelRequest struct {
	To      string
	From    string
	Subject string
	Content string
	HTML    string
	Text    string
	Type    NotificationType
	Meta    map[string]string
}

// ============================================================
// EMAIL SENDER INTERFACE
// ============================================================

// EmailSender defines the email sending interface
type EmailSender interface {
	SendEmail(ctx context.Context, req SendEmailRequest) error
}

// SendEmailRequest represents an email send request
type SendEmailRequest struct {
	To      string
	From    string
	Subject string
	HTML    string
	Text    string
}