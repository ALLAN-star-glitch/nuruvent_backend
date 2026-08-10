package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// ============================================================
	// ACCOUNT OPERATIONS
	// ============================================================

	GetAccountByID(ctx context.Context, id string) (*domain.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*domain.Account, error)
	GetAccountByPhone(ctx context.Context, phone string) (*domain.Account, error)
	UpdateAccount(ctx context.Context, account *domain.Account) error
	DeleteAccount(ctx context.Context, id string) error

	// ============================================================
	// PROFILE OPERATIONS
	// ============================================================

	UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*domain.Account, error)
	UpdateProfessionalType(ctx context.Context, id string, professionalTypeID *string) (*domain.Account, error)

	// ============================================================
	// ACCOUNT TYPE OPERATIONS
	// ============================================================

	GetAccountTypeBySlug(ctx context.Context, slug string) (*domain.AccountType, error)
	ListAccountTypes(ctx context.Context) ([]*domain.AccountType, error)

	// ============================================================
	// PROFESSIONAL TYPE OPERATIONS
	// ============================================================

	GetProfessionalTypeBySlug(ctx context.Context, slug string) (*domain.ProfessionalType, error)
	ListProfessionalTypes(ctx context.Context) ([]*domain.ProfessionalType, error)

	// ============================================================
	// INSTITUTION OPERATIONS
	// ============================================================

	GetInstitutionByID(ctx context.Context, id string) (*domain.Institution, error)
	GetInstitutionBySlug(ctx context.Context, slug string) (*domain.Institution, error)
	UpdateInstitution(ctx context.Context, institution *domain.Institution) error
	DeleteInstitution(ctx context.Context, id string) error

	// ============================================================
	// INSTITUTION TYPE OPERATIONS
	// ============================================================

	GetInstitutionTypeBySlug(ctx context.Context, slug string) (*domain.InstitutionType, error)
	ListInstitutionTypes(ctx context.Context) ([]*domain.InstitutionType, error)

	// ============================================================
	// TEAM MEMBER OPERATIONS
	// ============================================================

	AddTeamMember(ctx context.Context, accountID, memberID, role string) (*domain.TeamMember, error)
	GetTeamMembersByAccount(ctx context.Context, accountID string) ([]*domain.TeamMember, error)
	GetTeamMemberByAccountAndMember(ctx context.Context, accountID, memberID string) (*domain.TeamMember, error)
	UpdateTeamMemberRole(ctx context.Context, id, role string) (*domain.TeamMember, error)
	RemoveTeamMember(ctx context.Context, id string) error
}

// ============================================================
// COMMANDS & REQUESTS
// ============================================================

type UpdateProfileRequest struct {
	Name        string `json:"name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type AddTeamMemberRequest struct {
	MemberID string `json:"member_id" binding:"required"`
	Role     string `json:"role" binding:"required"`
	JobTitle string `json:"job_title,omitempty"`
}