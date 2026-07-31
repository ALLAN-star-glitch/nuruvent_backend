package authorization

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var (
	instance *Enforcer
	once     sync.Once
	initErr  error
)

// Enforcer wraps the Casbin enforcer with additional functionality - singleton pattern, auto-reload, and thread safety
type Enforcer struct {
	*casbin.Enforcer
	mu      sync.RWMutex
	cfg     *config.CasbinConfig
	db      *gorm.DB
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

// InitEnforcer initializes the Casbin enforcer (singleton)
func InitEnforcer(db *gorm.DB, cfg *config.Config) (*Enforcer, error) {
	once.Do(func() {
		instance, initErr = newEnforcer(db, cfg)
	})

	if initErr != nil {
		return nil, initErr
	}
	return instance, nil
}

func newEnforcer(db *gorm.DB, cfg *config.Config) (*Enforcer, error) {
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

	// Load policies from database
	err = e.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
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

	// Start auto-reload if enabled
	if cfg.Casbin.AutoLoad {
		go enforcer.autoLoadPolicies()
	}

	return enforcer, nil
}

// GetEnforcer returns the singleton enforcer instance
func GetEnforcer() *Enforcer {
	if instance == nil {
		panic("Casbin enforcer not initialized")
	}
	return instance
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
	return e.Enforcer.Enforce(userID, domain, resource, action)
}

// EnforceWithContext checks permission with explicit resource and action strings
func (e *Enforcer) EnforceWithContext(userID string, domain string, resource string, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
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
func (e *Enforcer) AddPlatformRole(userID string, role Role) (bool, error) {
	return e.AddRoleForUserInDomain(userID, role.String(), DomainPlatform)
}

// RemovePlatformRole removes a platform-level role from a user
func (e *Enforcer) RemovePlatformRole(userID string, role Role) (bool, error) {
	return e.RemoveRoleForUserInDomain(userID, role.String(), DomainPlatform)
}

// GetUserPlatformRoles returns all platform-level roles for a user
func (e *Enforcer) GetUserPlatformRoles(userID string) []string {
	return e.GetRolesForUserInDomain(userID, DomainPlatform)
}

// HasPlatformRole checks if a user has a specific platform role
func (e *Enforcer) HasPlatformRole(userID, role string) bool {
	return e.HasRoleForUserInDomain(userID, role, DomainPlatform)
}

// IsAdmin checks if a user is a platform admin
func (e *Enforcer) IsAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, RoleAdmin.String(), DomainPlatform)
}

// IsSuperAdmin checks if a user is a super admin
func (e *Enforcer) IsSuperAdmin(userID string) bool {
	return e.HasRoleForUserInDomain(userID, RoleSuperAdmin.String(), DomainPlatform)
}

// IsAttendee checks if a user has the attendee role at platform level
func (e *Enforcer) IsAttendee(userID string) bool {
	return e.HasPlatformRole(userID, RoleAttendee.String())
}

// IsPremiumAttendee checks if a user has the premium attendee role at platform level
func (e *Enforcer) IsPremiumAttendee(userID string) bool {
	return e.HasPlatformRole(userID, RolePremiumAttendee.String())
}

// ================================================
// BUSINESS ACCESS CHECK METHODS
// ================================================

