// internal/modules/auth/authdomain/repository.go

package authdomain

import "context"

type Repository interface {
	// ============================================================
	// USER OPERATIONS
	// ============================================================

	UserExistsByEmail(email string) (bool, error)
	UserExistsByPhone(phone string) (bool, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	GetUserByID(id string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(userID string) error
	ReactivateUser(userID string) error

	// ============================================================
	// ACCOUNT TYPE OPERATIONS
	// ============================================================

	GetAccountTypeByID(id string) (*AccountType, error)
	GetAccountTypeBySlug(slug string) (*AccountType, error)
	GetAccountTypeByName(name string) (*AccountType, error)

	// ============================================================
	// REFRESH TOKEN OPERATIONS
	// ============================================================

	CreateRefreshToken(token *RefreshToken) error
	GetRefreshTokenByToken(token string) (*RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllRefreshTokensForUser(userID string) error
	UpdateRefreshTokenContext(token, userAgent, ipAddress string) error

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


	//============================================================
	// PROFESSIONAL TYPE OPERATIONS
	// ============================================================

	GetProfessionalTypeByID(id string) (*ProfessionalType, error)
	GetProfessionalTypeBySlug(slug string) (*ProfessionalType, error)
	GetProfessionalTypeByName(name string) (*ProfessionalType, error)
	ListProfessionalTypes(ctx context.Context) ([]*ProfessionalType, error)
}