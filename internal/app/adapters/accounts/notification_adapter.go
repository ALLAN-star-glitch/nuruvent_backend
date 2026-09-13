// internal/app/adapters/accounts/notification_adapter.go

package accounts

import (
	"context"

	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationAdapter adapts the notification module service to the account module's NotificationService interface
type NotificationAdapter struct {
	notifSvc notificationDomain.NotificationService
}

// NewNotificationAdapter creates a new notification adapter for the account module
func NewNotificationAdapter(notifSvc notificationDomain.NotificationService) accountService.NotificationService {
	return &NotificationAdapter{
		notifSvc: notifSvc,
	}
}

// SendAccountInvite maps account invite requests to the underlying team invite notification call
// ✅ Updated: Uses SendTeamInviteExistingUser (no role)
func (a *NotificationAdapter) SendAccountInvite(ctx context.Context, req accountService.SendAccountInviteRequest) error {
	return a.notifSvc.SendTeamInviteExistingUser(ctx, notificationDomain.SendTeamInviteExistingUserRequest{
		To:         req.To,
		UserName:   req.UserName,
		InvitedBy:  req.InvitedBy,
		TeamName:   req.AccountName,
		TeamID:     req.AccountID,
		AcceptLink: req.InviteLink,
		ExpiresIn:  "7 days",
	})
}


// SendAccountWelcome maps account welcome requests to the individual welcome notification call
func (a *NotificationAdapter) SendAccountWelcome(ctx context.Context, req accountService.SendAccountWelcomeRequest) error {
	return a.notifSvc.SendIndividualWelcome(ctx, notificationDomain.SendWelcomeRequest{
		To:   req.To,
		Name: req.UserName,
	})
}