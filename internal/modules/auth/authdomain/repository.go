// internal/modules/auth/authdomain/repository.go

package authdomain

import "context"

type Repository interface {

	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error

	
	// ============================================================
	// USER OPERATIONS
	// ============================================================

	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserExistsByPhone(ctx context.Context, phone string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	ReactivateUser(ctx context.Context, userID string) error

	// ============================================================
	// ACCOUNT TYPE OPERATIONS
	// ============================================================

	GetAccountTypeByID(ctx context.Context, id string) (*AccountType, error)
	GetAccountTypeBySlug(ctx context.Context, slug string) (*AccountType, error)
	GetAccountTypeByName(ctx context.Context, name string) (*AccountType, error)

	// ============================================================
	// INSTITUTION TYPE OPERATIONS
	// ============================================================

	GetInstitutionTypeByID(ctx context.Context, id string) (*InstitutionType, error)
	GetInstitutionTypeBySlug(ctx context.Context, slug string) (*InstitutionType, error)
	GetInstitutionTypeByName(ctx context.Context, name string) (*InstitutionType, error)
	ListInstitutionTypes(ctx context.Context) ([]*InstitutionType, error)

	// ============================================================
	// PROFESSIONAL TYPE OPERATIONS
	// ============================================================

	GetProfessionalTypeByID(ctx context.Context, id string) (*ProfessionalType, error)
	GetProfessionalTypeBySlug(ctx context.Context, slug string) (*ProfessionalType, error)
	GetProfessionalTypeByName(ctx context.Context, name string) (*ProfessionalType, error)
	ListProfessionalTypes(ctx context.Context) ([]*ProfessionalType, error)

	// ============================================================
	// REFRESH TOKEN OPERATIONS
	// ============================================================

	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error
	UpdateRefreshTokenContext(ctx context.Context, token, userAgent, ipAddress string) error

	// ============================================================
	// ACCOUNT OPERATIONS
	// ============================================================

	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountBySlug(ctx context.Context, slug string) (*Account, error)
	GetAccountsByUserID(ctx context.Context, userID string) ([]*Account, error)
	CreateAccount(ctx context.Context, account *Account) error
	UpdateAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id string) error
	AccountExists(ctx context.Context, id string) (bool, error)

	// ============================================================
	// ACCOUNT MEMBER OPERATIONS
	// ============================================================

	CreateAccountMember(ctx context.Context, member *AccountMember) error
	GetAccountMemberByID(ctx context.Context, id string) (*AccountMember, error)
	GetAccountMemberByAccountAndUser(ctx context.Context, accountID, userID string) (*AccountMember, error)
	GetAccountMembersByAccount(ctx context.Context, accountID string) ([]*AccountMember, error)
	GetAccountMembersByUser(ctx context.Context, userID string) ([]*AccountMember, error)
	UpdateAccountMember(ctx context.Context, member *AccountMember) error
	DeleteAccountMember(ctx context.Context, accountID, userID string) error
	IsAccountMember(ctx context.Context, accountID, userID string) (bool, error)
	IsAccountAdmin(ctx context.Context, accountID, userID string) (bool, error)
	IsAccountTrainer(ctx context.Context, accountID, userID string) (bool, error)
	CountAccountMembers(ctx context.Context, accountID string) (int64, error)

	// ============================================================
	// PLATFORM ADMIN CHECKS
	// ============================================================

	IsPlatformAdmin(ctx context.Context, userID string) (bool, error)
	IsSuperAdmin(ctx context.Context, userID string) (bool, error)
}