// internal/modules/auth/authorization/enforcer.go

package authorization

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// ============================================================
// ENFORCER - Internal implementation detail
// ============================================================

// Enforcer wraps the Casbin enforcer with additional functionality
type Enforcer struct {
	*casbin.Enforcer
	mu      sync.RWMutex
	cfg     *config.CasbinConfig
	db      *gorm.DB
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

func NewEnforcer(db *gorm.DB, cfg *config.Config) (*Enforcer, error) {
	// Create GORM adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Create enforcer with model and adapter
	e, err := casbin.NewEnforcer(cfg.Casbin.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Enable auto-save for policy changes
	e.EnableAutoSave(true)

	// Load policies from database with error handling
	err = e.LoadPolicy()
	if err != nil {
		log.Printf("⚠️ Error loading policies: %v", err)
		log.Printf("ℹ️ Attempting to clean up invalid policies...")

		if cleanErr := cleanInvalidPolicies(db); cleanErr != nil {
			return nil, fmt.Errorf("failed to clean invalid policies: %w", cleanErr)
		}

		if err := e.LoadPolicy(); err != nil {
			return nil, fmt.Errorf("failed to load policies after cleanup: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	enforcer := &Enforcer{
		Enforcer: e,
		cfg:      &cfg.Casbin,
		db:       db,
		ctx:      ctx,
		cancel:   cancel,
	}

	log.Println("Casbin enforcer initialized successfully")

	if cfg.Casbin.AutoLoad {
		log.Printf("🚀 Auto-load enabled (interval: %v)", cfg.Casbin.AutoLoadInterval)
		go enforcer.autoLoadPolicies()
	} else {
		log.Println("ℹ️ Auto-load is disabled")
	}

	return enforcer, nil
}

// cleanInvalidPolicies removes invalid policy rules from the database
func cleanInvalidPolicies(db *gorm.DB) error {
	result := db.Exec(`
		DELETE FROM casbin_rule 
		WHERE ptype = 'p' 
		AND (v3 IS NULL OR v3 = '')
	`)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("✅ Removed %d invalid policy rules", result.RowsAffected)
	return nil
}

// Close stops the enforcer and cleans up resources
func (e *Enforcer) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.stopped {
		e.cancel()
		e.stopped = true
	}
}

// autoLoadPolicies periodically reloads policies from the database
func (e *Enforcer) autoLoadPolicies() {
	ticker := time.NewTicker(e.cfg.AutoLoadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			log.Println("Auto-load stopped")
			return
		case <-ticker.C:
			e.mu.Lock()
			err := e.LoadPolicy()
			if err != nil {
				log.Printf("Failed to auto-load policies: %v", err)
			} else {
				log.Println("Policies auto-loaded successfully")
			}
			e.mu.Unlock()
		}
	}
}

// ================================================
// PERMISSION CHECK METHODS
// ================================================

// Enforce checks if a user has permission in a domain
func (e *Enforcer) Enforce(userID string, domain string, resource string, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// ✅ Super Admin bypass - has access to everything
	if e.IsSuperAdmin(userID) {
		return true, nil
	}

	return e.Enforcer.Enforce(userID, domain, resource, action)
}

// BatchEnforce checks multiple permissions in one call
func (e *Enforcer) BatchEnforce(requests [][]interface{}) ([]bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.BatchEnforce(requests)
}

// ================================================
// POLICY MANAGEMENT METHODS
// ================================================

// AddPolicy adds a new policy rule
func (e *Enforcer) AddPolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddPolicy(sub, dom, obj, act)
}

// RemovePolicy removes a policy rule
func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemovePolicy(sub, dom, obj, act)
}

// AddPolicies adds multiple policy rules
func (e *Enforcer) AddPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddPolicies(rules)
}

// RemovePolicies removes multiple policy rules
func (e *Enforcer) RemovePolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemovePolicies(rules)
}

// AddGroupingPolicies adds multiple grouping policies (role assignments)
func (e *Enforcer) AddGroupingPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddGroupingPolicies(rules)
}

// RemoveGroupingPolicies removes multiple grouping policies
func (e *Enforcer) RemoveGroupingPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveGroupingPolicies(rules)
}

// HasPolicy checks if a policy exists
func (e *Enforcer) HasPolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.HasPolicy(sub, dom, obj, act)
}

