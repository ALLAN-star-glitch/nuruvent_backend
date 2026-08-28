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
// TEAM ROLE ASSIGNMENT
// ============================================================

// AssignPersonalTeamAdmin assigns account_admin role to a user for their personal team
// Domain: personal:team:{user_id}
func (s *Service) AssignPersonalTeamAdmin(ctx context.Context, userID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Assigning account_admin role to user: %s for personal team", userID)

	// Add personal team policies first
	if err := s.AddPersonalTeamPolicies(ctx, userID); err != nil {
		return fmt.Errorf("failed to add personal team policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign account_admin role: %w", err)
	}

	log.Printf("✅ Assigned account_admin for personal team to user: %s", userID)
	return nil
}

// AssignInstitutionAdminRole assigns account_admin role to a user for an institution team
// Domain: institution:team:{institution_id}
func (s *Service) AssignInstitutionAdminRole(ctx context.Context, institutionID, userID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Assigning account_admin role to user: %s for institution: %s", userID, institutionID)

	// Add institution policies first
	if err := s.AddInstitutionPolicies(ctx, institutionID); err != nil {
		return fmt.Errorf("failed to add institution policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign account_admin role: %w", err)
	}

	log.Printf("✅ Assigned account_admin for institution %s to user: %s", institutionID, userID)
	return nil
}

// AssignEventManagerRoleForPersonalTeam assigns event_manager role to a user for their personal team
// Domain: personal:team:{user_id}
func (s *Service) AssignEventManagerRoleForPersonalTeam(ctx context.Context, userID, targetUserID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Assigning event_manager role to user: %s for personal team: %s", targetUserID, userID)

	// Ensure personal team policies exist
	if err := s.ensurePersonalTeamPoliciesExist(ctx, userID); err != nil {
		return fmt.Errorf("failed to ensure personal team policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(targetUserID, authdomain.RoleEventManager.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign event_manager role: %w", err)
	}

	log.Printf("✅ Assigned event_manager for personal team %s to user: %s", userID, targetUserID)
	return nil
}

// AssignTeamMemberRoleForPersonalTeam assigns team_member role to a user for their personal team
// Domain: personal:team:{user_id}
func (s *Service) AssignTeamMemberRoleForPersonalTeam(ctx context.Context, userID, targetUserID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Assigning team_member role to user: %s for personal team: %s", targetUserID, userID)

	// Ensure personal team policies exist
	if err := s.ensurePersonalTeamPoliciesExist(ctx, userID); err != nil {
		return fmt.Errorf("failed to ensure personal team policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(targetUserID, authdomain.RoleTeamMember.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign team_member role: %w", err)
	}

	log.Printf("✅ Assigned team_member for personal team %s to user: %s", userID, targetUserID)
	return nil
}

// AssignEventManagerRoleForInstitution assigns event_manager role to a user for an institution team
// Domain: institution:team:{institution_id}
func (s *Service) AssignEventManagerRoleForInstitution(ctx context.Context, institutionID, userID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Assigning event_manager role to user: %s for institution: %s", userID, institutionID)

	// Ensure institution team policies exist
	if err := s.ensureInstitutionPoliciesExist(ctx, institutionID); err != nil {
		return fmt.Errorf("failed to ensure institution policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign event_manager role: %w", err)
	}

	log.Printf("✅ Assigned event_manager for institution %s to user: %s", institutionID, userID)
	return nil
}

// AssignTeamMemberRoleForInstitution assigns team_member role to a user for an institution team
// Domain: institution:team:{institution_id}
func (s *Service) AssignTeamMemberRoleForInstitution(ctx context.Context, institutionID, userID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Assigning team_member role to user: %s for institution: %s", userID, institutionID)

	// Ensure institution team policies exist
	if err := s.ensureInstitutionPoliciesExist(ctx, institutionID); err != nil {
		return fmt.Errorf("failed to ensure institution policies: %w", err)
	}

	_, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain)
	if err != nil {
		return fmt.Errorf("failed to assign team_member role: %w", err)
	}

	log.Printf("✅ Assigned team_member for institution %s to user: %s", institutionID, userID)
	return nil
}

// ============================================================
// TEAM POLICY MANAGEMENT
// ============================================================

