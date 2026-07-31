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

// AssignBusinessAdminRole assigns the business_admin role to a user
func (s *Service) AssignBusinessAdminRole(ctx context.Context, userID string, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Assigning business_admin role to user: %s for business: %s", userID, businessID)

	// 1. Add business policies
	err := s.AddBusinessPolicies(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to add business policies: %w", err)
	}

	// 2. Assign the business_admin role
	_, err = s.enforcer.AddRoleForUserInDomain(userID, RoleBusinessAdmin.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign business_admin role: %w", err)
	}

	return nil
}

// AssignEventManagerRole assigns the event manager role to a user
func (s *Service) AssignEventManagerRole(ctx context.Context, userID string, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Assigning event_manager role to user: %s for business: %s", userID, businessID)

	if err := s.ensureBusinessPoliciesExist(ctx, businessID); err != nil {
		return fmt.Errorf("failed to ensure business policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, RoleEventManager.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign event_manager role: %w", err)
	}

	return nil
}

// AssignMemberRole assigns the member role to a user
func (s *Service) AssignMemberRole(ctx context.Context, userID string, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Assigning member role to user: %s for business: %s", userID, businessID)

	if err := s.ensureBusinessPoliciesExist(ctx, businessID); err != nil {
		return fmt.Errorf("failed to ensure business policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, RoleMember.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign member role: %w", err)
	}

	return nil
}

// AssignAttendeeRole assigns the attendee role to a user (platform level)
func (s *Service) AssignAttendeeRole(ctx context.Context, userID string) error {
	log.Printf("Assigning attendee role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, RoleAttendee)
	if err != nil {
		return fmt.Errorf("failed to assign attendee role: %w", err)
	}
	return nil
}

// AssignPremiumAttendeeRole assigns the premium_attendee role to a user (platform level)
func (s *Service) AssignPremiumAttendeeRole(ctx context.Context, userID string) error {
	log.Printf("Assigning premium_attendee role to user: %s", userID)
	_, err := s.enforcer.AddPlatformRole(userID, RolePremiumAttendee)
	if err != nil {
		return fmt.Errorf("failed to assign premium_attendee role: %w", err)
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

// RemoveRole removes a role from a user in a business
func (s *Service) RemoveRole(ctx context.Context, userID string, businessID string, role Role) error {
	domain := BusinessDomain(businessID)
	log.Printf("Removing role %s from user: %s for business: %s", role, userID, businessID)

	_, err := s.enforcer.RemoveRoleForUserInDomain(userID, role.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// RemoveAllBusinessRoles removes all business roles from a user
func (s *Service) RemoveAllBusinessRoles(ctx context.Context, userID string, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Removing all roles for user: %s from business: %s", userID, businessID)

	roles := s.enforcer.GetRolesForUserInDomain(userID, domain)

	for _, role := range roles {
		_, err := s.enforcer.RemoveRoleForUserInDomain(userID, role, domain)
		if err != nil {
			return fmt.Errorf("failed to remove role %s: %w", role, err)
		}
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
// BUSINESS POLICY MANAGEMENT
// ============================================================

// AddBusinessPolicies adds default policies for a new business
func (s *Service) AddBusinessPolicies(ctx context.Context, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Adding business policies for business: %s", businessID)

	// 1. Get all business policies
	allPolicies := GetAllBusinessPolicies(domain)

	// 2. Add all policies
	_, err := s.enforcer.AddPolicies(allPolicies)
	if err != nil {
		return fmt.Errorf("failed to add business policies: %w", err)
	}

	// 3. Add role hierarchy
	hierarchyPolicies := GetBusinessRoleHierarchy(domain)
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("Added %d policies and %d hierarchy rules for business: %s",
		len(allPolicies), len(hierarchyPolicies), businessID)
	return nil
}

// ensureBusinessPoliciesExist checks if business policies exist and adds them if not
func (s *Service) ensureBusinessPoliciesExist(ctx context.Context, businessID string) error {
	domain := BusinessDomain(businessID)

	// Check if business policies already exist
	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	// If no policies exist, add them
	if len(policies) == 0 {
		log.Printf("No policies found for business %s, adding default policies", businessID)
		return s.AddBusinessPolicies(ctx, businessID)
	}

	return nil
}

// RemoveBusinessPolicies removes all policies for a business
func (s *Service) RemoveBusinessPolicies(ctx context.Context, businessID string) error {
	domain := BusinessDomain(businessID)
	log.Printf("Removing policies for business: %s", businessID)

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

	log.Printf("Removed %d policies and %d grouping policies for business: %s",
		len(policies), len(groupingPolicies), businessID)
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

// CanManageEvent checks if user can manage events in a business
func (s *Service) CanManageEvent(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.HasPermission(ctx, userID, domain, ResourceEvent, ActionManage)
}

// CanIssueCertificate checks if user can issue certificates in a business
func (s *Service) CanIssueCertificate(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.HasPermission(ctx, userID, domain, ResourceCertificate, ActionIssue)
}

// CanManageBusiness checks if user can manage a business
func (s *Service) CanManageBusiness(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.HasPermission(ctx, userID, domain, ResourceBusiness, ActionManage)
}

// ============================================================
// BUSINESS ROLE CHECK METHODS
// ============================================================

// IsBusinessAdmin checks if user is a business admin
func (s *Service) IsBusinessAdmin(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleBusinessAdmin.String(), domain)
}

// IsEventManager checks if user is an event manager
func (s *Service) IsEventManager(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleEventManager.String(), domain)
}

// IsMember checks if user is a member
func (s *Service) IsMember(ctx context.Context, userID, businessID string) bool {
	domain := BusinessDomain(businessID)
	return s.enforcer.HasRoleForUserInDomain(userID, RoleMember.String(), domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

// GetUserRoles returns all roles for a user in a domain
func (s *Service) GetUserRoles(ctx context.Context, userID string, domain string) []string {
	return s.enforcer.GetRolesForUserInDomain(userID, domain)
}

// GetUserBusinesses returns all businesses a user is a member of
func (s *Service) GetUserBusinesses(ctx context.Context, userID string) []string {
	return s.enforcer.GetDomainsForUser(userID)
}

// HasBusinessAccess checks if user has any business role
func (s *Service) HasBusinessAccess(ctx context.Context, userID string) bool {
	return s.enforcer.HasAnyBusinessRole(userID)
}

// ============================================================
// DEPRECATED METHODS (for backward compatibility - will be removed)
// ============================================================

// AssignHostRole is deprecated. Use AssignBusinessAdminRole instead.
// Deprecated: Use AssignBusinessAdminRole
func (s *Service) AssignHostRole(ctx context.Context, userID string, businessID string) error {
	return s.AssignBusinessAdminRole(ctx, userID, businessID)
}

// IsHost is deprecated. Use IsBusinessAdmin instead.
// Deprecated: Use IsBusinessAdmin
func (s *Service) IsHost(ctx context.Context, userID, businessID string) bool {
	return s.IsBusinessAdmin(ctx, userID, businessID)
}