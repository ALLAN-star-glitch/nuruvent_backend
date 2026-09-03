// internal/modules/profile/domain/scope.go

package domain

// ScopeType defines the type of authorization boundary
type ScopeType string

const (
    // ScopeTypePersonalTeam represents a personal team scope
    ScopeTypePersonalTeam ScopeType = "personal_team"
    
    // ScopeTypeInstitutionTeam represents an institution team scope
    ScopeTypeInstitutionTeam ScopeType = "institution_team"
)

// Scope defines the authorization boundary for profiles
type Scope struct {
    Type ScopeType
    ID   string // user_id or institution_id
}

// NewPersonalTeamScope creates a personal team scope
func NewPersonalTeamScope(userID string) Scope {
    return Scope{Type: ScopeTypePersonalTeam, ID: userID}
}

// NewInstitutionTeamScope creates an institution team scope
func NewInstitutionTeamScope(institutionID string) Scope {
    return Scope{Type: ScopeTypeInstitutionTeam, ID: institutionID}
}

// IsPersonal checks if this is a personal team scope
func (s Scope) IsPersonal() bool {
    return s.Type == ScopeTypePersonalTeam
}

// IsInstitution checks if this is an institution team scope
func (s Scope) IsInstitution() bool {
    return s.Type == ScopeTypeInstitutionTeam
}

// String returns a string representation
func (s Scope) String() string {
    return string(s.Type) + ":" + s.ID
}