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
		ID:        uuid.New().String(),
		Slug:      generateSlug("member"),
		Name:      "Team Member",
		AccountID: accountID,
		MemberID:  memberID,
		Role:      role,
		IsActive:  true,
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}