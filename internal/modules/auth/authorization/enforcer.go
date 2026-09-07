// internal/modules/auth/authorization/enforcer.go

package authorization

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// ============================================================
// ENFORCER
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

// ============================================================
// PERMISSION CHECK METHODS
// ============================================================

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

// ============================================================
// POLICY MANAGEMENT METHODS
// ============================================================

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

// ============================================================
// ROLE MANAGEMENT METHODS
// ============================================================

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

// ============================================================
// PLATFORM ROLE METHODS
// ============================================================

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

// ============================================================
// ACCOUNT ROLE METHODS (ROLES ARE AT ACCOUNT LEVEL)
// Domain: account:{account_id}
// ============================================================

// AddAccountRole adds a role for a user in an account
// Roles: "account_admin" or "trainer"
func (e *Enforcer) AddAccountRole(userID, accountID, role string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	return e.AddRoleForUserInDomain(userID, role, domain)
}

// RemoveAccountRole removes a role from a user in an account
func (e *Enforcer) RemoveAccountRole(userID, accountID, role string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	return e.RemoveRoleForUserInDomain(userID, role, domain)
}

// RemoveAllAccountRoles removes all roles from a user in an account
func (e *Enforcer) RemoveAllAccountRoles(userID, accountID string) (bool, error) {
	domain := authdomain.AccountDomain(accountID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
}

// GetUserAccountRoles returns all roles for a user in an account
func (e *Enforcer) GetUserAccountRoles(userID, accountID string) []string {
	domain := authdomain.AccountDomain(accountID)
	return e.GetRolesForUserInDomain(userID, domain)
}

// HasAccountRole checks if a user has a specific role in an account
func (e *Enforcer) HasAccountRole(userID, accountID, role string) bool {
	domain := authdomain.AccountDomain(accountID)
	return e.HasRoleForUserInDomain(userID, role, domain)
}

// IsAccountAdmin checks if a user is an account admin in an account
func (e *Enforcer) IsAccountAdmin(userID, accountID string) bool {
	return e.HasAccountRole(userID, accountID, authdomain.RoleAccountAdmin.String())
}

// IsAccountTrainer checks if a user is a trainer in an account
func (e *Enforcer) IsAccountTrainer(userID, accountID string) bool {
	return e.HasAccountRole(userID, accountID, authdomain.RoleTrainer.String())
}

// GetUserRoleInAccount returns the user's role in a specific account
// Returns: "account_admin", "trainer", or empty string if not a member
func (e *Enforcer) GetUserRoleInAccount(userID, accountID string) string {
	domain := authdomain.AccountDomain(accountID)
	roles := e.GetRolesForUserInDomain(userID, domain)

	if len(roles) > 0 {
		return roles[0] // User has exactly one role per account
	}
	return ""
}

// IsAccountMember checks if user is a member (any role) of the account
func (e *Enforcer) IsAccountMember(userID, accountID string) bool {
	role := e.GetUserRoleInAccount(userID, accountID)
	return role != ""
}

// GetUserAccountRolesMap returns all roles for a user across all accounts
// Returns map[accountID]role
func (e *Enforcer) GetUserAccountRolesMap(userID string) map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	domains := e.GetDomainsForUser(userID)
	result := make(map[string]string)

	for _, domain := range domains {
		if strings.HasPrefix(domain, "account:") {
			accountID := strings.TrimPrefix(domain, "account:")
			if accountID != "" {
				roles := e.GetRolesForUserInDomain(userID, domain)
				if len(roles) > 0 {
					result[accountID] = roles[0]
				}
			}
		}
	}
	return result
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

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

// GetUserAccountDomains returns all account domains where a user has membership
// Returns domains in format: "account:{account_id}"
func (e *Enforcer) GetUserAccountDomains(userID string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	domains := e.GetDomainsForUser(userID)
	var accountDomains []string

	for _, domain := range domains {
		if strings.HasPrefix(domain, "account:") {
			accountDomains = append(accountDomains, domain)
		}
	}
	return accountDomains
}

// GetUserAccountIDs returns all account IDs where a user has membership
func (e *Enforcer) GetUserAccountIDs(userID string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	domains := e.GetDomainsForUser(userID)
	accountIDs := []string{}

	for _, domain := range domains {
		if strings.HasPrefix(domain, "account:") {
			accountID := strings.TrimPrefix(domain, "account:")
			if accountID != "" {
				accountIDs = append(accountIDs, accountID)
			}
		}
	}
	return accountIDs
}

func (e *Enforcer) HasAnyAccountRole(userID string) bool {
	domains := e.GetDomainsForUser(userID)

	for _, domain := range domains {
		if strings.HasPrefix(domain, "account:") {
			roles := e.GetRolesForUserInDomain(userID, domain)
			if len(roles) > 0 {
				return true
			}
		}
	}
	return false
}

// ============================================================
// ONBOARDING HELPERS
// ============================================================

func (e *Enforcer) SetupAccount(adminUserID, accountID string) error {
	_, err := e.AddAccountRole(adminUserID, accountID, authdomain.RoleAccountAdmin.String())
	if err != nil {
		return fmt.Errorf("failed to setup account %s for admin %s: %w", accountID, adminUserID, err)
	}
	log.Printf("✅ Account setup for admin: %s in account: %s", adminUserID, accountID)
	return nil
}

func (e *Enforcer) AddUserToAccount(userID, accountID, role string) error {
	if !authdomain.IsAccountRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}
	_, err := e.AddAccountRole(userID, accountID, role)
	if err != nil {
		return fmt.Errorf("failed to add user %s to account %s with role %s: %w", userID, accountID, role, err)
	}
	return nil
}

// ============================================================
// DEPRECATED - REMOVED
// ============================================================

// The following methods have been removed because roles are at account level, not team level:
// - AddPersonalTeamRole, RemovePersonalTeamRole, GetUserPersonalTeamRoles
// - AddInstitutionTeamRole, RemoveInstitutionTeamRole, GetUserInstitutionTeamRoles
// - SetupPersonalTeam, SetupInstitutionTeam
// - GetUserTeamIDs, GetUserTeamDomains
// - GetUserPersonalTeamIDs, GetUserInstitutionTeamIDs
//
// Use account-level methods instead:
// - AddAccountRole, RemoveAccountRole, GetUserAccountRoles
// - SetupAccount, AddUserToAccount
// - GetUserAccountIDs, GetUserAccountDomains