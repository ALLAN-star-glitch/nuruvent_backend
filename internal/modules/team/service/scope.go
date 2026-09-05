// internal/modules/team/service/scope.go

package service

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"

type Scope interface {
    IsPersonal() bool
    IsInstitution() bool
    GetID() string
    String() string
}


// ============================================================
// SCOPE IMPLEMENTATIONS
// ============================================================

// PersonalTeamScope represents a personal team scope
type PersonalTeamScope struct {
    UserID string
}

func NewPersonalTeamScope(userID string) PersonalTeamScope {
    return PersonalTeamScope{UserID: userID}
}

func (s PersonalTeamScope) IsPersonal() bool {
    return true
}

func (s PersonalTeamScope) IsInstitution() bool {
    return false
}

func (s PersonalTeamScope) GetID() string {
    return s.UserID
}

func (s PersonalTeamScope) String() string {
    return "personal:team:" + s.UserID
}

// InstitutionTeamScope represents an institution team scope
type InstitutionTeamScope struct {
    InstitutionID string
}

func NewInstitutionTeamScope(institutionID string) InstitutionTeamScope {
    return InstitutionTeamScope{InstitutionID: institutionID}
}

func (s InstitutionTeamScope) IsPersonal() bool {
    return false
}

func (s InstitutionTeamScope) IsInstitution() bool {
    return true
}

func (s InstitutionTeamScope) GetID() string {
    return s.InstitutionID
}

func (s InstitutionTeamScope) String() string {
    return "institution:team:" + s.InstitutionID
}

// NewTeamScope creates a scope from a team
func NewTeamScope(team *teamdomain.Team) interface {
    IsPersonal() bool
    IsInstitution() bool
    GetID() string
    String() string
} {
    if team.IsPersonal() {
        return NewPersonalTeamScope(team.ID)
    }
    return NewInstitutionTeamScope(team.ID)
}