// HasGroupingPolicy checks if a grouping policy exists (role assignment)
func (e *Enforcer) HasGroupingPolicy(user, role, domain string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.HasGroupingPolicy(user, role, domain)
}

// GetPolicy returns all policies
func (e *Enforcer) GetPolicy() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetPolicy()
}

// GetGroupingPolicy returns all grouping policies (role assignments)
func (e *Enforcer) GetGroupingPolicy() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetGroupingPolicy()
}

// GetFilteredPolicy gets policies filtered by field values
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}

// GetFilteredGroupingPolicy gets grouping policies filtered by field values
func (e *Enforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// ================================================
// ROLE MANAGEMENT METHODS
// ================================================

// AddRoleForUserInDomain adds a role for a user in a specific domain
func (e *Enforcer) AddRoleForUserInDomain(userID, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddGroupingPolicy(userID, role, domain)
}

// RemoveRoleForUserInDomain removes a role for a user in a specific domain
func (e *Enforcer) RemoveRoleForUserInDomain(userID, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveGroupingPolicy(userID, role, domain)
}

// GetRolesForUserInDomain returns all roles for a user in a domain
func (e *Enforcer) GetRolesForUserInDomain(userID, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetRolesForUserInDomain(userID, domain)
}

// GetImplicitRolesForUser returns all roles for a user including inherited ones
func (e *Enforcer) GetImplicitRolesForUser(userID string, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetImplicitRolesForUser(userID, domain)
}

// GetImplicitPermissionsForUser returns all permissions for a user including inherited ones
func (e *Enforcer) GetImplicitPermissionsForUser(userID string, domain string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetImplicitPermissionsForUser(userID, domain)
}

// HasRoleForUserInDomain checks if a user has a specific role in a domain
func (e *Enforcer) HasRoleForUserInDomain(userID, role, domain string) bool {
	roles := e.GetRolesForUserInDomain(userID, domain)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasImplicitRoleForUserInDomain checks if a user has a role (including inherited) in a domain
func (e *Enforcer) HasImplicitRoleForUserInDomain(userID, role, domain string) (bool, error) {
	roles, err := e.GetImplicitRolesForUser(userID, domain)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// ================================================
// PLATFORM ROLE METHODS
// ================================================

// AddPlatformRole adds a platform-level role for a user
func (e *Enforcer) AddPlatformRole(userID string, role authdomain.Role) (bool, error) {
	return e.AddRoleForUserInDomain(userID, role.String(), authdomain.DomainPlatform)
}

// RemovePlatformRole removes a platform-level role from a user
func (e *Enforcer) RemovePlatformRole(userID string, role authdomain.Role) (bool, error) {
	return e.RemoveRoleForUserInDomain(userID, role.String(), authdomain.DomainPlatform)
}

// GetUserPlatformRoles returns all platform-level roles for a user
func (e *Enforcer) GetUserPlatformRoles(userID string) []string {
	return e.GetRolesForUserInDomain(userID, authdomain.DomainPlatform)
}

// HasPlatformRole checks if a user has a specific platform role
func (e *Enforcer) HasPlatformRole(userID, role string) bool {
	return e.HasRoleForUserInDomain(userID, role, authdomain.DomainPlatform)
}

// IsSuperAdmin checks if a user is a super admin
func (e *Enforcer) IsSuperAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform)
}

// IsAdmin checks if a user is a platform admin
func (e *Enforcer) IsAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, authdomain.RoleAdmin.String(), authdomain.DomainPlatform)
}

// ================================================
// PERSONAL TEAM METHODS
// Domain: personal:team:{user_id}
// ================================================

