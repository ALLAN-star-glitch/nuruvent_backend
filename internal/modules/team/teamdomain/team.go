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
func NewTeam(name, displayName, slug string, teamType TeamType) (*Team, error) {
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
func NewPersonalTeam(userID, userName string) (*Team, error) {
    // Personal team naming convention: {user_name}'s Personal Team
    displayName := userName + "'s Personal Team"
    slug := "personal-" + userID

    return NewTeam("personal_"+userID, displayName, slug, TeamTypePersonal)
}

// NewInstitutionTeam creates a new institution team
func NewInstitutionTeam(name, displayName, slug string) (*Team, error) {
    return NewTeam(name, displayName, slug, TeamTypeInstitution)
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