// internal/modules/auth/authdomain/repository.go

package authdomain

import "context"

type Repository interface {
	// ============================================================
	// ACCOUNT OPERATIONS
	// ============================================================

	AccountExistsByEmail(email string) (bool, error)
	AccountExistsByPhone(phone string) (bool, error)
	GetAccountByEmail(email string) (*Account, error)
	GetAccountByPhone(phone string) (*Account, error)
	GetAccountByID(id string) (*Account, error)
	CreateAccount(account *Account) error
	UpdateAccount(account *Account) error
	UpdateAccountInstitutionID(accountID string, institutionID string) error // ✅ NEW
	GetAccountTypeByID(id string) (*AccountType, error)
	GetAccountTypeBySlug(slug string) (*AccountType, error)

	// ============================================================
	// REFRESH TOKEN OPERATIONS
	// ============================================================

	CreateRefreshToken(token *RefreshToken) error
	GetRefreshTokenByToken(token string) (*RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllAccountRefreshTokens(accountID string) error

	// ============================================================
	// INSTITUTION OPERATIONS
	// ============================================================

	CreateInstitution(institution *Institution) error
	GetInstitutionByID(id string) (*Institution, error)
	GetInstitutionByAccountID(accountID string) (*Institution, error)
	GetInstitutionTypeBySlug(slug string) (*InstitutionType, error)
	UpdateInstitution(institution *Institution) error
	InstitutionExists(id string) (bool, error)
	GetInstitutionsByType(institutionTypeID string) ([]*Institution, error)

	// ============================================================
	// TEAM MEMBER OPERATIONS
	// ============================================================

	CreateTeamMember(member *TeamMember) error
	GetTeamMemberByID(id string) (*TeamMember, error)
	GetTeamMemberByAccountAndInstitution(accountID, institutionID string) (*TeamMember, error)
	UpdateTeamMember(member *TeamMember) error
	DeleteTeamMember(id string) error
	GetTeamMembersByInstitution(institutionID string) ([]*TeamMember, error)
	GetTeamMembersByAccount(accountID string) ([]*TeamMember, error)
	IsMemberOfInstitution(ctx context.Context, accountID, institutionID string) (bool, error)
	GetMemberRole(ctx context.Context, accountID, institutionID string) (string, error)

	// ============================================================
	// INSTITUTION ADMIN CHECKS
	// ============================================================

	IsInstitutionAdmin(ctx context.Context, accountID, institutionID string) (bool, error)

	// ============================================================
	// PLATFORM ADMIN CHECKS
	// ============================================================

	IsPlatformAdmin(ctx context.Context, accountID string) (bool, error)
	IsSuperAdmin(ctx context.Context, accountID string) (bool, error)
}