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

// GenerateInvitationRequest contains all data needed for AI to generate content
type GenerateInvitationRequest struct {
	// User information
	RecipientName  string // Name of the person being invited (empty if new user)
	RecipientEmail string // Email of the person being invited
	IsExistingUser bool   // true if user already has an account

	// Inviter information
	InviterName string // Name of the person sending the invitation
	InviterRole string // Role of the inviter (account_admin, trainer)

	// Team information
	TeamName        string   // Name of the team
	TeamType        string   // "personal" or "institution"
	TeamMemberCount int      // Number of members in the team
	TeamEventCount  int      // Number of events hosted by the team

	// Additional context (optional enrichment)
	RecentEventNames []string // Names of recent events (for personalization)
	TeamMemberNames  []string // Names of team members (for personalization)
}