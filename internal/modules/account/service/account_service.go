package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
)

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type accountService struct {
	repo    domain.Repository
	permSvc domain.PermissionService
}

func NewService(repo domain.Repository, permSvc domain.PermissionService) Service {
	return &accountService{
		repo:    repo,
		permSvc: permSvc,
	}
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (s *accountService) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	if id == "" {
		return nil, errors.New("account ID is required")
	}
	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}
	return account, nil
}

func (s *accountService) GetAccountByEmail(ctx context.Context, email string) (*domain.Account, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	account, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}
	return account, nil
}

func (s *accountService) GetAccountByPhone(ctx context.Context, phone string) (*domain.Account, error) {
	if phone == "" {
		return nil, errors.New("phone is required")
	}
	account, err := s.repo.GetAccountByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}
	return account, nil
}

func (s *accountService) UpdateAccount(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return errors.New("account is required")
	}
	if account.ID == "" {
		return errors.New("account ID is required")
	}
	return s.repo.UpdateAccount(ctx, account)
}

func (s *accountService) DeleteAccount(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("account ID is required")
	}
	return s.repo.DeleteAccount(ctx, id)
}

// ============================================================
// PROFILE OPERATIONS
// ============================================================

func (s *accountService) UpdateProfile(ctx context.Context, id string, req UpdateProfileRequest) (*domain.Account, error) {
	account, err := s.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		if err := account.UpdateName(req.Name); err != nil {
			return nil, err
		}
	}
	if req.Phone != "" {
		if err := account.UpdatePhone(req.Phone); err != nil {
			return nil, err
		}
	}
	if req.DisplayName != "" {
		account.DisplayName = req.DisplayName
		account.UpdatedAt = time.Now()
	}

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return account, nil
}

func (s *accountService) UpdateProfessionalType(ctx context.Context, id string, professionalTypeID *string) (*domain.Account, error) {
	account, err := s.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	account.ProfessionalTypeID = professionalTypeID
	account.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update professional type: %w", err)
	}

	return account, nil
}

// ============================================================
// ACCOUNT TYPE OPERATIONS
// ============================================================

func (s *accountService) GetAccountTypeBySlug(ctx context.Context, slug string) (*domain.AccountType, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.GetAccountTypeBySlug(ctx, slug)
}

func (s *accountService) ListAccountTypes(ctx context.Context) ([]*domain.AccountType, error) {
	return s.repo.ListAccountTypes(ctx)
}

// ============================================================
// PROFESSIONAL TYPE OPERATIONS
// ============================================================

func (s *accountService) GetProfessionalTypeBySlug(ctx context.Context, slug string) (*domain.ProfessionalType, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.GetProfessionalTypeBySlug(ctx, slug)
}

func (s *accountService) ListProfessionalTypes(ctx context.Context) ([]*domain.ProfessionalType, error) {
	return s.repo.ListProfessionalTypes(ctx)
}

// ============================================================
// INSTITUTION OPERATIONS
// ============================================================

func (s *accountService) GetInstitutionByID(ctx context.Context, id string) (*domain.Institution, error) {
	if id == "" {
		return nil, errors.New("institution ID is required")
	}
	institution, err := s.repo.GetInstitutionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if institution == nil {
		return nil, domain.ErrInstitutionNotFound
	}
	return institution, nil
}

func (s *accountService) GetInstitutionBySlug(ctx context.Context, slug string) (*domain.Institution, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.GetInstitutionBySlug(ctx, slug)
}

func (s *accountService) UpdateInstitution(ctx context.Context, institution *domain.Institution) error {
	if institution == nil {
		return errors.New("institution is required")
	}
	if institution.ID == "" {
		return errors.New("institution ID is required")
	}
	return s.repo.UpdateInstitution(ctx, institution)
}

func (s *accountService) DeleteInstitution(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("institution ID is required")
	}
	return s.repo.DeleteInstitution(ctx, id)
}

// ============================================================
// INSTITUTION TYPE OPERATIONS
// ============================================================

func (s *accountService) GetInstitutionTypeBySlug(ctx context.Context, slug string) (*domain.InstitutionType, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.GetInstitutionTypeBySlug(ctx, slug)
}

func (s *accountService) ListInstitutionTypes(ctx context.Context) ([]*domain.InstitutionType, error) {
	return s.repo.ListInstitutionTypes(ctx)
}

// ============================================================
// TEAM MEMBER OPERATIONS
// ============================================================

func (s *accountService) AddTeamMember(ctx context.Context, accountID, memberID, role string) (*domain.TeamMember, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}

	// Check if member already exists
	existing, err := s.repo.GetTeamMemberByAccountAndMember(ctx, accountID, memberID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrTeamMemberExists
	}

	// Create domain entity
	member, err := domain.NewTeamMember(accountID, memberID, role)
	if err != nil {
		return nil, err
	}

	// Save to database
	if err := s.repo.CreateTeamMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	// Assign role in authorization system (source of truth)
	if err := s.assignRole(ctx, accountID, memberID, role); err != nil {
		// Rollback: delete the team member if role assignment fails
		_ = s.repo.DeleteTeamMember(ctx, member.ID)
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return member, nil
}

func (s *accountService) GetTeamMembersByAccount(ctx context.Context, accountID string) ([]*domain.TeamMember, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	return s.repo.GetTeamMembersByAccount(ctx, accountID)
}

func (s *accountService) GetTeamMemberByAccountAndMember(ctx context.Context, accountID, memberID string) (*domain.TeamMember, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if memberID == "" {
		return nil, errors.New("member ID is required")
	}
	return s.repo.GetTeamMemberByAccountAndMember(ctx, accountID, memberID)
}

func (s *accountService) UpdateTeamMemberRole(ctx context.Context, id, newRole string) (*domain.TeamMember, error) {
	if id == "" {
		return nil, errors.New("team member ID is required")
	}
	if newRole == "" {
		return nil, errors.New("role is required")
	}

	member, err := s.repo.GetTeamMemberByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrTeamMemberNotFound
	}

	// Update role in authorization system (source of truth)
	if err := s.assignRole(ctx, member.AccountID, member.MemberID, newRole); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update role in domain entity for reference
	member.Role = newRole
	member.UpdatedAt = time.Now()
	if err := s.repo.UpdateTeamMember(ctx, member); err != nil {
		// Log but don't fail - authorization is already updated
	}

	return member, nil
}

func (s *accountService) RemoveTeamMember(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("team member ID is required")
	}

	member, err := s.repo.GetTeamMemberByID(ctx, id)
	if err != nil {
		return err
	}
	if member == nil {
		return domain.ErrTeamMemberNotFound
	}

	// Remove role from authorization system
	if err := s.permSvc.RemoveRole(ctx, member.AccountID, member.MemberID, member.Role); err != nil {
		// Log but continue - we still want to delete the team member
	}

	// Delete from database
	if err := s.repo.DeleteTeamMember(ctx, id); err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// assignRole assigns a role to a user in the authorization system
// This is the SINGLE SOURCE OF TRUTH for roles
func (s *accountService) assignRole(ctx context.Context, accountID, userID, role string) error {
	switch role {
	case "account_admin":
		return s.permSvc.AssignAccountAdminRole(ctx, accountID, userID)
	case "event_manager":
		return s.permSvc.AssignEventManagerRole(ctx, accountID, userID)
	case "team_member":
		return s.permSvc.AssignTeamMemberRole(ctx, accountID, userID)
	default:
		return domain.ErrInvalidRole
	}
}