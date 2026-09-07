// internal/app/adapters/team/notification_adapter.go

package team

import (
	"context"

	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// TeamNotificationAdapter adapts notificationdomain.NotificationService to teamService.NotificationService
type TeamNotificationAdapter struct {
	notifSvc notificationdomain.NotificationService
}

// NewTeamNotificationAdapter creates a new team notification adapter
func NewTeamNotificationAdapter(notifSvc notificationdomain.NotificationService) teamService.NotificationService {
	return &TeamNotificationAdapter{notifSvc: notifSvc}
}

// SendTeamInvite sends a team invite email
func (a *TeamNotificationAdapter) SendTeamInvite(ctx context.Context, req teamService.SendTeamInviteRequest) error {
	notifReq := notificationdomain.SendTeamInviteRequest{
		To:         req.To,
		UserName:   req.UserName,
		InvitedBy:  req.InvitedBy,
		TeamName:   req.TeamName,
		TeamID:     req.TeamID,
		Role:       req.Role,
		InviteLink: req.InviteLink,
		ExpiresIn:  req.ExpiresIn,
	}
	return a.notifSvc.SendTeamInvite(ctx, notifReq)
}

// SendTeamInviteRegistration sends a team invite registration email
func (a *TeamNotificationAdapter) SendTeamInviteRegistration(ctx context.Context, req teamService.SendTeamInviteRegistrationRequest) error {
	notifReq := notificationdomain.SendTeamInviteRegistrationRequest{
		To:               req.To,
		Name:             req.Name,
		InvitedBy:        req.InvitedBy,
		TeamName:         req.TeamName,
		Role:             req.Role,
		RegistrationLink: req.RegistrationLink,
	}
	return a.notifSvc.SendTeamInviteRegistration(ctx, notifReq)
}

// SendTeamInviteAccepted sends a team invite accepted notification
func (a *TeamNotificationAdapter) SendTeamInviteAccepted(ctx context.Context, req teamService.SendTeamInviteAcceptedRequest) error {
	notifReq := notificationdomain.SendTeamInviteAcceptedRequest{
		To:        req.To,
		AdminName: req.AdminName,
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
		TeamName:  req.TeamName,
	}
	return a.notifSvc.SendTeamInviteAccepted(ctx, notifReq)
}

// SendTeamInviteDeclined sends a team invite declined notification
func (a *TeamNotificationAdapter) SendTeamInviteDeclined(ctx context.Context, req teamService.SendTeamInviteDeclinedRequest) error {
	notifReq := notificationdomain.SendTeamInviteDeclinedRequest{
		To:        req.To,
		AdminName: req.AdminName,
		UserName:  req.UserName,
		UserEmail: req.UserEmail,
		TeamName:  req.TeamName,
	}
	return a.notifSvc.SendTeamInviteDeclined(ctx, notifReq)
}