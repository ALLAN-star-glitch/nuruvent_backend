// internal/modules/auth/authdomain/token_context.go

package authdomain

// TokenContext holds all user context for token generation
type TokenContext struct {
	// UserID is the unique identifier of the user
	UserID string

	// Email is the user's email address
	Email string

	// Role is the user's role (super_admin, admin, trainer, manager, member, attendee)
	Role string

	// AccountTypeID is the ID of the account type from the AccountType value object
	AccountTypeID string

	// AccountTypeSlug is the slug of the account type (individual, institution, etc.)
	AccountTypeSlug string

	// AccountID is the account identifier (always set for all users)
	AccountID string

	// InstitutionID is the institution identifier (only for institution accounts)
	InstitutionID string
}

// NewIndividualTokenContext creates a new token context for an individual user
func NewIndividualTokenContext(userID, email, role, accountTypeID, accountTypeSlug, accountID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		Role:            role,
		AccountTypeID:   accountTypeID,
		AccountTypeSlug: accountTypeSlug,
		AccountID:       accountID,
	}
}

// NewInstitutionTokenContext creates a new token context for an institution user
func NewInstitutionTokenContext(userID, email, role, accountTypeID, accountTypeSlug, accountID, institutionID string) *TokenContext {
	return &TokenContext{
		UserID:          userID,
		Email:           email,
		Role:            role,
		AccountTypeID:   accountTypeID,
		AccountTypeSlug: accountTypeSlug,
		AccountID:       accountID,
		InstitutionID:   institutionID,
	}
}

// IsIndividual checks if the account is an individual account
func (c *TokenContext) IsIndividual() bool {
	return c.AccountTypeSlug == "individual"
}

// IsInstitution checks if the account is an institution account
func (c *TokenContext) IsInstitution() bool {
	return c.AccountTypeSlug == "institution"
}

// HasInstitution checks if the user belongs to an institution
func (c *TokenContext) HasInstitution() bool {
	return c.InstitutionID != ""
}