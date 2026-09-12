// internal/modules/auth/authdomain/team.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// TEAM TYPE CONSTANTS
// ============================================================
//
// Team types are data, not authorization domains. They determine:
//   - Whether a team lives under a personal or institution account
//   - UI behavior and default settings
//   - Onboarding flows
//
// They do NOT affect Casbin domains. Both personal and institution teams
// resolve to their parent account domain for authorization.

const (
	TeamTypePersonal    = "personal"
	TeamTypeInstitution = "institution"
)

// Team represents a team (personal or institution).
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

// NewTeam creates a new team.
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

// IsPersonal reports whether the team is a personal team.
func (t *Team) IsPersonal() bool {
	return t.Type == TeamTypePersonal
}

// IsInstitution reports whether the team is an institution team.
func (t *Team) IsInstitution() bool {
	return t.Type == TeamTypeInstitution
}

// Deactivate marks the team inactive.
func (t *Team) Deactivate() {
	t.IsActive = false
	t.UpdatedAt = time.Now()
}

// Activate marks the team active.
func (t *Team) Activate() {
	t.IsActive = true
	t.UpdatedAt = time.Now()
}

// ============================================================
// AUTHORIZATION NOTE
// ============================================================
//
// Teams do NOT have their own Casbin domains. The previous GetDomain()
// method returned "personal:team:<id>" or "institution:team:<id>", which
// are no longer valid.
//
// To authorize a team-scoped action:
//
//   1. Resolve the team's parent account:
//        accountDomain := AccountDomain(team.AccountID)
//
//   2. Check the user's role permission in that account domain:
//        allowed, _ := checker.HasPermission(ctx, userID, accountDomain, resource, action)
//
//   3. Enforce team visibility separately (in the service layer):
//        isMember, _ := teamRepo.IsMember(ctx, userID, team.ID)
//
// Gate 1 is role-permission on the account domain.
// Gate 2 is team membership, enforced as data.
//
// See "Authorization Revamp Design Document" §2.3 and §2.4.