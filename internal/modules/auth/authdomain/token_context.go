// internal/modules/auth/authdomain/token_context.go

package authdomain

// TokenContext holds all user context for token generation
type TokenContext struct {
	// Core user information
	UserID      string
	Email       string
	DisplayName string
	Role        string

	// Account type - uses slug (snake_case) from account_types table
	// Examples: "personal", "institution"
	AccountTypeSlug string

	// Team information - uses slug (kebab-case) from team_types table
	// Examples: "personal-team", "institution-team"
	TeamTypeSlug string

	// Team ID - either user_id (personal) or institution_id (institution)
	TeamID string

	// Institution ID (only for institution team members) - DEPRECATED: Use TeamID instead
	// Deprecated: Use TeamID instead
	InstitutionID string

	// User status
	IsVerified bool
	IsActive   bool
}

// ============================================================
// CONSTRUCTORS
// ============================================================

// NewPersonalTokenContext creates a new token context for a personal team user
// Domain: personal:team:{userID}
func NewPersonalTokenContext(userID, email, displayName, role, accountTypeSlug string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountTypeSlug: accountTypeSlug,  // snake_case: "personal"
		TeamTypeSlug:    "personal-team",  // kebab-case
		TeamID:          userID,           // For personal teams, TeamID = UserID
		InstitutionID:   "",               // No institution for personal teams
		IsVerified:      true,
		IsActive:        true,
	}
}

// NewInstitutionTokenContext creates a new token context for an institution team user
// Domain: institution:team:{institutionID}
func NewInstitutionTokenContext(userID, email, displayName, role, accountTypeSlug, institutionID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountTypeSlug: accountTypeSlug,    // snake_case: "institution"
		TeamTypeSlug:    "institution-team", // kebab-case
		TeamID:          institutionID,      // For institution teams, TeamID = InstitutionID
		InstitutionID:   institutionID,      // Keep for backward compatibility
		IsVerified:      true,
		IsActive:        true,
	}
}

// ============================================================
// TEAM CONTEXT HELPERS
// ============================================================

// GetPersonalTeamDomain returns the personal team domain for this user
// Format: "personal:team:{user_id}"
func (c *TokenContext) GetPersonalTeamDomain() string {
	return PersonalTeamDomain(c.UserID)
}

// GetInstitutionTeamDomain returns the institution team domain for this user
// Format: "institution:team:{institution_id}"
// Returns empty string if user is not in an institution team
func (c *TokenContext) GetInstitutionTeamDomain() string {
	if c.InstitutionID != "" {
		return InstitutionTeamDomain(c.InstitutionID)
	}
	return ""
}

// GetTeamDomain returns the appropriate team domain for this user
// For personal teams: "personal:team:{user_id}"
// For institution teams: "institution:team:{institution_id}"
func (c *TokenContext) GetTeamDomain() string {
	if c.IsInstitutionTeam() && c.InstitutionID != "" {
		return InstitutionTeamDomain(c.InstitutionID)
	}
	return PersonalTeamDomain(c.UserID)
}

// GetTeamID returns the team ID (either institution ID or user ID)
func (c *TokenContext) GetTeamID() string {
	if c.TeamID != "" {
		return c.TeamID
	}
	if c.InstitutionID != "" {
		return c.InstitutionID
	}
	return c.UserID
}

// IsPersonalTeam returns true if the user has a personal team
func (c *TokenContext) IsPersonalTeam() bool {
	return c.TeamTypeSlug == "personal-team"
}

// IsInstitutionTeam returns true if the user is part of an institution team
func (c *TokenContext) IsInstitutionTeam() bool {
	return c.TeamTypeSlug == "institution-team" || c.InstitutionID != ""
}

// ============================================================
// ROLE HELPERS
// ============================================================

// GetRole returns the user's role as a Role type
func (c *TokenContext) GetRole() Role {
	return Role(c.Role)
}

// HasRole checks if the user has a specific role
func (c *TokenContext) HasRole(role Role) bool {
	return c.Role == role.String()
}

// HasAnyRole checks if the user has any of the specified roles
func (c *TokenContext) HasAnyRole(roles ...Role) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// ============================================================
// PERMISSION HELPERS
// ============================================================

// IsSuperAdmin checks if the user is a super admin
func (c *TokenContext) IsSuperAdmin() bool {
	return c.Role == RoleSuperAdmin.String()
}

// IsAdmin checks if the user is a platform admin
func (c *TokenContext) IsAdmin() bool {
	return c.Role == RoleAdmin.String()
}

// IsAccountAdmin checks if the user is an account admin
func (c *TokenContext) IsAccountAdmin() bool {
	return c.Role == RoleAccountAdmin.String()
}

// IsEventManager checks if the user is an event manager
func (c *TokenContext) IsEventManager() bool {
	return c.Role == RoleEventManager.String()
}

// IsTeamMember checks if the user is a team member
func (c *TokenContext) IsTeamMember() bool {
	return c.Role == RoleTeamMember.String()
}

// IsGuest checks if the user is a guest
func (c *TokenContext) IsGuest() bool {
	return c.Role == RoleGuest.String()
}

// IsPlatformRole checks if the user has a platform-level role
func (c *TokenContext) IsPlatformRole() bool {
	return IsPlatformRole(c.Role)
}

// IsTeamRole checks if the user has a team-level role
func (c *TokenContext) IsTeamRole() bool {
	return IsTeamRole(c.Role)
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

// IsValid checks if the token context has all required fields
func (c *TokenContext) IsValid() bool {
	return c.UserID != "" && c.Email != "" && c.Role != ""
}

// HasTeamAccess checks if the user has any team access
func (c *TokenContext) HasTeamAccess() bool {
	return c.IsPersonalTeam() || c.IsInstitutionTeam()
}