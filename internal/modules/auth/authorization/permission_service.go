// internal/modules/auth/authorization/permission_service.go

package authorization

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// Ensure Service implements domain.PermissionService
var _ authdomain.PermissionService = (*Service)(nil)

// Service provides high-level permission management
// Implements domain.PermissionService (outbound adapter)
type Service struct {
	enforcer *Enforcer
}

// NewService creates a new permission service
func NewService(enforcer *Enforcer) authdomain.PermissionService {
	return &Service{
		enforcer: enforcer,
	}
}

// ============================================================
// ROLE ASSIGNMENT METHODS
// ============================================================

func (s *Service) AssignAccountAdminRole(ctx context.Context, accountID, userID string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Assigning account_admin role to user: %s for account: %s", userID, accountID)

	err := s.AddAccountPolicies(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	_, err = s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign account_admin role: %w", err)
	}

	return nil
}

func (s *Service) AssignEventManagerRole(ctx context.Context, accountID, userID string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Assigning event_manager role to user: %s for account: %s", userID, accountID)

	if err := s.ensureAccountPoliciesExist(ctx, accountID); err != nil {
		return fmt.Errorf("failed to ensure account policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign event_manager role: %w", err)
	}

	return nil
}

func (s *Service) AssignTeamMemberRole(ctx context.Context, accountID, userID string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Assigning team_member role to user: %s for account: %s", userID, accountID)

	if err := s.ensureAccountPoliciesExist(ctx, accountID); err != nil {
		return fmt.Errorf("failed to ensure account policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign team_member role: %w", err)
	}

	return nil
}

func (s *Service) AssignAdminRole(ctx context.Context, userID string) error {
	log.Printf("Assigning admin role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, authdomain.RoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}
	return nil
}

func (s *Service) AssignSuperAdminRole(ctx context.Context, userID string) error {
	log.Printf("Assigning super_admin role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, authdomain.RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("failed to assign super_admin role: %w", err)
	}
	return nil
}

// ============================================================
// ROLE REMOVAL METHODS
// ============================================================

func (s *Service) RemoveRole(ctx context.Context, accountID, userID string, role string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Removing role %s from user: %s for account: %s", role, userID, accountID)

	_, err := s.enforcer.RemoveRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

func (s *Service) RemoveAllAccountRoles(ctx context.Context, accountID, userID string) error {
	log.Printf("Removing all roles for user: %s from account: %s", userID, accountID)

	_, err := s.enforcer.RemoveAllAccountRoles(accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove all account roles: %w", err)
	}

	return nil
}

func (s *Service) RemovePlatformRole(ctx context.Context, userID string, role string) error {
	log.Printf("Removing platform role %s from user: %s", role, userID)
	_, err := s.enforcer.RemovePlatformRole(userID, authdomain.Role(role))
	if err != nil {
		return fmt.Errorf("failed to remove platform role: %w", err)
	}
	return nil
}

// ============================================================
// ACCOUNT POLICY MANAGEMENT
// ============================================================

func (s *Service) AddAccountPolicies(ctx context.Context, accountID string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Adding account policies for account: %s", accountID)

	allPolicies := GetAccountPolicies(domain)

	_, err := s.enforcer.AddPolicies(allPolicies)
	if err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	hierarchyPolicies := GetAccountRoleHierarchy(domain)
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("Added %d policies and %d hierarchy rules for account: %s",
		len(allPolicies), len(hierarchyPolicies), accountID)
	return nil
}

func (s *Service) ensureAccountPoliciesExist(ctx context.Context, accountID string) error {
	domain := authdomain.AccountDomain(accountID)

	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for account %s, adding default policies", accountID)
		return s.AddAccountPolicies(ctx, accountID)
	}

	return nil
}

func (s *Service) RemoveAccountPolicies(ctx context.Context, accountID string) error {
	domain := authdomain.AccountDomain(accountID)
	log.Printf("Removing policies for account: %s", accountID)

	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to get filtered policies: %w", err)
	}

	if len(policies) > 0 {
		_, err := s.enforcer.RemovePolicies(policies)
		if err != nil {
			return fmt.Errorf("failed to remove policies: %w", err)
		}
	}

	groupingPolicies, err := s.enforcer.GetFilteredGroupingPolicy(2, domain)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	if len(groupingPolicies) > 0 {
		_, err := s.enforcer.RemoveGroupingPolicies(groupingPolicies)
		if err != nil {
			return fmt.Errorf("failed to remove grouping policies: %w", err)
		}
	}

	log.Printf("Removed %d policies and %d grouping policies for account: %s",
		len(policies), len(groupingPolicies), accountID)
	return nil
}

func (s *Service) AddPlatformPolicies(ctx context.Context) error {
	log.Println("Adding platform policies")

	platformPolicies := GetPlatformPolicies()

	_, err := s.enforcer.AddPolicies(platformPolicies)
	if err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}

	hierarchyPolicies := GetPlatformRoleHierarchy()
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add platform role hierarchy: %w", err)
	}

	log.Printf("Added %d platform policies and %d hierarchy rules",
		len(platformPolicies), len(hierarchyPolicies))
	return nil
}

// ============================================================
// PERMISSION CHECK METHODS
// ============================================================

func (s *Service) HasPermission(ctx context.Context, userID, domain, resource, action string) bool {
	allowed, err := s.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return false
	}
	return allowed
}

func (s *Service) CanManageEvent(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String())
}

func (s *Service) CanIssueCertificate(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String())
}

func (s *Service) CanManageAccount(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, authdomain.ResourceAccount.String(), authdomain.ActionManage.String())
}

// ============================================================
// ACCOUNT ROLE CHECK METHODS
// ============================================================

func (s *Service) IsAccountAdmin(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
}

func (s *Service) IsEventManager(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain)
}

func (s *Service) IsTeamMember(ctx context.Context, accountID, userID string) bool {
	domain := authdomain.AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

func (s *Service) GetUserRoles(ctx context.Context, userID, domain string) []string {
	return s.enforcer.GetRolesForUserInDomain(userID, domain)
}

func (s *Service) GetUserAccounts(ctx context.Context, userID string) []string {
	return s.enforcer.GetUserAccountIDs(userID)
}

func (s *Service) HasAccountAccess(ctx context.Context, userID string) bool {
	return s.enforcer.HasAnyAccountRole(userID)
}

// ============================================================
// GET ROLES FOR USER (for token service)
// ============================================================

func (s *Service) GetRolesForUser(ctx context.Context, userID, domain string) ([]string, error) {
	roles := s.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
}