// AddPersonalTeamPolicies adds default policies for a personal team
// Domain: personal:team:{user_id}
func (s *Service) AddPersonalTeamPolicies(ctx context.Context, userID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Adding personal team policies for user: %s", userID)

	allPolicies := GetPersonalTeamPolicies(domain)

	_, err := s.enforcer.AddPolicies(allPolicies)
	if err != nil {
		return fmt.Errorf("failed to add personal team policies: %w", err)
	}

	hierarchyPolicies := GetTeamRoleHierarchy(domain)
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("Added %d policies and %d hierarchy rules for personal team: %s",
		len(allPolicies), len(hierarchyPolicies), userID)
	return nil
}

// AddInstitutionPolicies adds default policies for an institution team
// Domain: institution:team:{institution_id}
func (s *Service) AddInstitutionPolicies(ctx context.Context, institutionID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Adding institution policies for institution: %s", institutionID)

	allPolicies := GetInstitutionTeamPolicies(domain)

	_, err := s.enforcer.AddPolicies(allPolicies)
	if err != nil {
		return fmt.Errorf("failed to add institution policies: %w", err)
	}

	hierarchyPolicies := GetTeamRoleHierarchy(domain)
	_, err = s.enforcer.AddGroupingPolicies(hierarchyPolicies)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("Added %d policies and %d hierarchy rules for institution: %s",
		len(allPolicies), len(hierarchyPolicies), institutionID)
	return nil
}

// RemovePersonalTeamPolicies removes all policies for a personal team
func (s *Service) RemovePersonalTeamPolicies(ctx context.Context, userID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	return s.removeTeamPolicies(ctx, domain)
}

// RemoveInstitutionTeamPolicies removes all policies for an institution team
func (s *Service) RemoveInstitutionTeamPolicies(ctx context.Context, institutionID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return s.removeTeamPolicies(ctx, domain)
}

// removeTeamPolicies removes all policies for a team domain
func (s *Service) removeTeamPolicies(ctx context.Context, domain string) error {
	log.Printf("Removing policies for domain: %s", domain)

	// Remove policy rules
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

	// Remove grouping policies (role assignments)
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

	log.Printf("Removed %d policies and %d grouping policies for domain: %s",
		len(policies), len(groupingPolicies), domain)
	return nil
}

// ensurePersonalTeamPoliciesExist ensures policies exist for a personal team
func (s *Service) ensurePersonalTeamPoliciesExist(ctx context.Context, userID string) error {
	domain := authdomain.PersonalTeamDomain(userID)

	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for personal team %s, adding default policies", userID)
		if err := s.AddPersonalTeamPolicies(ctx, userID); err != nil {
			return fmt.Errorf("failed to add personal team policies: %w", err)
		}
	}

	return nil
}

// ensureInstitutionPoliciesExist ensures policies exist for an institution team
func (s *Service) ensureInstitutionPoliciesExist(ctx context.Context, institutionID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)

	policies, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for institution %s, adding default policies", institutionID)
		if err := s.AddInstitutionPolicies(ctx, institutionID); err != nil {
			return fmt.Errorf("failed to add institution policies: %w", err)
		}
	}

	return nil
}

// ============================================================
// PLATFORM ROLE METHODS
// ============================================================

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

