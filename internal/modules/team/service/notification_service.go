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
// REQUEST STRUCTS (Using shared types)
// ============================================================

type SendTeamInviteExistingUserRequest struct {
	To                  string
	UserName            string
	InvitedBy           string
	TeamName            string
	TeamID              string
	AcceptLink          string
	ExpiresIn           string
	PersonalizedContent *types.PersonalizedInvitationContent // ✅ Uses shared type
}

type SendTeamInviteRegistrationRequest struct {
	To                  string
	Name                string
	InvitedBy           string
	TeamName            string
	TeamID              string
	RegistrationLink    string
	ExpiresIn           string
	PersonalizedContent *types.PersonalizedInvitationContent // ✅ Uses shared type
}

type SendTeamInviteAcceptedRequest struct {
	To                  string
	AdminName           string
	UserName            string
	UserEmail           string
	TeamName            string
	TeamID              string
	PersonalizedContent *types.PersonalizedInvitationContent // ✅ Uses shared type
}

type SendTeamInviteDeclinedRequest struct {
	To                  string
	AdminName           string
	UserName            string
	UserEmail           string
	TeamName            string
	TeamID              string
	PersonalizedContent *types.PersonalizedInvitationContent // ✅ Uses shared type
}