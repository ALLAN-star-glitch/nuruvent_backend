// internal/modules/auth/authdomain/token_context.go

package authdomain

// TokenContext holds all user context for token generation.
//
// ============================================================
// DOMAIN CONTEXT (post-revamp)
// ============================================================
//
// The only authorization domain carried by a token is the account domain
// ("account:<account_id>"). Team context is informational — it tells the
// client which team the user is currently viewing — but it does NOT
// participate in authorization checks.
//
// Team membership is enforced as data (team_members) in the service layer.
// Casbin policies only ever see the account domain.
type TokenContext struct {
	// Core user information
	UserID      string
	Email       string
	DisplayName string
	Role        string

	// Account ID - the account the user belongs to.
	// For personal accounts: user_id
	// For institution accounts: account_id
	AccountID string

	// Account type slug (kebab-case)
	// Examples: "personal", "institution"
	AccountTypeSlug string

	// Team ID - UUID from the teams table (optional, informational only)
	// This is the team the user is currently viewing. Not used for authz.
	TeamID string

	// Team type slug (kebab-case)
	// Examples: "personal", "institution"
	TeamTypeSlug string

	// User status
	IsVerified bool
	IsActive   bool
}

// ============================================================
// CONSTRUCTORS
// ============================================================

// NewTokenContext creates a new token context with all fields.
func NewTokenContext(userID, email, displayName, role, accountID, accountTypeSlug, teamID, teamTypeSlug string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountID:       accountID,
		AccountTypeSlug: accountTypeSlug,
		TeamID:          teamID,
		TeamTypeSlug:    teamTypeSlug,
		IsVerified:      true,
		IsActive:        true,
	}
}

// NewPersonalTokenContext creates a token context for a personal account user.
func NewPersonalTokenContext(userID, email, displayName, role, teamID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountID:       userID,
		AccountTypeSlug: "personal",
		TeamID:          teamID,
		TeamTypeSlug:    "personal",
		IsVerified:      true,
		IsActive:        true,
	}
}

// NewInstitutionTokenContext creates a token context for an institution account user.
func NewInstitutionTokenContext(userID, email, displayName, role, accountID, teamID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		DisplayName:     displayName,
		Role:            role,
		AccountID:       accountID,
		AccountTypeSlug: "institution",
		TeamID:          teamID,
		TeamTypeSlug:    "institution",
		IsVerified:      true,
		IsActive:        true,
	}
}

// ============================================================
// DOMAIN HELPERS
// ============================================================

// GetAccountDomain returns the account domain for this user.
// Format: "account:{account_id}"
//
// This is the ONLY authorization domain in the post-revamp model.
// Team domains no longer exist.
func (c *TokenContext) GetAccountDomain() string {
	return AccountDomain(c.AccountID)
}

// ============================================================
// TYPE CHECKERS
// ============================================================

// IsPersonal returns true if the user has a personal account.
func (c *TokenContext) IsPersonal() bool {
	return c.AccountTypeSlug == "personal"
}

// IsInstitution returns true if the user has an institution account.
func (c *TokenContext) IsInstitution() bool {
	return c.AccountTypeSlug == "institution"
}

// HasTeamContext returns true if the user has an active team context.
// This is informational only — team context does not affect authorization.
func (c *TokenContext) HasTeamContext() bool {
	return c.TeamID != ""
}

// HasAccountContext returns true if the user has an account context.
func (c *TokenContext) HasAccountContext() bool {
	return c.AccountID != ""
}

// ============================================================
// ROLE HELPERS
// ============================================================

// GetRole returns the user's role as a Role type.
func (c *TokenContext) GetRole() Role {
	return Role(c.Role)
}

// HasRole checks if the user has a specific role.
func (c *TokenContext) HasRole(role Role) bool {
	return c.Role == role.String()
}

// HasAnyRole checks if the user has any of the specified roles.
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

// IsSuperAdmin checks if the user is a super admin.
func (c *TokenContext) IsSuperAdmin() bool {
	return c.Role == RoleSuperAdmin.String()
}

// IsAdmin checks if the user is a platform admin.
func (c *TokenContext) IsAdmin() bool {
	return c.Role == RoleAdmin.String()
}

// IsAccountAdmin checks if the user is an account admin.
func (c *TokenContext) IsAccountAdmin() bool {
	return c.Role == RoleAccountAdmin.String()
}

// IsTrainer checks if the user is a trainer.
func (c *TokenContext) IsTrainer() bool {
	return c.Role == RoleTrainer.String()
}

// IsPlatformRole checks if the user has a platform-level role.
func (c *TokenContext) IsPlatformRole() bool {
	return IsPlatformRole(c.Role)
}

// IsAccountRole checks if the user has an account-level role.
func (c *TokenContext) IsAccountRole() bool {
	return IsAccountRole(c.Role)
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

// IsValid checks if the token context has all required fields.
func (c *TokenContext) IsValid() bool {
	return c.UserID != "" && c.Email != "" && c.Role != "" && c.AccountID != ""
}

// HasAccess checks if the user has any access.
func (c *TokenContext) HasAccess() bool {
	return c.HasAccountContext()
}