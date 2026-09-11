// internal/modules/auth/authdomain/team.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Team represents a team (personal or institution)
type Team struct {
	ID          string
	AccountID   string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	Description string
	LogoURL     string
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewTeam creates a new team
func NewTeam(accountID, name, displayName, slug, teamType, createdBy string) (*Team, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if name == "" {
		return nil, errors.New("team name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if teamType == "" {
		return nil, errors.New("team type is required")
	}
	if teamType != TeamTypePersonal && teamType != TeamTypeInstitution {
		return nil, errors.New("invalid team type")
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
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// IsPersonal checks if team is personal
func (t *Team) IsPersonal() bool {
	return t.Type == TeamTypePersonal
}

// IsInstitution checks if team is institution
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
		return PersonalTeamDomain(t.ID)
	}
	return InstitutionTeamDomain(t.ID)
}