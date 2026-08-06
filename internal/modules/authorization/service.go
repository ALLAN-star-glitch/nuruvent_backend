// internal/modules/authorization/service.go

package authorization

import (
	"context"
	"fmt"
	"log"
)

// Service provides high-level permission management
type Service struct {
	enforcer *Enforcer
}

// NewService creates a new permission service
func NewService(enforcer *Enforcer) *Service {
	return &Service{
		enforcer: enforcer,
	}
}

// ============================================================
// ROLE ASSIGNMENT METHODS
// ============================================================

// AssignAccountAdminRole assigns the account_admin role to a user
func (s *Service) AssignAccountAdminRole(ctx context.Context, accountID, userID string) error {
	domain := AccountDomain(accountID)
	log.Printf("Assigning account_admin role to user: %s for account: %s", userID, accountID)

	// 1. Add account policies
	err := s.AddAccountPolicies(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	// 2. Assign the account_admin role
	_, err = s.enforcer.AddRoleForUserInDomain(userID, RoleAccountAdmin.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign account_admin role: %w", err)
	}

	return nil
}

// AssignEventManagerRole assigns the event manager role to a user
func (s *Service) AssignEventManagerRole(ctx context.Context, accountID, userID string) error {
	domain := AccountDomain(accountID)
	log.Printf("Assigning event_manager role to user: %s for account: %s", userID, accountID)

	if err := s.ensureAccountPoliciesExist(ctx, accountID); err != nil {
		return fmt.Errorf("failed to ensure account policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, RoleEventManager.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign event_manager role: %w", err)
	}

	return nil
}

// AssignTeamMemberRole assigns the team_member role to a user
func (s *Service) AssignTeamMemberRole(ctx context.Context, accountID, userID string) error {
	domain := AccountDomain(accountID)
	log.Printf("Assigning team_member role to user: %s for account: %s", userID, accountID)

	if err := s.ensureAccountPoliciesExist(ctx, accountID); err != nil {
		return fmt.Errorf("failed to ensure account policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, RoleTeamMember.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign team_member role: %w", err)
	}

	return nil
}

// AssignAdminRole assigns the admin role to a user (platform level)
func (s *Service) AssignAdminRole(ctx context.Context, userID string) error {
	log.Printf("Assigning admin role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, RoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}
	return nil
}

// AssignSuperAdminRole assigns the super_admin role to a user (platform level)
func (s *Service) AssignSuperAdminRole(ctx context.Context, userID string) error {
	log.Printf("Assigning super_admin role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, RoleSuperAdmin)
	if err != nil {
		return fmt.Errorf("failed to assign super_admin role: %w", err)
	}
	return nil
}

// ============================================================
// ROLE REMOVAL METHODS
// ============================================================

// RemoveRole removes a role from a user in an account
func (s *Service) RemoveRole(ctx context.Context, accountID, userID string, role Role) error {
	domain := AccountDomain(accountID)
	log.Printf("Removing role %s from user: %s for account: %s", role, userID, accountID)

	_, err := s.enforcer.RemoveRoleForUserInDomain(userID, role.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}


// RemoveAllAccountRoles removes all account roles from a user
func (s *Service) RemoveAllAccountRoles(ctx context.Context, accountID, userID string) error {
	log.Printf("Removing all roles for user: %s from account: %s", userID, accountID)

	_, err := s.enforcer.RemoveAllAccountRoles(accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove all account roles: %w", err)
	}

	return nil
}

// RemovePlatformRole removes a platform-level role from a user
func (s *Service) RemovePlatformRole(ctx context.Context, userID string, role Role) error {
	log.Printf("Removing platform role %s from user: %s", role, userID)
	_, err := s.enforcer.RemovePlatformRole(userID, role)
	if err != nil {
		return fmt.Errorf("failed to remove platform role: %w", err)
	}
	return nil
}

// ============================================================
// ACCOUNT POLICY MANAGEMENT
// ============================================================

// AddAccountPolicies adds default policies for a new account
func (s *Service) AddAccountPolicies(ctx context.Context, accountID string) error {
	domain := AccountDomain(accountID)
	log.Printf("Adding account policies for account: %s", accountID)

	// 1. Get all account policies
	allPolicies := GetAccountPolicies(domain)

	// 2. Add all policies
	_, err := s.enforcer.AddPolicies(allPolicies)
	if err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	// 3. Add role hierarchy
	hierarchyPolicies := GetAccountRoleHierarchy(domain)
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("Added %d policies and %d hierarchy rules for account: %s",
		len(allPolicies), len(hierarchyPolicies), accountID)
	return nil
}

// ensureAccountPoliciesExist checks if account policies exist and adds them if not
func (s *Service) ensureAccountPoliciesExist(ctx context.Context, accountID string) error {
	domain := AccountDomain(accountID)

	// Check if account policies already exist
	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	// If no policies exist, add them
	if len(policies) == 0 {
		log.Printf("No policies found for account %s, adding default policies", accountID)
		return s.AddAccountPolicies(ctx, accountID)
	}

	return nil
}

// RemoveAccountPolicies removes all policies for an account
func (s *Service) RemoveAccountPolicies(ctx context.Context, accountID string) error {
	domain := AccountDomain(accountID)
	log.Printf("Removing policies for account: %s", accountID)

	// 1. Get all policies for this domain
	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to get filtered policies: %w", err)
	}

	// 2. Remove policies
	if len(policies) > 0 {
		_, err := s.enforcer.RemovePolicies(policies)
		if err != nil {
			return fmt.Errorf("failed to remove policies: %w", err)
		}
	}

	// 3. Get all grouping policies for this domain
	groupingPolicies, err := s.enforcer.GetFilteredGroupingPolicy(2, domain)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	// 4. Remove grouping policies
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

// AddPlatformPolicies adds platform-level policies
func (s *Service) AddPlatformPolicies(ctx context.Context) error {
	log.Println("Adding platform policies")

	// 1. Get platform policies
	platformPolicies := GetPlatformPolicies()

	// 2. Add all policies
	_, err := s.enforcer.AddPolicies(platformPolicies)
	if err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}

	// 3. Add platform role hierarchy
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

// HasPermission checks if a user has a specific permission in a domain
func (s *Service) HasPermission(ctx context.Context, userID, domain string, resource Resource, action Action) bool {
	allowed, err := s.enforcer.Enforce(userID, domain, resource.String(), action.String())
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return false
	}
	return allowed
}

// CanManageEvent checks if user can manage events in an account
func (s *Service) CanManageEvent(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, ResourceEvent, ActionManage)
}

// CanIssueCertificate checks if user can issue certificates in an account
func (s *Service) CanIssueCertificate(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, ResourceCertificate, ActionIssue)
}

// CanManageAccount checks if user can manage an account
func (s *Service) CanManageAccount(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.HasPermission(ctx, userID, domain, ResourceAccount, ActionManage)
}

// ============================================================
// ACCOUNT ROLE CHECK METHODS
// ============================================================

// IsAccountAdmin checks if user is an account admin
func (s *Service) IsAccountAdmin(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleAccountAdmin.String(), domain)
}

// IsEventManager checks if user is an event manager
func (s *Service) IsEventManager(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleEventManager.String(), domain)
}

// IsTeamMember checks if user is a team member
func (s *Service) IsTeamMember(ctx context.Context, accountID, userID string) bool {
	domain := AccountDomain(accountID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleTeamMember.String(), domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

// GetUserRoles returns all roles for a user in a domain
func (s *Service) GetUserRoles(ctx context.Context, userID string, domain string) []string {
	return s.enforcer.GetRolesForUserInDomain(userID, domain)
}

// GetUserAccounts returns all accounts a user is a member of
func (s *Service) GetUserAccounts(ctx context.Context, userID string) []string {
	return s.enforcer.GetUserAccountIDs(userID)
}

// HasAccountAccess checks if user has any account role
func (s *Service) HasAccountAccess(ctx context.Context, userID string) bool {
	return s.enforcer.HasAnyAccountRole(userID)
}