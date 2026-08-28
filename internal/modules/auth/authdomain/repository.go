// internal/modules/auth/authdomain/repository.go

package authdomain

import "context"

type Repository interface {
	// ============================================================
	// USER OPERATIONS (formerly Account)
	// ============================================================

	UserExistsByEmail(email string) (bool, error)
	UserExistsByPhone(phone string) (bool, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	GetUserByID(id string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	UpdateUserInstitutionID(userID string, institutionID *string) error
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
	// INSTITUTION OPERATIONS
	// ============================================================

	CreateInstitution(institution *Institution) error
	GetInstitutionByID(id string) (*Institution, error)
	GetInstitutionByUserID(userID string) (*Institution, error)
	GetInstitutionTypeBySlug(slug string) (*InstitutionType, error)
	GetInstitutionTypeByName(name string) (*InstitutionType, error)
	UpdateInstitution(institution *Institution) error
	InstitutionExists(id string) (bool, error)
	GetInstitutionsByType(institutionTypeID string) ([]*Institution, error)

	// ============================================================
	// PROFESSIONAL TYPE OPERATIONS
	// ============================================================

	GetProfessionalTypeByID(id string) (*ProfessionalType, error)
	GetProfessionalTypeBySlug(slug string) (*ProfessionalType, error)
	GetProfessionalTypeByName(name string) (*ProfessionalType, error)
	ListProfessionalTypes(ctx context.Context) ([]*ProfessionalType, error)

	// ============================================================
	// TEAM MEMBER OPERATIONS
	// ============================================================

	CreateTeamMember(member *TeamMember) error
	GetTeamMemberByID(id string) (*TeamMember, error)
	GetTeamMemberByMemberAndInstitution(memberID, institutionID string) (*TeamMember, error)
	UpdateTeamMember(member *TeamMember) error
	DeleteTeamMember(id string) error
	GetTeamMembersByInstitution(institutionID string) ([]*TeamMember, error)
	GetTeamMembersByMember(memberID string) ([]*TeamMember, error)
	GetTeamMembersByTeamType(teamTypeID string) ([]*TeamMember, error)
	IsMemberOfInstitution(ctx context.Context, memberID, institutionID string) (bool, error)

	// ============================================================
	// TEAM TYPE OPERATIONS
	// ============================================================

	GetTeamTypeByID(id string) (*TeamType, error)
	GetTeamTypeBySlug(slug string) (*TeamType, error)
	GetTeamTypeByName(name string) (*TeamType, error)
	CreateTeamType(teamType *TeamType) error
	UpdateTeamType(teamType *TeamType) error
	DeleteTeamType(id string) error
	ListTeamTypes(ctx context.Context) ([]*TeamType, error)

	// ============================================================
	// INSTITUTION ADMIN CHECKS
	// ============================================================

	IsInstitutionAdmin(ctx context.Context, memberID, institutionID string) (bool, error)

	// ============================================================
	// PLATFORM ADMIN CHECKS
	// ============================================================

	IsPlatformAdmin(ctx context.Context, userID string) (bool, error)
	IsSuperAdmin(ctx context.Context, userID string) (bool, error)
}