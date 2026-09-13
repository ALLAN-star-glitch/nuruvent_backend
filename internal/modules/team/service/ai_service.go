// internal/modules/team/service/ai_service.go

package service

import (
	"context"

	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// OUTBOUND PORT: AIService
// ============================================================

// AIService defines the AI operations needed by the team module.
// This is an OUTBOUND PORT - the core defines it, infrastructure implements it.
type AIService interface {
	// GenerateInvitationContent generates personalized email content for invitations
	GenerateInvitationContent(ctx context.Context, req GenerateInvitationRequest) (*types.PersonalizedInvitationContent, error)
}

// ============================================================
// REQUEST STRUCT
// ============================================================

// GenerateInvitationRequest contains all data needed for AI to generate content.

type GenerateInvitationRequest struct {
	RecipientName  string
	RecipientEmail string
	IsExistingUser bool

	// IsExistingAccountMember is true when the recipient already belongs
	// to the account the team is under. In that case, the invitation
	// only adds them to the team; their role is unchanged and should
	// not be mentioned in the email.
	IsExistingAccountMember bool

	InviterName string
	InviterRole string

	TeamName        string
	TeamType        string
	TeamMemberCount int
	TeamEventCount  int

	InvitedRole string

	RecentEventNames []string
	TeamMemberNames  []string
}