// GetDomainsForUser returns all domains where a user has roles
func (e *Enforcer) GetDomainsForUser(userID string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Get all grouping policies (role assignments)
	policies, err := e.Enforcer.GetGroupingPolicy()
	if err != nil {
		log.Printf("Error getting grouping policy: %v", err)
		return []string{}
	}

	domains := make(map[string]bool)

	for _, policy := range policies {
		// policy format: [user, role, domain]
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

// HasAnyBusinessRole checks if a user has any business-related role in any domain
func (e *Enforcer) HasAnyBusinessRole(userID string) bool {
	domains := e.GetDomainsForUser(userID)

	for _, domain := range domains {
		if IsBusinessDomain(domain) {
			roles := e.GetRolesForUserInDomain(userID, domain)
			if len(roles) > 0 {
				return true
			}
		}
	}
	return false
}

// HasAnyBusinessRoleInBusiness checks if a user has any business role in a specific business
func (e *Enforcer) HasAnyBusinessRoleInBusiness(userID string, businessID string) bool {
	domain := BusinessDomain(businessID)
	roles := e.GetRolesForUserInDomain(userID, domain)
	return len(roles) > 0
}

// GetUserBusinesses returns all business IDs where a user has roles
func (e *Enforcer) GetUserBusinesses(userID string) []string {
	domains := e.GetDomainsForUser(userID)
	businesses := []string{}

	for _, domain := range domains {
		if IsBusinessDomain(domain) {
			businessID := ExtractBusinessID(domain)
			if businessID != "" {
				businesses = append(businesses, businessID)
			}
		}
	}
	return businesses
}

// GetUserBusinessRolesInAllBusinesses returns all business roles for a user across all businesses
func (e *Enforcer) GetUserBusinessRolesInAllBusinesses(userID string) map[string][]string {
	domains := e.GetDomainsForUser(userID)
	result := make(map[string][]string)

	for _, domain := range domains {
		if IsBusinessDomain(domain) {
			businessID := ExtractBusinessID(domain)
			if businessID != "" {
				roles := e.GetRolesForUserInDomain(userID, domain)
				if len(roles) > 0 {
					result[businessID] = roles
				}
			}
		}
	}
	return result
}

// GetUserBusinessIDsWithRole returns business IDs where a user has a specific role
func (e *Enforcer) GetUserBusinessIDsWithRole(userID string, role string) []string {
	domains := e.GetDomainsForUser(userID)
	businesses := []string{}

	for _, domain := range domains {
		if IsBusinessDomain(domain) && e.HasRoleForUserInDomain(userID, role, domain) {
			businessID := ExtractBusinessID(domain)
			if businessID != "" {
				businesses = append(businesses, businessID)
			}
		}
	}
	return businesses
}

// GetUserBusinessIDsWithBusinessAdminRole returns business IDs where a user has business_admin role
func (e *Enforcer) GetUserBusinessIDsWithBusinessAdminRole(userID string) []string {
	return e.GetUserBusinessIDsWithRole(userID, RoleBusinessAdmin.String())
}

// ================================================
// BUSINESS ROLE CHECK METHODS
// ================================================

// IsBusinessAdmin checks if a user is a business admin for a specific business
func (e *Enforcer) IsBusinessAdmin(userID string, businessID string) bool {
	domain := BusinessDomain(businessID)
	return e.HasRoleForUserInDomain(userID, RoleBusinessAdmin.String(), domain)
}

// IsEventManager checks if a user is an event manager for a specific business
func (e *Enforcer) IsEventManager(userID string, businessID string) bool {
	domain := BusinessDomain(businessID)
	return e.HasRoleForUserInDomain(userID, RoleEventManager.String(), domain)
}

// IsMember checks if a user is a member for a specific business
func (e *Enforcer) IsMember(userID string, businessID string) bool {
	domain := BusinessDomain(businessID)
	return e.HasRoleForUserInDomain(userID, RoleMember.String(), domain)
}

// ================================================
// RESOURCE-SPECIFIC PERMISSION HELPERS
// ================================================

// CanManageBusiness checks if user can manage a business
func (e *Enforcer) CanManageBusiness(userID, businessID string) bool {
	allowed, err := e.Enforce(userID, BusinessDomain(businessID), ResourceBusiness.String(), ActionManage.String())
	if err != nil {
		return false
	}
	return allowed
}

// CanManageEvent checks if user can manage events
func (e *Enforcer) CanManageEvent(userID, businessID string) bool {
	allowed, err := e.Enforce(userID, BusinessDomain(businessID), ResourceEvent.String(), ActionManage.String())
	if err != nil {
		return false
	}
	return allowed
}

// CanIssueCertificate checks if user can issue certificates
func (e *Enforcer) CanIssueCertificate(userID, businessID string) bool {
	allowed, err := e.Enforce(userID, BusinessDomain(businessID), ResourceCertificate.String(), ActionIssue.String())
	if err != nil {
		return false
	}
	return allowed
}

// HasPermission checks if user has a specific permission
func (e *Enforcer) HasPermission(userID, businessID string, resource Resource, action Action) bool {
	allowed, err := e.Enforce(userID, BusinessDomain(businessID), resource.String(), action.String())
	if err != nil {
		return false
	}
	return allowed
}