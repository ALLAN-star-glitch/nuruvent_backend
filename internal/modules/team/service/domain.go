// internal/modules/team/service/domain.go

package service

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"

// Domain represents a permission domain
type Domain interface {
    IsPersonal() bool
    IsInstitution() bool
    GetID() string
    String() string // Returns the domain string for Casbin
}

// ============================================================
// DOMAIN IMPLEMENTATIONS
// ============================================================

// PersonalTeamDomain represents a personal team domain
// Format: personal:team:{team_id}
type PersonalTeamDomain struct {
    TeamID string
}

func NewPersonalTeamDomain(teamID string) PersonalTeamDomain {
    return PersonalTeamDomain{TeamID: teamID}
}

func (d PersonalTeamDomain) IsPersonal() bool {
    return true
}

func (d PersonalTeamDomain) IsInstitution() bool {
    return false
}

func (d PersonalTeamDomain) GetID() string {
    return d.TeamID
}

func (d PersonalTeamDomain) String() string {
    return "personal:team:" + d.TeamID
}

// InstitutionTeamDomain represents an institution team domain
// Format: institution:team:{team_id}
type InstitutionTeamDomain struct {
    TeamID string
}

func NewInstitutionTeamDomain(teamID string) InstitutionTeamDomain {
    return InstitutionTeamDomain{TeamID: teamID}
}

func (d InstitutionTeamDomain) IsPersonal() bool {
    return false
}

func (d InstitutionTeamDomain) IsInstitution() bool {
    return true
}

func (d InstitutionTeamDomain) GetID() string {
    return d.TeamID
}

func (d InstitutionTeamDomain) String() string {
    return "institution:team:" + d.TeamID
}

// AccountDomain represents an account domain
// Format: account:{account_id}
type AccountDomain struct {
    AccountID string
}

func NewAccountDomain(accountID string) AccountDomain {
    return AccountDomain{AccountID: accountID}
}

func (d AccountDomain) IsPersonal() bool {
    return false
}

func (d AccountDomain) IsInstitution() bool {
    return false
}

func (d AccountDomain) GetID() string {
    return d.AccountID
}

func (d AccountDomain) String() string {
    return "account:" + d.AccountID
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// NewTeamDomain creates a domain from a team
func NewTeamDomain(team *teamdomain.Team) Domain {
    if team.IsPersonal() {
        return NewPersonalTeamDomain(team.ID)
    }
    return NewInstitutionTeamDomain(team.ID)
}

// NewDomainFromTeam creates a domain from a team
// Alias for NewTeamDomain
func NewDomainFromTeam(team *teamdomain.Team) Domain {
    return NewTeamDomain(team)
}

// NewDomainFromTeamID creates a domain from team ID and type
func NewDomainFromTeamID(teamID, teamType string) Domain {
    if teamType == "personal" {
        return NewPersonalTeamDomain(teamID)
    }
    return NewInstitutionTeamDomain(teamID)
}