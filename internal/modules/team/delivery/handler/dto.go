// internal/modules/team/delivery/handler/dto.go

package handler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// TEAM
// ============================================================

type TeamResponse struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
}

func NewTeamResponse(t *teamdomain.Team) TeamResponse {
	if t == nil {
		return TeamResponse{}
	}
	return TeamResponse{
		ID:          t.ID,
		AccountID:   t.AccountID,
		Name:        t.Name,
		DisplayName: t.DisplayName,
		Slug:        t.Slug,
		Type:        string(t.Type),
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
		DeletedAt:   formatNullableTime(t.DeletedAt),
	}
}

type TeamsListResponse struct {
	Teams []TeamResponse `json:"teams"`
	Count int            `json:"count"`
}

func NewTeamsListResponse(teams []*teamdomain.Team) TeamsListResponse {
	out := make([]TeamResponse, len(teams))
	for i, t := range teams {
		out[i] = NewTeamResponse(t)
	}
	return TeamsListResponse{Teams: out, Count: len(out)}
}

// ============================================================
// MEMBER
// ============================================================

type MemberResponse struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
	IsActive  bool   `json:"is_active"`
	JoinedAt  string `json:"joined_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewMemberResponse(m *teamdomain.Member) MemberResponse {
	if m == nil {
		return MemberResponse{}
	}
	return MemberResponse{
		ID:        m.ID,
		TeamID:    m.TeamID,
		UserID:    m.UserID,
		IsActive:  m.IsActive,
		JoinedAt:  m.JoinedAt.Format(time.RFC3339),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func NewMembersListResponse(members []*teamdomain.Member) []MemberResponse {
	out := make([]MemberResponse, len(members))
	for i, m := range members {
		out[i] = NewMemberResponse(m)
	}
	return out
}

// ============================================================
// INVITATION
// ============================================================

type InvitationResponse struct {
	ID         string  `json:"id"`
	TeamID     string  `json:"team_id"`
	Email      string  `json:"email"`
	Role       string  `json:"role"`
	Token      string  `json:"token,omitempty"`
	Status     string  `json:"status"`
	InvitedBy  string  `json:"invited_by"`
	ExpiresAt  string  `json:"expires_at"`
	AcceptedAt *string `json:"accepted_at,omitempty"`
	DeclinedAt *string `json:"declined_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func NewInvitationResponse(i *teamdomain.Invitation) InvitationResponse {
	if i == nil {
		return InvitationResponse{}
	}
	return InvitationResponse{
		ID:         i.ID,
		TeamID:     i.TeamID,
		Email:      i.Email,
		Role:       i.Role,
		Token:      i.Token,
		Status:     string(i.Status),
		InvitedBy:  i.InvitedBy,
		ExpiresAt:  i.ExpiresAt.Format(time.RFC3339),
		AcceptedAt: formatNullableTime(i.AcceptedAt),
		DeclinedAt: formatNullableTime(i.DeclinedAt),
		CreatedAt:  i.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  i.UpdatedAt.Format(time.RFC3339),
	}
}

func NewInvitationsListResponse(invitations []*teamdomain.Invitation) []InvitationResponse {
	out := make([]InvitationResponse, len(invitations))
	for i, inv := range invitations {
		out[i] = NewInvitationResponse(inv)
	}
	return out
}

// ============================================================
// HELPERS
// ============================================================

func formatNullableTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}