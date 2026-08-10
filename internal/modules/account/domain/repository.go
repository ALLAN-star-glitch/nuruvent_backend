package domain

import "context"

// ============================================================
// REPOSITORY INTERFACE (Outbound Port)
// ============================================================

type Repository interface {
	// ============================================================
	// ACCOUNT OPERATIONS
	// ============================================================

	// CreateAccount creates a new account
	CreateAccount(ctx context.Context, account *Account) error

	// GetAccountByID retrieves an account by ID
	GetAccountByID(ctx context.Context, id string) (*Account, error)

	// GetAccountByEmail retrieves an account by email
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)

	// GetAccountByPhone retrieves an account by phone
	GetAccountByPhone(ctx context.Context, phone string) (*Account, error)

	// UpdateAccount updates an existing account
	UpdateAccount(ctx context.Context, account *Account) error

	// DeleteAccount soft deletes an account
	DeleteAccount(ctx context.Context, id string) error

	// AccountExistsByEmail checks if an account exists by email
	AccountExistsByEmail(ctx context.Context, email string) (bool, error)

	// AccountExistsByPhone checks if an account exists by phone
	AccountExistsByPhone(ctx context.Context, phone string) (bool, error)

	// ============================================================
	// ACCOUNT TYPE OPERATIONS
	// ============================================================

	// GetAccountTypeByID retrieves an account type by ID
	GetAccountTypeByID(ctx context.Context, id string) (*AccountType, error)

	// GetAccountTypeBySlug retrieves an account type by slug
	GetAccountTypeBySlug(ctx context.Context, slug string) (*AccountType, error)

	// CreateAccountType creates a new account type
	CreateAccountType(ctx context.Context, accountType *AccountType) error

	// UpdateAccountType updates an existing account type
	UpdateAccountType(ctx context.Context, accountType *AccountType) error

	// ListAccountTypes retrieves all account types
	ListAccountTypes(ctx context.Context) ([]*AccountType, error)

	// ============================================================
	// PROFESSIONAL TYPE OPERATIONS
	// ============================================================

	// GetProfessionalTypeByID retrieves a professional type by ID
	GetProfessionalTypeByID(ctx context.Context, id string) (*ProfessionalType, error)

	// GetProfessionalTypeBySlug retrieves a professional type by slug
	GetProfessionalTypeBySlug(ctx context.Context, slug string) (*ProfessionalType, error)

	// ListProfessionalTypes retrieves all professional types
	ListProfessionalTypes(ctx context.Context) ([]*ProfessionalType, error)

	// ============================================================
	// INSTITUTION OPERATIONS
	// ============================================================

	// CreateInstitution creates a new institution
	CreateInstitution(ctx context.Context, institution *Institution) error

	// GetInstitutionByID retrieves an institution by ID
	GetInstitutionByID(ctx context.Context, id string) (*Institution, error)

	// GetInstitutionBySlug retrieves an institution by slug
	GetInstitutionBySlug(ctx context.Context, slug string) (*Institution, error)

	// UpdateInstitution updates an existing institution
	UpdateInstitution(ctx context.Context, institution *Institution) error

	// DeleteInstitution soft deletes an institution
	DeleteInstitution(ctx context.Context, id string) error

	// ============================================================
	// INSTITUTION TYPE OPERATIONS
	// ============================================================

	// GetInstitutionTypeByID retrieves an institution type by ID
	GetInstitutionTypeByID(ctx context.Context, id string) (*InstitutionType, error)

	// GetInstitutionTypeBySlug retrieves an institution type by slug
	GetInstitutionTypeBySlug(ctx context.Context, slug string) (*InstitutionType, error)

	// ListInstitutionTypes retrieves all institution types
	ListInstitutionTypes(ctx context.Context) ([]*InstitutionType, error)

	// ============================================================
	// TEAM MEMBER OPERATIONS
	// ============================================================

	// CreateTeamMember creates a new team member
	CreateTeamMember(ctx context.Context, member *TeamMember) error

	// GetTeamMemberByID retrieves a team member by ID
	GetTeamMemberByID(ctx context.Context, id string) (*TeamMember, error)

	// GetTeamMembersByAccount retrieves all team members for an account
	GetTeamMembersByAccount(ctx context.Context, accountID string) ([]*TeamMember, error)

	// GetTeamMemberByAccountAndMember retrieves a team member by account and member
	GetTeamMemberByAccountAndMember(ctx context.Context, accountID, memberID string) (*TeamMember, error)

	// UpdateTeamMember updates an existing team member
	UpdateTeamMember(ctx context.Context, member *TeamMember) error

	// DeleteTeamMember soft deletes a team member
	DeleteTeamMember(ctx context.Context, id string) error
}