// AddPersonalTeamRole adds a role for a user in their personal team
func (e *Enforcer) AddPersonalTeamRole(userID, role string) (bool, error) {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

// RemovePersonalTeamRole removes a role from a user's personal team
func (e *Enforcer) RemovePersonalTeamRole(userID, role string) (bool, error) {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

// GetUserPersonalTeamRoles returns all roles for a user in their personal team
func (e *Enforcer) GetUserPersonalTeamRoles(userID string) []string {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.GetRolesForUserInDomain(userID, domain)
}

// IsUserInPersonalTeam checks if a user has any role in their personal team
func (e *Enforcer) IsUserInPersonalTeam(userID string) bool {
	domain := authdomain.PersonalTeamDomain(userID)
	roles := e.GetRolesForUserInDomain(userID, domain)
	return len(roles) > 0
}

// HasPersonalTeamRole checks if a user has a specific role in their personal team
func (e *Enforcer) HasPersonalTeamRole(userID, role string) bool {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

// IsPersonalTeamAdmin checks if a user is an admin of their personal team
func (e *Enforcer) IsPersonalTeamAdmin(userID string) bool {
	return e.HasPersonalTeamRole(userID, authdomain.RoleAccountAdmin.String())
}

// ================================================
// INSTITUTION TEAM METHODS
// Domain: institution:team:{institution_id}
// ================================================

// AddInstitutionTeamRole adds a role for a user in an institution team
func (e *Enforcer) AddInstitutionTeamRole(userID, institutionID, role string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

// RemoveInstitutionTeamRole removes a role from a user in an institution team
func (e *Enforcer) RemoveInstitutionTeamRole(userID, institutionID, role string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

// RemoveAllInstitutionTeamRoles removes all roles for a user in an institution team
func (e *Enforcer) RemoveAllInstitutionTeamRoles(userID, institutionID string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
}

// GetUserInstitutionTeamRoles returns all roles for a user in an institution team
func (e *Enforcer) GetUserInstitutionTeamRoles(userID, institutionID string) []string {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.GetRolesForUserInDomain(userID, domain)
}

// HasInstitutionTeamRole checks if a user has a specific role in an institution team
func (e *Enforcer) HasInstitutionTeamRole(userID, institutionID, role string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

// IsInstitutionTeamAdmin checks if a user is an admin of an institution team
func (e *Enforcer) IsInstitutionTeamAdmin(userID, institutionID string) bool {
	return e.HasInstitutionTeamRole(userID, institutionID, authdomain.RoleAccountAdmin.String())
}

// IsInstitutionEventManager checks if a user is an event manager in an institution team
func (e *Enforcer) IsInstitutionEventManager(userID, institutionID string) bool {
	return e.HasInstitutionTeamRole(userID, institutionID, authdomain.RoleEventManager.String())
}

// IsInstitutionTeamMember checks if a user is a team member in an institution team
func (e *Enforcer) IsInstitutionTeamMember(userID, institutionID string) bool {
	return e.HasInstitutionTeamRole(userID, institutionID, authdomain.RoleTeamMember.String())
}

// ================================================
// USER INFORMATION METHODS
// ================================================

// HasAnyTeamRole checks if a user has any team-related role in any domain
func (e *Enforcer) HasAnyTeamRole(userID string) bool {
	domains := e.GetDomainsForUser(userID)

	for _, domain := range domains {
		if authdomain.IsTeamDomain(domain) {
			roles := e.GetRolesForUserInDomain(userID, domain)
			if len(roles) > 0 {
				return true
			}
		}
	}
	return false
}

// GetDomainsForUser returns all domains where a user has roles
func (e *Enforcer) GetDomainsForUser(userID string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, err := e.Enforcer.GetGroupingPolicy()
	if err != nil {
		log.Printf("Error getting grouping policy: %v", err)
		return []string{}
	}

	domains := make(map[string]bool)

	for _, policy := range policies {
		if len(policy) >= 3 && policy[0] == userID {
			domains[policy[2]] = true
		}
	}

	result := make([]string, 0, len(domains))
	for domain := range domains {
		result = append(result, domain)
	}
	return result
}

// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
func (e *Enforcer) GetUserPersonalTeamIDs(userID string) []string {
	domains := e.GetDomainsForUser(userID)
	teams := []string{}

	for _, domain := range domains {
		if authdomain.IsPersonalTeamDomain(domain) {
			teamID := authdomain.ExtractTeamID(domain)
			if teamID != "" {
				teams = append(teams, teamID)
			}
		}
	}
	return teams
}

// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
func (e *Enforcer) GetUserInstitutionTeamIDs(userID string) []string {
	domains := e.GetDomainsForUser(userID)
	teams := []string{}

	for _, domain := range domains {
		if authdomain.IsInstitutionTeamDomain(domain) {
			teamID := authdomain.ExtractTeamID(domain)
			if teamID != "" {
				teams = append(teams, teamID)
			}
		}
	}
	return teams
}

// GetUserTeamIDs returns all team IDs where a user has roles (both personal and institution)
func (e *Enforcer) GetUserTeamIDs(userID string) []string {
	teams := []string{}
	teams = append(teams, e.GetUserPersonalTeamIDs(userID)...)
	teams = append(teams, e.GetUserInstitutionTeamIDs(userID)...)
	return teams
}

// ================================================
// RESOURCE-SPECIFIC PERMISSION HELPERS
// ================================================

// CanManageTeam checks if user can manage a team (personal or institution)
func (e *Enforcer) CanManageTeam(userID, teamID string) bool {
	// Check personal team domain
	if e.HasPermission(userID, teamID, authdomain.ResourceTeam, authdomain.ActionManage) {
		return true
	}
	// Check institution team domain
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceTeam.String(), authdomain.ActionManage.String())
	if err == nil && allowed {
		return true
	}
	return false
}

// CanManageEvent checks if user can manage events in a team
func (e *Enforcer) CanManageEvent(userID, teamID string) bool {
	// Try personal team domain first
	if e.HasPermission(userID, teamID, authdomain.ResourceEvent, authdomain.ActionManage) {
		return true
	}
	// Try institution team domain
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String())
	if err == nil && allowed {
		return true
	}
	return false
}

// CanManageMember checks if user can manage members in a team
func (e *Enforcer) CanManageMember(userID, teamID string) bool {
	if e.HasPermission(userID, teamID, authdomain.ResourceMember, authdomain.ActionManage) {
		return true
	}
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceMember.String(), authdomain.ActionManage.String())
	if err == nil && allowed {
		return true
	}
	return false
}

// CanIssueCertificate checks if user can issue certificates in a team
func (e *Enforcer) CanIssueCertificate(userID, teamID string) bool {
	if e.HasPermission(userID, teamID, authdomain.ResourceCertificate, authdomain.ActionIssue) {
		return true
	}
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceCertificate.String(), authdomain.ActionIssue.String())
	if err == nil && allowed {
		return true
	}
	return false
}

// HasPermission checks if user has a specific permission in a personal team
func (e *Enforcer) HasPermission(userID, teamID string, resource authdomain.Resource, action authdomain.Action) bool {
	domain := authdomain.PersonalTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, resource.String(), action.String())
	if err != nil {
		return false
	}
	return allowed
}

// ================================================
// ONBOARDING HELPERS
// ================================================

// SetupPersonalTeam sets up a user's personal team with account_admin role
// Called when a new user is onboarded
// Domain: personal:team:{user_id}
func (e *Enforcer) SetupPersonalTeam(userID string) error {
	// Add user as account_admin of their personal team
	_, err := e.AddPersonalTeamRole(userID, authdomain.RoleAccountAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to setup personal team for user %s: %w", userID, err)
	}
	log.Printf("✅ Personal team setup for user: %s", userID)
	return nil
}

// SetupInstitutionTeam sets up an institution team with the admin as account_admin
// Called when a new institution is onboarded
// Domain: institution:team:{institution_id}
func (e *Enforcer) SetupInstitutionTeam(adminUserID, institutionID string) error {
	// Add admin as account_admin of the institution team
	_, err := e.AddInstitutionTeamRole(adminUserID, institutionID, authdomain.RoleAccountAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to setup institution team for admin %s in institution %s: %w", adminUserID, institutionID, err)
	}
	log.Printf("✅ Institution team setup for admin: %s in institution: %s", adminUserID, institutionID)
	return nil
}

// AddUserToInstitutionTeam adds a user to an institution team with a specific role
func (e *Enforcer) AddUserToInstitutionTeam(userID, institutionID, role string) error {
	if !authdomain.IsValidTeamRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}
	_, err := e.AddInstitutionTeamRole(userID, institutionID, role)
	if err != nil {
		return fmt.Errorf("failed to add user %s to institution %s with role %s: %w", userID, institutionID, role, err)
	}
	return nil
}

// ================================================
// HELPER METHODS
// ================================================

// GetRolesForUser gets all roles for a user (used by service)
func (e *Enforcer) GetRolesForUser(userID string, domain string) ([]string, error) {
	return e.GetRolesForUserInDomain(userID, domain), nil
}

// RemoveFilteredGroupingPolicy removes grouping policies matching filters
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}