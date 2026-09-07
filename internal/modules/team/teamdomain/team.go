// internal/modules/team/teamdomain/team.go

package teamdomain

import (
    "errors"
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
    AccountID   string // ✅ Added: ID of the account this team belongs to
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

    now := time.Now()
    return &Team{
        ID:          uuid.New().String(),
        AccountID:   accountID,
        Name:        name,
        DisplayName: displayName,
        Slug:        slug,
        Type:        teamType,
        IsActive:    true,
        CreatedAt:   now,
        UpdatedAt:   now,
    }, nil
}

// NewPersonalTeam creates a new personal team
// For personal teams, AccountID = UserID
func NewPersonalTeam(userID, userName string) (*Team, error) {
    if userID == "" {
        return nil, errors.New("user ID is required")
    }
    if userName == "" {
        return nil, errors.New("user name is required")
    }

    displayName := userName + "'s Personal Team"
    slug := "personal-" + userID
    name := "personal_" + userID

    return NewTeam(userID, name, displayName, slug, TeamTypePersonal)
}

// NewInstitutionTeam creates a new institution team
// For institution teams, AccountID = InstitutionID
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