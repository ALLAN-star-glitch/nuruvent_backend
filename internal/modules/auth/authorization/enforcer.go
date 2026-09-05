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
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	e, err := casbin.NewEnforcer(cfg.Casbin.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	e.EnableAutoSave(true)

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

func (e *Enforcer) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.stopped {
		e.cancel()
		e.stopped = true
	}
}

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

func (e *Enforcer) Enforce(userID string, domain string, resource string, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.IsSuperAdmin(userID) {
		return true, nil
	}

	return e.Enforcer.Enforce(userID, domain, resource, action)
}

func (e *Enforcer) BatchEnforce(requests [][]interface{}) ([]bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.BatchEnforce(requests)
}

// ================================================
// POLICY MANAGEMENT METHODS
// ================================================

func (e *Enforcer) AddPolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddPolicy(sub, dom, obj, act)
}

func (e *Enforcer) RemovePolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemovePolicy(sub, dom, obj, act)
}

func (e *Enforcer) AddPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddPolicies(rules)
}

func (e *Enforcer) RemovePolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemovePolicies(rules)
}

func (e *Enforcer) AddGroupingPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddGroupingPolicies(rules)
}

func (e *Enforcer) RemoveGroupingPolicies(rules [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveGroupingPolicies(rules)
}

func (e *Enforcer) HasPolicy(sub, dom, obj, act string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.HasPolicy(sub, dom, obj, act)
}

func (e *Enforcer) HasGroupingPolicy(user, role, domain string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.HasGroupingPolicy(user, role, domain)
}

func (e *Enforcer) GetPolicy() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetPolicy()
}

func (e *Enforcer) GetGroupingPolicy() ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetGroupingPolicy()
}

func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}

func (e *Enforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// ================================================
// ROLE MANAGEMENT METHODS
// ================================================

func (e *Enforcer) AddRoleForUserInDomain(userID, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.AddGroupingPolicy(userID, role, domain)
}

func (e *Enforcer) RemoveRoleForUserInDomain(userID, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveGroupingPolicy(userID, role, domain)
}

func (e *Enforcer) GetRolesForUserInDomain(userID, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetRolesForUserInDomain(userID, domain)
}

func (e *Enforcer) GetImplicitRolesForUser(userID string, domain string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetImplicitRolesForUser(userID, domain)
}

func (e *Enforcer) GetImplicitPermissionsForUser(userID string, domain string) ([][]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Enforcer.GetImplicitPermissionsForUser(userID, domain)
}

func (e *Enforcer) HasRoleForUserInDomain(userID, role, domain string) bool {
	roles := e.GetRolesForUserInDomain(userID, domain)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

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

func (e *Enforcer) AddPlatformRole(userID string, role authdomain.Role) (bool, error) {
	return e.AddRoleForUserInDomain(userID, role.String(), authdomain.DomainPlatform)
}

func (e *Enforcer) RemovePlatformRole(userID string, role authdomain.Role) (bool, error) {
	return e.RemoveRoleForUserInDomain(userID, role.String(), authdomain.DomainPlatform)
}

func (e *Enforcer) GetUserPlatformRoles(userID string) []string {
	return e.GetRolesForUserInDomain(userID, authdomain.DomainPlatform)
}

func (e *Enforcer) HasPlatformRole(userID, role string) bool {
	return e.HasRoleForUserInDomain(userID, role, authdomain.DomainPlatform)
}

func (e *Enforcer) IsSuperAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, authdomain.RoleSuperAdmin.String(), authdomain.DomainPlatform)
}

func (e *Enforcer) IsAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, authdomain.RoleAdmin.String(), authdomain.DomainPlatform)
}

// ================================================
// PERSONAL TEAM METHODS
// Domain: personal:team:{user_id}
// ================================================

func (e *Enforcer) AddPersonalTeamRole(userID, role string) (bool, error) {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) RemovePersonalTeamRole(userID, role string) (bool, error) {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) GetUserPersonalTeamRoles(userID string) []string {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.GetRolesForUserInDomain(userID, domain)
}

func (e *Enforcer) IsUserInPersonalTeam(userID string) bool {
	domain := authdomain.PersonalTeamDomain(userID)
	roles := e.GetRolesForUserInDomain(userID, domain)
	return len(roles) > 0
}

