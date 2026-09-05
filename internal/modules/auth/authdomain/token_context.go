// internal/modules/auth/authdomain/token_context.go

package authdomain

// TokenContext holds all user context for token generation
type TokenContext struct {
	// Core user information
	UserID      string
	Email       string
	DisplayName string
	Role        string

	// Account ID - the account the user belongs to
	// For personal accounts: user_id
	// For institution accounts: institution_id
	AccountID string

	// Account type slug (kebab-case)
	// Examples: "account-type-personal", "account-type-institution"
	AccountTypeSlug string

	// Team ID - UUID from the teams table
	TeamID string

	// Team type slug (kebab-case)
	// Examples: "personal-team", "institution-team"
	TeamTypeSlug string

	// User status
	IsVerified bool
	IsActive   bool
}

// ============================================================
// CONSTRUCTORS
// ============================================================

// NewPersonalTokenContext creates a new token context for a personal team user
func NewPersonalTokenContext(userID, email, displayName, role, accountTypeSlug, teamID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountID:       userID,
		AccountTypeSlug: accountTypeSlug,
		TeamID:          teamID,           // UUID from teams table
		TeamTypeSlug:    "personal-team",
		IsVerified:      true,
		IsActive:        true,
	}
}

// NewInstitutionTokenContext creates a new token context for an institution team user
func NewInstitutionTokenContext(userID, email, displayName, role, accountTypeSlug, accountID, teamID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountID:       accountID,
		AccountTypeSlug: accountTypeSlug,
		TeamID:          teamID,              // UUID from teams table
		TeamTypeSlug:    "institution-team",
		IsVerified:      true,
		IsActive:        true,
	}
}

// ============================================================
// DOMAIN HELPERS
// ============================================================

// GetTeamDomain returns the team domain for this user
// Format: "personal:team:{team_id}" or "institution:team:{team_id}"
func (c *TokenContext) GetTeamDomain() string {
	if c.IsInstitutionTeam() {
		return InstitutionTeamDomain(c.TeamID)
	}
	return PersonalTeamDomain(c.TeamID)
}

// GetAccountDomain returns the account domain for this user
// Format: "account:{account_id}"
func (c *TokenContext) GetAccountDomain() string {
	return AccountDomain(c.AccountID)
}

// ============================================================
// TYPE CHECKERS
// ============================================================

// IsPersonalTeam returns true if the user has a personal team
func (c *TokenContext) IsPersonalTeam() bool {
	return c.TeamTypeSlug == "personal-team"
}

// IsInstitutionTeam returns true if the user is part of an institution team
func (c *TokenContext) IsInstitutionTeam() bool {
	return c.TeamTypeSlug == "institution-team"
}

// IsPersonalAccount returns true if the user has a personal account
func (c *TokenContext) IsPersonalAccount() bool {
	return c.AccountTypeSlug == "account-type-personal"
}

// IsInstitutionAccount returns true if the user has an institution account
func (c *TokenContext) IsInstitutionAccount() bool {
	return c.AccountTypeSlug == "account-type-institution"
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

// IsTrainer checks if the user is a trainer
func (c *TokenContext) IsTrainer() bool {
	return c.Role == RoleTrainer.String()
}

// IsGuest checks if the user is a guest
func (c *TokenContext) IsGuest() bool {
	return c.Role == RoleGuest.String()
}

// IsPlatformRole checks if the user has a platform-level role
func (c *TokenContext) IsPlatformRole() bool {
	return IsPlatformRole(c.Role)
}

// IsAccountRole checks if the user has an account-level role
func (c *TokenContext) IsAccountRole() bool {
	return IsAccountRole(c.Role)
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

// IsValid checks if the token context has all required fields
func (c *TokenContext) IsValid() bool {
	return c.UserID != "" && c.Email != "" && c.Role != "" && c.TeamID != ""
}

// HasTeamAccess checks if the user has any team access
func (c *TokenContext) HasTeamAccess() bool {
	return c.TeamID != ""
}

// HasAccountAccess checks if the user has account access
func (c *TokenContext) HasAccountAccess() bool {
	return c.AccountID != ""
}