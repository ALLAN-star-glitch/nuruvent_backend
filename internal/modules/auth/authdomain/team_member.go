// internal/modules/auth/authdomain/team_member.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TeamMember entity (links users to institutions or personal teams)
type TeamMember struct {
	ID            string
	MemberID      string   // The user who is a member
	InstitutionID *string  // The institution they belong to (NULL for personal teams)
	TeamTypeID    string   // institution_team or personal_team
	InvitedBy     *string  // Who invited them (NULL if self-created)
	CreatedBy     *string  // Who created this record
	IsActive      bool
	JoinedAt      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// NewTeamMember creates a validated team member for an institution (self-created)
func NewTeamMember(memberID, institutionID, teamTypeID string) (*TeamMember, error) {
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if institutionID == "" {
		return nil, errors.New("institution ID is required for institution team")
	}
	if teamTypeID == "" {
		return nil, errors.New("team type is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InstitutionID: &institutionID,
		TeamTypeID:    teamTypeID,
		InvitedBy:     nil,               // Self-created, no inviter
		CreatedBy:     &memberID,          // Self-created, creator is the member
		IsActive:      true,
		JoinedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     nil,
	}, nil
}

// NewTeamMemberWithInviter creates a team member with a specific inviter (institution team)
func NewTeamMemberWithInviter(memberID, institutionID, teamTypeID, invitedBy string) (*TeamMember, error) {
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if institutionID == "" {
		return nil, errors.New("institution ID is required")
	}
	if teamTypeID == "" {
		return nil, errors.New("team type is required")
	}
	if invitedBy == "" {
		return nil, errors.New("invited by is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InstitutionID: &institutionID,
		TeamTypeID:    teamTypeID,
		InvitedBy:     &invitedBy,
		CreatedBy:     &invitedBy, // The inviter created the record
		IsActive:      true,
		JoinedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     nil,
	}, nil
}

// NewPersonalTeamMember creates a team member for a personal team (self-created)
func NewPersonalTeamMember(memberID, teamTypeID string) (*TeamMember, error) {
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if teamTypeID == "" {
		return nil, errors.New("team type is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InstitutionID: nil,           // Personal team has no institution
		TeamTypeID:    teamTypeID,
		InvitedBy:     nil,           // Self-created, no inviter
		CreatedBy:     &memberID,     // Self-created, creator is the member
		IsActive:      true,
		JoinedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     nil,
	}, nil
}

// NewPersonalTeamMemberWithInviter creates a personal team member with a specific inviter
func NewPersonalTeamMemberWithInviter(memberID, teamTypeID, invitedBy string) (*TeamMember, error) {
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if teamTypeID == "" {
		return nil, errors.New("team type is required")
	}
	if invitedBy == "" {
		return nil, errors.New("invited by is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		InstitutionID: nil,
		TeamTypeID:    teamTypeID,
		InvitedBy:     &invitedBy,
		CreatedBy:     &invitedBy, // The inviter created the record
		IsActive:      true,
		JoinedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     nil,
	}, nil
}

// Behaviors
func (tm *TeamMember) Deactivate() {
	tm.IsActive = false
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) Activate() {
	tm.IsActive = true
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) IsInstitutionTeam() bool {
	return tm.InstitutionID != nil
}

func (tm *TeamMember) IsPersonalTeam() bool {
	return tm.InstitutionID == nil
}

func (tm *TeamMember) IsActiveMember() bool {
	return tm.IsActive
}

// Soft delete
func (tm *TeamMember) Delete() {
	now := time.Now()
	tm.DeletedAt = &now
	tm.IsActive = false
	tm.UpdatedAt = now
}