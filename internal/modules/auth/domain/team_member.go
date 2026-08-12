package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TeamMember struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	AccountID   string
	MemberID    string
	Role        string
	JobTitle    string
	IsActive    bool
	CreatedBy   *string // ✅ Add this field (nullable)
	JoinedAt    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTeamMember(accountID, memberID, role string) (*TeamMember, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:          uuid.New().String(),
		Slug:        "team-" + uuid.New().String()[:8],
		Name:        "",
		DisplayName: "",
		AccountID:   accountID,
		MemberID:    memberID,
		Role:        role,
		JobTitle:    "",
		IsActive:    true,
		CreatedBy:   nil, // ✅ Self-created, no creator
		JoinedAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewTeamMemberWithCreator creates a team member with a creator
func NewTeamMemberWithCreator(accountID, memberID, role, createdBy string) (*TeamMember, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}
	if createdBy == "" {
		return nil, errors.New("created by is required")
	}

	now := time.Now()
	return &TeamMember{
		ID:          uuid.New().String(),
		Slug:        "team-" + uuid.New().String()[:8],
		Name:        "",
		DisplayName: "",
		AccountID:   accountID,
		MemberID:    memberID,
		Role:        role,
		JobTitle:    "",
		IsActive:    true,
		CreatedBy:   &createdBy, // ✅ Set creator
		JoinedAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (tm *TeamMember) Deactivate() {
	tm.IsActive = false
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) Activate() {
	tm.IsActive = true
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) UpdateRole(role string) error {
	if role == "" {
		return errors.New("role is required")
	}
	tm.Role = role
	tm.UpdatedAt = time.Now()
	return nil
}

func (tm *TeamMember) UpdateJobTitle(jobTitle string) {
	tm.JobTitle = jobTitle
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) UpdateName(name string) {
	tm.Name = name
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) UpdateDisplayName(displayName string) {
	tm.DisplayName = displayName
	tm.UpdatedAt = time.Now()
}

func (tm *TeamMember) IsActiveMember() bool {
	return tm.IsActive
}