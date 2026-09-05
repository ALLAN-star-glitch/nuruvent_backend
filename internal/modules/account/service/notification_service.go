// internal/modules/account/service/notification_service.go

package service

import "context"

// NotificationService defines the notification operations needed by the account module
type NotificationService interface {
    // SendAccountInvite sends an account invitation email
    SendAccountInvite(ctx context.Context, req SendAccountInviteRequest) error

    // SendAccountWelcome sends a welcome email for new account members
    SendAccountWelcome(ctx context.Context, req SendAccountWelcomeRequest) error
}

type SendAccountInviteRequest struct {
    To         string
    UserName   string
    InvitedBy  string
    AccountName string
    AccountID   string
    Role        string
    InviteLink  string
}

type SendAccountWelcomeRequest struct {
    To          string
    UserName    string
    AccountName string
    Role        string
}