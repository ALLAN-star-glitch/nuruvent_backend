package bizservice

import (
	"context"
	"fmt"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizrepo"
	"github.com/google/uuid"
)

type MemberService struct {
	memberRepo   *bizrepo.BusinessMemberRepository
	businessRepo *bizrepo.BusinessRepository
	enforcer     *authorization.Enforcer
	permService  *authorization.Service
}

func NewMemberService(
	memberRepo *bizrepo.BusinessMemberRepository,
	businessRepo *bizrepo.BusinessRepository,
	enforcer *authorization.Enforcer,
	permService *authorization.Service,
) *MemberService {
	return &MemberService{
		memberRepo:   memberRepo,
		businessRepo: businessRepo,
		enforcer:     enforcer,
		permService:  permService,
	}
}

// ================================================
// MEMBER MANAGEMENT
// ================================================

// AddMember adds a member to a business
func (s *MemberService) AddMember(ctx context.Context, adminID, businessID, userID uuid.UUID, role string) (*models.BusinessMember, error) {
	// Check if admin has permission to manage members
	canManage, err := s.enforcer.Enforce(adminID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceMember.String(), authorization.ActionCreate.String())
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canManage {
		return nil, fmt.Errorf("insufficient permissions to add members")
	}

	// Validate role
	if !s.isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s. Valid roles: host, event_manager, member", role)
	}

	// Check if user is already a member
	exists, err := s.memberRepo.Exists(userID, businessID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user is already a member of this business")
	}

	// Create member
	member := &models.BusinessMember{
		BusinessID: businessID,
		UserID:     userID,
		Role:       role,
		IsActive:   true,
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Assign role in Casbin
	if err := s.assignCasbinRole(ctx, userID.String(), businessID.String(), role); err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return member, nil
}

// RemoveMember removes a member from a business
func (s *MemberService) RemoveMember(ctx context.Context, adminID, businessID, userID uuid.UUID) error {
	// Check if admin has permission to manage members
	canManage, err := s.enforcer.Enforce(adminID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceMember.String(), authorization.ActionDelete.String())
	if err != nil {
		return fmt.Errorf("authorization error: %w", err)
	}
	if !canManage {
		return fmt.Errorf("insufficient permissions to remove members")
	}

	// Get member to check if exists
	member, err := s.memberRepo.GetByUserAndBusiness(userID, businessID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("member not found")
	}

	// Remove all business roles from user
	if err := s.permService.RemoveAllBusinessRoles(ctx, userID.String(), businessID.String()); err != nil {
		return fmt.Errorf("failed to remove roles: %w", err)
	}

	// Delete member
	if err := s.memberRepo.Delete(userID, businessID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// UpdateMemberRole updates a member's role
func (s *MemberService) UpdateMemberRole(ctx context.Context, adminID, businessID, userID uuid.UUID, newRole string) error {
	// Check if admin has permission to manage members
	canManage, err := s.enforcer.Enforce(adminID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceMember.String(), authorization.ActionUpdate.String())
	if err != nil {
		return fmt.Errorf("authorization error: %w", err)
	}
	if !canManage {
		return fmt.Errorf("insufficient permissions to update member role")
	}

	// Validate role
	if !s.isValidRole(newRole) {
		return fmt.Errorf("invalid role: %s. Valid roles: host, event_manager, member", newRole)
	}

	// Check if member exists
	member, err := s.memberRepo.GetByUserAndBusiness(userID, businessID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("member not found")
	}

	// Remove old roles
	if err := s.permService.RemoveAllBusinessRoles(ctx, userID.String(), businessID.String()); err != nil {
		return fmt.Errorf("failed to remove old roles: %w", err)
	}

	// Assign new role
	if err := s.assignCasbinRole(ctx, userID.String(), businessID.String(), newRole); err != nil {
		return fmt.Errorf("failed to assign new role: %w", err)
	}

	// Update member role
	if err := s.memberRepo.UpdateRole(userID, businessID, newRole); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetBusinessMembers gets all members of a business
func (s *MemberService) GetBusinessMembers(ctx context.Context, userID, businessID uuid.UUID) ([]models.BusinessMember, error) {
	// Check if user has access to view members
	canRead, err := s.enforcer.Enforce(userID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceMember.String(), authorization.ActionRead.String())
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canRead {
		return nil, fmt.Errorf("insufficient permissions to view members")
	}

	return s.memberRepo.GetMembersByBusiness(businessID)
}

// CheckMembership checks if a user is a member of a business
func (s *MemberService) CheckMembership(ctx context.Context, userID, businessID uuid.UUID) (bool, string, error) {
	member, err := s.memberRepo.GetByUserAndBusiness(userID, businessID)
	if err != nil {
		return false, "", err
	}
	if member == nil {
		return false, "", nil
	}
	return true, member.Role, nil
}

// GetMyBusinessesWithRole gets all businesses a user belongs to with their roles
func (s *MemberService) GetMyBusinessesWithRole(ctx context.Context, userID uuid.UUID) ([]models.BusinessMember, error) {
	return s.memberRepo.GetBusinessesByUser(userID)
}

// ================================================
// HELPER METHODS
// ================================================

// isValidRole checks if a role is valid
func (s *MemberService) isValidRole(role string) bool {
	validRoles := map[string]bool{
		"host":          true,
		"event_manager": true,
		"member":        true,
	}
	return validRoles[role]
}

// assignCasbinRole assigns a role in Casbin
func (s *MemberService) assignCasbinRole(ctx context.Context, userID, businessID, role string) error {
	switch role {
	case "host":
		return s.permService.AssignHostRole(ctx, userID, businessID)
	case "event_manager":
		return s.permService.AssignEventManagerRole(ctx, userID, businessID)
	case "member":
		return s.permService.AssignMemberRole(ctx, userID, businessID)
	default:
		return fmt.Errorf("invalid role: %s", role)
	}
}