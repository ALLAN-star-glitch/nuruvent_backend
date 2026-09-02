// internal/modules/auth/authdomain/scope.go

package authdomain

import "fmt"

// ScopeType defines the type of authorization boundary
type ScopeType string

const (
	// ScopeTypePlatform represents platform-wide scope
	ScopeTypePlatform ScopeType = "platform"
	
	// ScopeTypePersonalTeam represents a personal team scope
	// ID is the user_id of the team owner
	ScopeTypePersonalTeam ScopeType = "personal_team"
	
	// ScopeTypeInstitutionTeam represents an institution team scope
	// ID is the institution_id
	ScopeTypeInstitutionTeam ScopeType = "institution_team"
)

// Scope defines the authorization boundary
type Scope struct {
	Type ScopeType
	ID   string // user_id, institution_id, or empty for platform
}

// NewPlatformScope creates a platform-wide scope
func NewPlatformScope() Scope {
	return Scope{Type: ScopeTypePlatform, ID: ""}
}

// NewPersonalTeamScope creates a personal team scope
func NewPersonalTeamScope(userID string) Scope {
	return Scope{Type: ScopeTypePersonalTeam, ID: userID}
}

// NewInstitutionTeamScope creates an institution team scope
func NewInstitutionTeamScope(institutionID string) Scope {
	return Scope{Type: ScopeTypeInstitutionTeam, ID: institutionID}
}

// Domain returns the Casbin domain string for this scope
func (s Scope) Domain() string {
	switch s.Type {
	case ScopeTypePlatform:
		return DomainPlatform
	case ScopeTypePersonalTeam:
		return PersonalTeamDomain(s.ID)
	case ScopeTypeInstitutionTeam:
		return InstitutionTeamDomain(s.ID)
	default:
		return ""
	}
}

// IsValid checks if the scope is valid
func (s Scope) IsValid() bool {
	switch s.Type {
	case ScopeTypePlatform:
		return true
	case ScopeTypePersonalTeam, ScopeTypeInstitutionTeam:
		return s.ID != ""
	default:
		return false
	}
}

// String returns a string representation of the scope
func (s Scope) String() string {
	return fmt.Sprintf("%s:%s", s.Type, s.ID)
}

// ============================================================
// TYPE CHECKERS
// ============================================================

// IsPersonalTeam checks if this is a personal team scope
func (s Scope) IsPersonalTeam() bool {
	return s.Type == ScopeTypePersonalTeam
}

// IsInstitutionTeam checks if this is an institution team scope
func (s Scope) IsInstitutionTeam() bool {
	return s.Type == ScopeTypeInstitutionTeam
}

// IsPlatform checks if this is a platform scope
func (s Scope) IsPlatform() bool {
	return s.Type == ScopeTypePlatform
}

// IsTeam checks if this is a team scope (personal or institution)
func (s Scope) IsTeam() bool {
	return s.IsPersonalTeam() || s.IsInstitutionTeam()
}

// ============================================================
// EXTRACTION HELPERS
// ============================================================

// TeamID returns the team ID if this is a team scope, empty string otherwise
func (s Scope) TeamID() string {
	if s.IsTeam() {
		return s.ID
	}
	return ""
}

// UserID returns the user ID if this is a personal team scope, empty string otherwise
func (s Scope) UserID() string {
	if s.IsPersonalTeam() {
		return s.ID
	}
	return ""
}

// InstitutionID returns the institution ID if this is an institution team scope, empty string otherwise
func (s Scope) InstitutionID() string {
	if s.IsInstitutionTeam() {
		return s.ID
	}
	return ""
}

// ============================================================
// CONVERSION HELPERS
// ============================================================

// ToDomain returns the Casbin domain string (same as Domain())
func (s Scope) ToDomain() string {
	return s.Domain()
}

// FromDomain creates a Scope from a domain string
func FromDomain(domain string) Scope {
	if domain == DomainPlatform {
		return NewPlatformScope()
	}
	if IsPersonalTeamDomain(domain) {
		return NewPersonalTeamScope(ExtractTeamID(domain))
	}
	if IsInstitutionTeamDomain(domain) {
		return NewInstitutionTeamScope(ExtractTeamID(domain))
	}
	return Scope{}
}

// ============================================================
// EQUALITY HELPERS
// ============================================================

// Equals checks if two scopes are equal
func (s Scope) Equals(other Scope) bool {
	return s.Type == other.Type && s.ID == other.ID
}

// IsEmpty checks if the scope is empty (invalid)
func (s Scope) IsEmpty() bool {
	return s.Type == "" && s.ID == ""
}

// ============================================================
// MOCK HELPERS (for testing)
// ============================================================

// MockScope creates a scope for testing purposes
func MockScope(scopeType ScopeType, id string) Scope {
	return Scope{Type: scopeType, ID: id}
}