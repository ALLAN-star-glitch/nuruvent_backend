// internal/modules/team/teamdomain/team.go

package teamdomain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TeamType represents the type of team
type TeamType string

const (
	TeamTypePersonal    TeamType = "personal"
	TeamTypeInstitution TeamType = "institution"
)

// Team represents a team (personal or institution)
type Team struct {
	ID          string
	AccountID   string
	Name        string
	DisplayName string
	Slug        string
	Type        TeamType
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewTeam creates a new team
func NewTeam(accountID, name, displayName, slug string, teamType TeamType) (*Team, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if name == "" {
		return nil, errors.New("team name is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if teamType == "" {
		return nil, errors.New("team type is required")
	}

	// Every team's display name carries the "Team" prefix so they're
	// visually distinct from accounts in the switcher and headers.
	prefixedDisplayName := "Team " + displayName

	now := time.Now()
	return &Team{
		ID:          uuid.New().String(),
		AccountID:   accountID,
		Name:        name,
		DisplayName: prefixedDisplayName,
		Slug:        slug,
		Type:        teamType,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// personalSlugSuffix returns a short, URL-safe suffix used to keep
// personal team names and slugs unique across multiple teams owned by
// the same user.
func personalSlugSuffix() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
}

// NewPersonalTeam creates a new personal team.
//
// Both the name and slug carry a per-team suffix so a user may create
// any number of personal teams without violating the (account_id, name)
// or (slug) unique constraints.
func NewPersonalTeam(userID, userName string) (*Team, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if userName == "" {
		return nil, errors.New("user name is required")
	}

	displayName := userName + "'s Personal Team"
	slug := "personal-" + userID + "-" + personalSlugSuffix()
	name := "personal_" + userID + "_" + personalSlugSuffix()

	return NewTeam(userID, name, displayName, slug, TeamTypePersonal)
}

// NewPersonalTeamWithAccount creates a new personal team with a
// specific account ID.
//
// Both the name and slug carry a per-team suffix so a user may create
// any number of personal teams under a given account without
// violating the (account_id, name) or (slug) unique constraints.
//
// displayName is the user-supplied label. When empty, falls back to
// "<userName>'s Personal Team".
func NewPersonalTeamWithAccount(
	userID, userName, accountID, displayName string,
) (*Team, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if userName == "" {
		return nil, errors.New("user name is required")
	}
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	if displayName == "" {
		displayName = userName + "'s Personal Team"
	}

	slug := "personal-" + userID + "-" + personalSlugSuffix()
	name := "personal_" + userID + "_" + personalSlugSuffix()

	return NewTeam(accountID, name, displayName, slug, TeamTypePersonal)
}

// NewInstitutionTeam creates a new institution team.
func NewInstitutionTeam(accountID, name, displayName, slug string) (*Team, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	return NewTeam(accountID, name, displayName, slug, TeamTypeInstitution)
}

// IsPersonal returns true if team is a personal team
func (t *Team) IsPersonal() bool {
	return t.Type == TeamTypePersonal
}

// IsInstitution returns true if team is an institution team
func (t *Team) IsInstitution() bool {
	return t.Type == TeamTypeInstitution
}

// Deactivate deactivates the team
func (t *Team) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now()
}

// Activate activates the team
func (t *Team) Activate() {
	t.IsActive = true
	t.UpdatedAt = time.Now()
}

// GetDomain returns the Casbin domain for this team
func (t *Team) GetDomain() string {
	if t.IsPersonal() {
		return "personal:team:" + t.ID
	}
	return "institution:team:" + t.ID
}