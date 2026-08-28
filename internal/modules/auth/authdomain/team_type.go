// internal/modules/auth/authdomain/team_type.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TeamType entity (lookup table for team types)
type TeamType struct {
	ID          string
	Name        string     // snake_case: institution_team, personal_team
	DisplayName string     // Human-readable: Institution Team, Personal Team
	Slug        string     // kebab-case: institution-team, personal-team
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewTeamType creates a validated team type
func NewTeamType(name, displayName, slug, description string) (*TeamType, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	now := time.Now()
	return &TeamType{
		ID:          uuid.New().String(),
		Name:        name,
		DisplayName: displayName,
		Slug:        slug,
		Description: description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewInstitutionTeamType creates a pre-configured institution team type
func NewInstitutionTeamType() (*TeamType, error) {
	return NewTeamType(
		"institution_team",
		"Institution Team",
		"institution-team",
		"Organization or company team",
	)
}

// NewPersonalTeamType creates a pre-configured personal team type
func NewPersonalTeamType() (*TeamType, error) {
	return NewTeamType(
		"personal_team",
		"Personal Team",
		"personal-team",
		"Individual user team",
	)
}

// Behaviors
func (tt *TeamType) Deactivate() {
	tt.IsActive = false
	tt.UpdatedAt = time.Now()
}

func (tt *TeamType) Activate() {
	tt.IsActive = true
	tt.UpdatedAt = time.Now()
}

func (tt *TeamType) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	tt.Name = name
	tt.UpdatedAt = time.Now()
	return nil
}

func (tt *TeamType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("display name cannot be empty")
	}
	tt.DisplayName = displayName
	tt.UpdatedAt = time.Now()
	return nil
}

func (tt *TeamType) UpdateSlug(slug string) error {
	if slug == "" {
		return errors.New("slug cannot be empty")
	}
	tt.Slug = slug
	tt.UpdatedAt = time.Now()
	return nil
}

func (tt *TeamType) UpdateDescription(description string) {
	tt.Description = description
	tt.UpdatedAt = time.Now()
}

func (tt *TeamType) IsActiveType() bool {
	return tt.IsActive
}

func (tt *TeamType) IsInstitutionType() bool {
	return tt.Slug == "institution-team"
}

func (tt *TeamType) IsPersonalType() bool {
	return tt.Slug == "personal-team"
}

// Soft delete
func (tt *TeamType) Delete() {
	now := time.Now()
	tt.DeletedAt = &now
	tt.IsActive = false
	tt.UpdatedAt = now
}