func (s *Service) RemovePlatformRole(ctx context.Context, userID string, role string) error {
	log.Printf("Removing platform role %s from user: %s", role, userID)
	_, err := s.enforcer.RemovePlatformRole(userID, authdomain.Role(role))
	if err != nil {
		return fmt.Errorf("failed to remove platform role: %w", err)
	}
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
// ROLE REMOVAL METHODS
// ============================================================

// RemovePersonalTeamRole removes a role from a user's personal team
func (s *Service) RemovePersonalTeamRole(ctx context.Context, userID, targetUserID string, role string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Removing role %s from user: %s for personal team: %s", role, targetUserID, userID)

	_, err := s.enforcer.RemoveRoleForUserInDomain(targetUserID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// RemoveInstitutionTeamRole removes a role from a user's institution team
func (s *Service) RemoveInstitutionTeamRole(ctx context.Context, institutionID, userID string, role string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Removing role %s from user: %s for institution: %s", role, userID, institutionID)

	_, err := s.enforcer.RemoveRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// RemoveAllPersonalTeamRoles removes all roles from a user's personal team
func (s *Service) RemoveAllPersonalTeamRoles(ctx context.Context, userID, targetUserID string) error {
	domain := authdomain.PersonalTeamDomain(userID)
	log.Printf("Removing all roles for user: %s from personal team: %s", targetUserID, userID)

	_, err := s.enforcer.RemoveFilteredGroupingPolicy(0, targetUserID, "", domain)
	if err != nil {
		return fmt.Errorf("failed to remove all team roles: %w", err)
	}

	return nil
}

// RemoveAllInstitutionTeamRoles removes all roles from a user's institution team
func (s *Service) RemoveAllInstitutionTeamRoles(ctx context.Context, institutionID, userID string) error {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	log.Printf("Removing all roles for user: %s from institution: %s", userID, institutionID)

	_, err := s.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	if err != nil {
		return fmt.Errorf("failed to remove all team roles: %w", err)
	}

	return nil
}

// ============================================================
// TEAM PERMISSION CHECKS
// ============================================================

// HasPersonalTeamPermission checks if user has permission in a personal team
func (s *Service) HasPersonalTeamPermission(ctx context.Context, userID, teamOwnerID, resource, action string) bool {
	domain := authdomain.PersonalTeamDomain(teamOwnerID)
	allowed, err := s.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return false
	}
	return allowed
}

// HasInstitutionTeamPermission checks if user has permission in an institution team
func (s *Service) HasInstitutionTeamPermission(ctx context.Context, userID, institutionID, resource, action string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	allowed, err := s.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return false
	}
	return allowed
}

func (s *Service) CanManageTeamEvent(ctx context.Context, teamID, userID string) bool {
	// Try personal team
	if s.HasPersonalTeamPermission(ctx, userID, teamID, authdomain.ResourceEvent.String(), authdomain.ActionManage.String()) {
		return true
	}
	// Try institution team
	return s.HasInstitutionTeamPermission(ctx, userID, teamID, authdomain.ResourceEvent.String(), authdomain.ActionManage.String())
}

func (s *Service) CanIssueTeamCertificate(ctx context.Context, teamID, userID string) bool {
	if s.HasPersonalTeamPermission(ctx, userID, teamID, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String()) {
		return true
	}
	return s.HasInstitutionTeamPermission(ctx, userID, teamID, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String())
}

func (s *Service) CanManageTeam(ctx context.Context, teamID, userID string) bool {
	if s.HasPersonalTeamPermission(ctx, userID, teamID, authdomain.ResourceTeam.String(), authdomain.ActionManage.String()) {
		return true
	}
	return s.HasInstitutionTeamPermission(ctx, userID, teamID, authdomain.ResourceTeam.String(), authdomain.ActionManage.String())
}

// ============================================================
// TEAM ROLE CHECKS
// ============================================================

// IsPersonalTeamAdmin checks if user is admin of a personal team
func (s *Service) IsPersonalTeamAdmin(ctx context.Context, userID, teamOwnerID string) bool {
	domain := authdomain.PersonalTeamDomain(teamOwnerID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
}

// IsInstitutionTeamAdmin checks if user is admin of an institution team
func (s *Service) IsInstitutionTeamAdmin(ctx context.Context, userID, institutionID string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
}

// IsPersonalTeamEventManager checks if user is an event manager in a personal team
func (s *Service) IsPersonalTeamEventManager(ctx context.Context, userID, teamOwnerID string) bool {
	domain := authdomain.PersonalTeamDomain(teamOwnerID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain)
}

// IsInstitutionTeamEventManager checks if user is an event manager in an institution team
func (s *Service) IsInstitutionTeamEventManager(ctx context.Context, userID, institutionID string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain)
}

// IsPersonalTeamMember checks if user is a member of a personal team
func (s *Service) IsPersonalTeamMember(ctx context.Context, userID, teamOwnerID string) bool {
	domain := authdomain.PersonalTeamDomain(teamOwnerID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain)
}

// IsInstitutionTeamMember checks if user is a member of an institution team
func (s *Service) IsInstitutionTeamMember(ctx context.Context, userID, institutionID string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

func (s *Service) GetUserRoles(ctx context.Context, userID, domain string) ([]string, error) {
	roles := s.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
}

func (s *Service) GetUserTeamIDs(ctx context.Context, userID string) []string {
	return s.enforcer.GetUserTeamIDs(userID)
}

func (s *Service) GetUserPersonalTeamIDs(ctx context.Context, userID string) []string {
	return s.enforcer.GetUserPersonalTeamIDs(userID)
}

func (s *Service) GetUserInstitutionTeamIDs(ctx context.Context, userID string) []string {
	return s.enforcer.GetUserInstitutionTeamIDs(userID)
}

func (s *Service) HasTeamAccess(ctx context.Context, userID string) bool {
	return s.enforcer.HasAnyTeamRole(userID)
}

// ============================================================
// GET ROLES FOR USER (for token service)
// ============================================================

func (s *Service) GetRolesForUser(ctx context.Context, userID, domain string) ([]string, error) {
	roles := s.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
}