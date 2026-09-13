// internal/modules/team/service/notification_service.go

package service

import (
	"context"

	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// NOTIFICATION SERVICE INTERFACE
// ============================================================

type NotificationService interface {
	SendTeamInviteExistingUser(ctx context.Context, req SendTeamInviteExistingUserRequest) error
	SendTeamInviteRegistration(ctx context.Context, req SendTeamInviteRegistrationRequest) error
	SendTeamInviteAccepted(ctx context.Context, req SendTeamInviteAcceptedRequest) error
	SendTeamInviteDeclined(ctx context.Context, req SendTeamInviteDeclinedRequest) error
}

// ============================================================
// REQUEST STRUCTS
// ============================================================

// SendTeamInviteExistingUserRequest carries the data needed to email an
// invitation to a user who already has an account on the platform.
type SendTeamInviteExistingUserRequest struct {
	To                  string
	UserName            string
	InvitedBy           string
	TeamName            string
	TeamID              string
	Role                string // "account_admin" or "trainer"
	AcceptLink          string
	ExpiresIn           string
	PersonalizedContent *types.PersonalizedInvitationContent
}

// SendTeamInviteRegistrationRequest carries the data needed to email an
// invitation to someone who has not registered yet.
type SendTeamInviteRegistrationRequest struct {
	To                  string
	Name                string
	InvitedBy           string
	TeamName            string
	TeamID              string
	Role                string // "account_admin" or "trainer"
	RegistrationLink    string
	ExpiresIn           string
	PersonalizedContent *types.PersonalizedInvitationContent
}

// SendTeamInviteAcceptedRequest notifies the inviter that the invitee has
// accepted the invitation.
type SendTeamInviteAcceptedRequest struct {
	To                  string
	AdminName           string
	UserName            string
	UserEmail           string
	TeamName            string
	TeamID              string
	Role                string // the role the invitee was granted
	PersonalizedContent *types.PersonalizedInvitationContent
}

// SendTeamInviteDeclinedRequest notifies the inviter that the invitee has
// declined the invitation.
type SendTeamInviteDeclinedRequest struct {
	To                  string
	AdminName           string
	UserName            string
	UserEmail           string
	TeamName            string
	TeamID              string
	Role                string
	PersonalizedContent *types.PersonalizedInvitationContent
}