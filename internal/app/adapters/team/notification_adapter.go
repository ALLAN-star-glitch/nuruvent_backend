// internal/app/adapters/team/notification_adapter.go

package team

import (
	"context"

	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// TeamNotificationAdapter adapts notificationdomain.NotificationService to
// teamService.NotificationService.
type TeamNotificationAdapter struct {
	notifSvc notificationdomain.NotificationService
}

// NewTeamNotificationAdapter creates a new team notification adapter.
func NewTeamNotificationAdapter(notifSvc notificationdomain.NotificationService) teamService.NotificationService {
	return &TeamNotificationAdapter{notifSvc: notifSvc}
}

// ============================================================
// TEAM INVITATION METHODS
// ============================================================

// SendTeamInviteExistingUser sends a team invitation to an existing user.
//
// Sent when the recipient already has a Nuruvent account. Includes the
// role they are being granted (only relevant if they aren't yet an
// account member; the email template decides whether to mention it).
func (a *TeamNotificationAdapter) SendTeamInviteExistingUser(
	ctx context.Context,
	req teamService.SendTeamInviteExistingUserRequest,
) error {
	notifReq := notificationdomain.SendTeamInviteExistingUserRequest{
		To:                  req.To,
		UserName:            req.UserName,
		InvitedBy:           req.InvitedBy,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		Role:                req.Role,
		AcceptLink:          req.AcceptLink,
		ExpiresIn:           req.ExpiresIn,
		PersonalizedContent: req.PersonalizedContent,
	}
	return a.notifSvc.SendTeamInviteExistingUser(ctx, notifReq)
}

// SendTeamInviteRegistration sends an invitation to a new user with a
// registration link.
//
// Sent when the recipient does not have a Nuruvent account. Includes the
// role they'll receive after registering.
func (a *TeamNotificationAdapter) SendTeamInviteRegistration(
	ctx context.Context,
	req teamService.SendTeamInviteRegistrationRequest,
) error {
	notifReq := notificationdomain.SendTeamInviteRegistrationRequest{
		To:                  req.To,
		Name:                req.Name,
		InvitedBy:           req.InvitedBy,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		Role:                req.Role,
		RegistrationLink:    req.RegistrationLink,
		ExpiresIn:           req.ExpiresIn,
		PersonalizedContent: req.PersonalizedContent,
	}
	return a.notifSvc.SendTeamInviteRegistration(ctx, notifReq)
}

// SendTeamInviteAccepted sends a notification to the inviter when a user
// accepts an invitation.
func (a *TeamNotificationAdapter) SendTeamInviteAccepted(
	ctx context.Context,
	req teamService.SendTeamInviteAcceptedRequest,
) error {
	notifReq := notificationdomain.SendTeamInviteAcceptedRequest{
		To:                  req.To,
		AdminName:           req.AdminName,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		Role:                req.Role,
		PersonalizedContent: req.PersonalizedContent,
	}
	return a.notifSvc.SendTeamInviteAccepted(ctx, notifReq)
}

// SendTeamInviteDeclined sends a notification to the inviter when a user
// declines an invitation.
func (a *TeamNotificationAdapter) SendTeamInviteDeclined(
	ctx context.Context,
	req teamService.SendTeamInviteDeclinedRequest,
) error {
	notifReq := notificationdomain.SendTeamInviteDeclinedRequest{
		To:                  req.To,
		AdminName:           req.AdminName,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		Role:                req.Role,
		PersonalizedContent: req.PersonalizedContent,
	}
	return a.notifSvc.SendTeamInviteDeclined(ctx, notifReq)
}