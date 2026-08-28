package notificationdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// NOTIFICATION ENTITY
// ============================================================

type Notification struct {
	ID          string
	UserID      string
	Channel     NotificationChannel
	Type        NotificationType
	Subject     string
	Content     string
	Status      NotificationStatus
	SentAt      *time.Time
	DeliveredAt *time.Time
	ReadAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewNotification(userID string, channel NotificationChannel, notifType NotificationType, subject, content string) (*Notification, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if channel == "" {
		return nil, errors.New("channel is required")
	}
	if notifType == "" {
		return nil, errors.New("notification type is required")
	}
	if subject == "" {
		return nil, errors.New("subject is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}

	now := time.Now()
	return &Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Channel:   channel,
		Type:      notifType,
		Subject:   subject,
		Content:   content,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ============================================================
// NOTIFICATION CHANNELS
// ============================================================

type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelWhatsApp NotificationChannel = "whatsapp"
	ChannelInApp    NotificationChannel = "in_app"
)

func (c NotificationChannel) String() string {
	return string(c)
}

func (c NotificationChannel) IsValid() bool {
	switch c {
	case ChannelEmail, ChannelSMS, ChannelWhatsApp, ChannelInApp:
		return true
	default:
		return false
	}
}

// ============================================================
// NOTIFICATION TYPES
// ============================================================

type NotificationType string

const (
	TypeVerificationOTP      NotificationType = "verification_otp"
	TypeWelcome              NotificationType = "welcome"
	TypePasswordReset        NotificationType = "password_reset"
	TypePasswordResetConfirm NotificationType = "password_reset_confirm"
	TypeTwoFactor            NotificationType = "two_factor"
	TypeLoginNotification    NotificationType = "login_notification"
	TypeWelcomeInstitution   NotificationType = "welcome_institution" // For Admin of institution
	TypeWelcomeInstitutionKYC NotificationType = "welcome_institution_kyc"
	TypeNewInstitutionAccountRegistration		NotificationType = "new_account_institution_registration_notice"
	TypeNewPersonalAccountRegistration			NotificationType = "new_account_personal_registration_notice"
)

func (t NotificationType) String() string {
	return string(t)
}

// ============================================================
// NOTIFICATION STATUS
// ============================================================

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusQueued    NotificationStatus = "queued"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRead      NotificationStatus = "read"
)

// ============================================================
// BEHAVIORS
// ============================================================

func (n *Notification) MarkAsQueued() {
	n.Status = StatusQueued
	n.UpdatedAt = time.Now()
}

func (n *Notification) MarkAsSent() {
	now := time.Now()
	n.Status = StatusSent
	n.SentAt = &now
	n.UpdatedAt = now
}

func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.Status = StatusDelivered
	n.DeliveredAt = &now
	n.UpdatedAt = now
}

func (n *Notification) MarkAsFailed() {
	n.Status = StatusFailed
	n.UpdatedAt = time.Now()
}

func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Status = StatusRead
	n.ReadAt = &now
	n.UpdatedAt = now
}

func (n *Notification) IsDelivered() bool {
	return n.Status == StatusDelivered || n.Status == StatusRead
}