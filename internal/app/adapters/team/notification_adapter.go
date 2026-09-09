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

// ============================================================
// TYPE CONVERSION HELPERS
// ============================================================

// convertPersonalizedContent converts from teamService type to notificationdomain type
func convertPersonalizedContent(src *teamService.PersonalizedInvitationContent) *notificationdomain.PersonalizedInvitationContent {
	if src == nil {
		return nil
	}
	return &notificationdomain.PersonalizedInvitationContent{
		Subject:      src.Subject,
		Greeting:     src.Greeting,
		Intro:        src.Intro,
		Body:         src.Body,
		Benefits:     src.Benefits,
		CallToAction: src.CallToAction,
		Closing:      src.Closing,
		PSS:          src.PSS,
	}
}

// ============================================================
// ✅ TEAM INVITATION METHODS (UPDATED - NO ROLE, AI-READY)
// ============================================================

// SendTeamInviteExistingUser sends a team invitation to an existing user
// Sent when: User already has a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks accept link to join
// ✅ AI-READY - PersonalizedContent field for future AI integration
func (a *TeamNotificationAdapter) SendTeamInviteExistingUser(ctx context.Context, req teamService.SendTeamInviteExistingUserRequest) error {
	notifReq := notificationdomain.SendTeamInviteExistingUserRequest{
		To:                  req.To,
		UserName:            req.UserName,
		InvitedBy:           req.InvitedBy,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		AcceptLink:          req.AcceptLink,
		ExpiresIn:           req.ExpiresIn,
		PersonalizedContent: convertPersonalizedContent(req.PersonalizedContent),
	}
	return a.notifSvc.SendTeamInviteExistingUser(ctx, notifReq)
}

// SendTeamInviteRegistration sends an invitation to a new user with registration link
// Sent when: User does NOT have a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks registration link with token embedded
// ✅ AI-READY - PersonalizedContent field for future AI integration
func (a *TeamNotificationAdapter) SendTeamInviteRegistration(ctx context.Context, req teamService.SendTeamInviteRegistrationRequest) error {
	notifReq := notificationdomain.SendTeamInviteRegistrationRequest{
		To:                  req.To,
		Name:                req.Name,
		InvitedBy:           req.InvitedBy,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		RegistrationLink:    req.RegistrationLink,
		ExpiresIn:           req.ExpiresIn,
		PersonalizedContent: convertPersonalizedContent(req.PersonalizedContent),
	}
	return a.notifSvc.SendTeamInviteRegistration(ctx, notifReq)
}

// SendTeamInviteAccepted sends notification to admin when user accepts invitation
// ✅ AI-READY - PersonalizedContent field for future AI integration
func (a *TeamNotificationAdapter) SendTeamInviteAccepted(ctx context.Context, req teamService.SendTeamInviteAcceptedRequest) error {
	notifReq := notificationdomain.SendTeamInviteAcceptedRequest{
		To:                  req.To,
		AdminName:           req.AdminName,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		PersonalizedContent: convertPersonalizedContent(req.PersonalizedContent),
	}
	return a.notifSvc.SendTeamInviteAccepted(ctx, notifReq)
}

// SendTeamInviteDeclined sends notification to admin when user declines invitation
// ✅ AI-READY - PersonalizedContent field for future AI integration
func (a *TeamNotificationAdapter) SendTeamInviteDeclined(ctx context.Context, req teamService.SendTeamInviteDeclinedRequest) error {
	notifReq := notificationdomain.SendTeamInviteDeclinedRequest{
		To:                  req.To,
		AdminName:           req.AdminName,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		TeamName:            req.TeamName,
		TeamID:              req.TeamID,
		PersonalizedContent: convertPersonalizedContent(req.PersonalizedContent),
	}
	return a.notifSvc.SendTeamInviteDeclined(ctx, notifReq)
}