func (e *Enforcer) HasPersonalTeamRole(userID, role string) bool {
	domain := authdomain.PersonalTeamDomain(userID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) IsPersonalTeamAdmin(userID string) bool {
	return e.HasPersonalTeamRole(userID, authdomain.RoleAccountAdmin.String())
}

// ================================================
// INSTITUTION TEAM METHODS
// Domain: institution:team:{institution_id}
// ================================================

func (e *Enforcer) AddInstitutionTeamRole(userID, institutionID, role string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) RemoveInstitutionTeamRole(userID, institutionID, role string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) RemoveAllInstitutionTeamRoles(userID, institutionID string) (bool, error) {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
}

func (e *Enforcer) GetUserInstitutionTeamRoles(userID, institutionID string) []string {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.GetRolesForUserInDomain(userID, domain)
}

func (e *Enforcer) HasInstitutionTeamRole(userID, institutionID, role string) bool {
	domain := authdomain.InstitutionTeamDomain(institutionID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) IsInstitutionTeamAdmin(userID, institutionID string) bool {
	return e.HasInstitutionTeamRole(userID, institutionID, authdomain.RoleAccountAdmin.String())
}

// ================================================
// ACCOUNT METHODS (NEW)
// Domain: account:{account_id}
// ================================================

func (e *Enforcer) AddAccountRole(userID, accountID, role string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) RemoveAccountRole(userID, accountID, role string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) RemoveAllAccountRoles(userID, accountID string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
}

func (e *Enforcer) GetUserAccountRoles(userID, accountID string) []string {
	domain := authdomain.AccountDomain(accountID)
	return e.GetRolesForUserInDomain(userID, domain)
}

func (e *Enforcer) HasAccountRole(userID, accountID, role string) bool {
	domain := authdomain.AccountDomain(accountID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

func (e *Enforcer) IsAccountAdmin(userID, accountID string) bool {
	return e.HasAccountRole(userID, accountID, authdomain.RoleAccountAdmin.String())
}

func (e *Enforcer) IsAccountTrainer(userID, accountID string) bool {
	return e.HasAccountRole(userID, accountID, authdomain.RoleTrainer.String())
}

// ================================================
// USER INFORMATION METHODS
// ================================================

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

func (e *Enforcer) GetUserTeamIDs(userID string) []string {
	teams := []string{}
	teams = append(teams, e.GetUserPersonalTeamIDs(userID)...)
	teams = append(teams, e.GetUserInstitutionTeamIDs(userID)...)
	return teams
}

// ================================================
// RESOURCE-SPECIFIC PERMISSION HELPERS
// ================================================

func (e *Enforcer) CanManageTeam(userID, teamID string) bool {
	if e.HasPermission(userID, teamID, authdomain.ResourceTeam, authdomain.ActionManage) {
		return true
	}
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceTeam.String(), authdomain.ActionManage.String())
	if err == nil && allowed {
		return true
	}
	return false
}

func (e *Enforcer) CanManageEvent(userID, teamID string) bool {
	if e.HasPermission(userID, teamID, authdomain.ResourceEvent, authdomain.ActionManage) {
		return true
	}
	domain := authdomain.InstitutionTeamDomain(teamID)
	allowed, err := e.Enforce(userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionManage.String())
	if err == nil && allowed {
		return true
	}
	return false
}

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

func (e *Enforcer) SetupPersonalTeam(userID string) error {
	_, err := e.AddPersonalTeamRole(userID, authdomain.RoleAccountAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to setup personal team for user %s: %w", userID, err)
	}
	log.Printf("✅ Personal team setup for user: %s", userID)
	return nil
}

func (e *Enforcer) SetupInstitutionTeam(adminUserID, institutionID string) error {
	_, err := e.AddInstitutionTeamRole(adminUserID, institutionID, authdomain.RoleAccountAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to setup institution team for admin %s in institution %s: %w", adminUserID, institutionID, err)
	}
	log.Printf("✅ Institution team setup for admin: %s in institution: %s", adminUserID, institutionID)
	return nil
}

func (e *Enforcer) AddUserToInstitutionTeam(userID, institutionID, role string) error {
	if !authdomain.IsAccountRole(role) {
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

func (e *Enforcer) GetRolesForUser(userID string, domain string) ([]string, error) {
	return e.GetRolesForUserInDomain(userID, domain), nil
}

func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}