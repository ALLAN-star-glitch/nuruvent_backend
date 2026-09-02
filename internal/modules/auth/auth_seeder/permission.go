// internal/modules/auth/auth_seeder/permission.go

package authseeder

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"gorm.io/gorm"
)

// ============================================================
// CONSTANTS
// ============================================================

// Current policy schema version - increment when policies change
const CURRENT_POLICY_VERSION = "v4"

// ============================================================
// POLICY VERSION TRACKING
// ============================================================

// PolicyVersion tracks which policy version is applied to the database
type PolicyVersion struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Version     string    `gorm:"type:varchar(20);uniqueIndex"`
	AppliedAt   time.Time `gorm:"autoCreateTime"`
	Description string    `gorm:"type:text"`
}

func (PolicyVersion) TableName() string {
	return "policy_versions"
}

// ============================================================
// PUBLIC ENTRY FUNCTION
// ============================================================

// SeedPermissions seeds or migrates platform permissions
// This is idempotent - safe to run multiple times
func SeedPermissions(db *gorm.DB) error {
	log.Println("🌱 Seeding platform permissions...")

	// Create policy_versions table if not exists
	if err := db.AutoMigrate(&PolicyVersion{}); err != nil {
		return fmt.Errorf("failed to create policy_versions table: %w", err)
	}

	// Check if current version already applied
	var existing PolicyVersion
	err := db.Where("version = ?", CURRENT_POLICY_VERSION).First(&existing).Error
	if err == nil {
		log.Printf("✅ Platform policies already up to date (version %s)", CURRENT_POLICY_VERSION)
		log.Printf("   Applied at: %s", existing.AppliedAt.Format(time.RFC3339))
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check policy version: %w", err)
	}

	cfg := config.Load()

	// Initialize enforcer
	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	policyManager := authorization.NewPolicyManager(enforcer)
	roleManager := authorization.NewRoleManager(enforcer)

	seeder := &permissionSeeder{
		db:            db,
		enforcer:      enforcer,
		policyManager: policyManager,
		roleManager:   roleManager,
	}

	// Check if we have old policies and need to migrate
	var oldVersion PolicyVersion
	err = db.Where("version != ?", CURRENT_POLICY_VERSION).First(&oldVersion).Error
	if err == nil {
		log.Printf("⚠️  Detected old policy version: %s", oldVersion.Version)
		log.Printf("🔄 Running migration to update policies...")

		if err := seeder.migratePolicies(); err != nil {
			return fmt.Errorf("failed to migrate policies: %w", err)
		}
	} else if err == gorm.ErrRecordNotFound {
		// Fresh seed - no previous versions found
		log.Println("📦 No previous policy versions found. Performing fresh seed...")

		if err := seeder.freshSeed(); err != nil {
			return fmt.Errorf("failed to seed policies: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check policy versions: %w", err)
	}

	// Record the new version
	version := PolicyVersion{
		Version:     CURRENT_POLICY_VERSION,
		Description: fmt.Sprintf("Policy schema version %s", CURRENT_POLICY_VERSION),
	}
	if err := db.Create(&version).Error; err != nil {
		log.Printf("⚠️  Failed to record policy version: %v", err)
	}

	log.Printf("✅ Platform permissions successfully updated to version %s", CURRENT_POLICY_VERSION)
	return nil
}

// ============================================================
// INTERNAL SEEDER
// ============================================================

type permissionSeeder struct {
	db            *gorm.DB
	enforcer      *authorization.Enforcer
	policyManager authdomain.PolicyManager
	roleManager   authdomain.RoleManager
}

// ============================================================
// CLEANUP METHODS
// ============================================================

// cleanAllPolicies removes ALL policies from the database using direct SQL
func (s *permissionSeeder) cleanAllPolicies() error {
	log.Println("🧹 Cleaning all existing policies...")

	// 1. Clean all platform policies
	if err := s.cleanPlatformPolicies(); err != nil {
		return err
	}

	// 2. Clean all team policies (personal and institution)
	if err := s.cleanTeamPolicies(); err != nil {
		return err
	}

	// 3. Clean all grouping policies (role assignments)
	if err := s.cleanGroupingPolicies(); err != nil {
		return err
	}

	log.Println("   ✅ All policies cleaned")
	return nil
}

// cleanPlatformPolicies removes all platform policies using direct SQL
func (s *permissionSeeder) cleanPlatformPolicies() error {
	// Get all platform policies
	platformPolicies, err := s.enforcer.GetFilteredPolicy(1, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get platform policies: %w", err)
	}
	if len(platformPolicies) > 0 {
		if _, err := s.enforcer.RemovePolicies(platformPolicies); err != nil {
			return fmt.Errorf("failed to remove platform policies: %w", err)
		}
		log.Printf("   ✅ Removed %d platform policies", len(platformPolicies))
	}

	// Get all platform grouping policies
	platformGrouping, err := s.enforcer.GetFilteredGroupingPolicy(2, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get platform grouping policies: %w", err)
	}
	if len(platformGrouping) > 0 {
		if _, err := s.enforcer.RemoveGroupingPolicies(platformGrouping); err != nil {
			return fmt.Errorf("failed to remove platform grouping policies: %w", err)
		}
		log.Printf("   ✅ Removed %d platform grouping policies", len(platformGrouping))
	}

	return nil
}

// cleanTeamPolicies removes all team policies using direct SQL
func (s *permissionSeeder) cleanTeamPolicies() error {
	// Get all policies
	allPolicies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all policies: %w", err)
	}

	var teamPolicies [][]string
	for _, policy := range allPolicies {
		if len(policy) >= 4 {
			domain := policy[1]
			// Check if it's a team domain
			if authdomain.IsPersonalTeamDomain(domain) || authdomain.IsInstitutionTeamDomain(domain) {
				teamPolicies = append(teamPolicies, policy)
			}
		}
	}

	if len(teamPolicies) > 0 {
		if _, err := s.enforcer.RemovePolicies(teamPolicies); err != nil {
			return fmt.Errorf("failed to remove team policies: %w", err)
		}
		log.Printf("   ✅ Removed %d team policies", len(teamPolicies))
	}

	return nil
}

// cleanGroupingPolicies removes all grouping policies (role assignments) using direct SQL
func (s *permissionSeeder) cleanGroupingPolicies() error {
	// Get all grouping policies
	allGrouping, err := s.enforcer.GetGroupingPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all grouping policies: %w", err)
	}

	var teamGrouping [][]string
	for _, group := range allGrouping {
		if len(group) >= 3 {
			domain := group[2]
			// Check if it's a team domain
			if authdomain.IsPersonalTeamDomain(domain) || authdomain.IsInstitutionTeamDomain(domain) {
				teamGrouping = append(teamGrouping, group)
			}
		}
	}

	if len(teamGrouping) > 0 {
		if _, err := s.enforcer.RemoveGroupingPolicies(teamGrouping); err != nil {
			return fmt.Errorf("failed to remove team grouping policies: %w", err)
		}
		log.Printf("   ✅ Removed %d team grouping policies", len(teamGrouping))
	}

	return nil
}

// ============================================================
// SEED METHODS
// ============================================================

// freshSeed performs a fresh seed (no existing policies)
func (s *permissionSeeder) freshSeed() error {
	log.Println("🌱 Performing fresh seed...")

	// Clean all existing policies first
	if err := s.cleanAllPolicies(); err != nil {
		return err
	}

	// Seed platform policies
	if err := s.seedPlatformPolicies(); err != nil {
		return err
	}

	// Seed platform role hierarchy
	if err := s.seedPlatformRoleHierarchy(); err != nil {
		return err
	}

	log.Println("✅ Fresh seed completed")
	return nil
}

// migratePolicies updates existing policies to the new schema
func (s *permissionSeeder) migratePolicies() error {
	log.Println("🔄 Migrating policies to new schema...")

	// 1. Clean all existing policies first (to avoid duplicates)
	if err := s.cleanAllPolicies(); err != nil {
		return err
	}

	// 2. Seed platform policies
	if err := s.seedPlatformPolicies(); err != nil {
		return err
	}

	// 3. Seed platform role hierarchy
	if err := s.seedPlatformRoleHierarchy(); err != nil {
		return err
	}

	// 4. Update team policies for existing users and institutions
	if err := s.updateTeamPolicies(); err != nil {
		return err
	}

	log.Println("✅ Policy migration completed successfully")
	return nil
}

// seedPlatformPolicies seeds platform policies from policies.go
func (s *permissionSeeder) seedPlatformPolicies() error {
	policies := authorization.GetPlatformPolicies()
	if _, err := s.enforcer.AddPolicies(policies); err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}
	log.Printf("   ✅ Seeded %d platform policies", len(policies))
	return nil
}

// seedPlatformRoleHierarchy seeds platform role hierarchy
func (s *permissionSeeder) seedPlatformRoleHierarchy() error {
	hierarchy := authorization.GetPlatformRoleHierarchy()
	if _, err := s.enforcer.AddGroupingPolicies(hierarchy); err != nil {
		return fmt.Errorf("failed to add platform role hierarchy: %w", err)
	}
	log.Printf("   ✅ Seeded %d platform role hierarchy entries", len(hierarchy))
	return nil
}

// ============================================================
// TEAM POLICY UPDATE METHODS
// ============================================================

// updateTeamPolicies updates team policies for all existing teams
func (s *permissionSeeder) updateTeamPolicies() error {
	log.Println("📝 Updating team policies for existing users and institutions...")

	// 1. Update personal team policies for all users
	if err := s.updatePersonalTeamPolicies(); err != nil {
		return err
	}

	// 2. Update institution team policies for all institutions
	if err := s.updateInstitutionTeamPolicies(); err != nil {
		return err
	}

	return nil
}

// updatePersonalTeamPolicies updates personal team policies for all users
func (s *permissionSeeder) updatePersonalTeamPolicies() error {
	var userIDs []string
	if err := s.db.Table("users").Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("failed to get user IDs: %w", err)
	}

	if len(userIDs) == 0 {
		log.Println("   No users found to update personal team policies")
		return nil
	}

	userUpdated := 0
	for _, userID := range userIDs {
		domain := authdomain.PersonalTeamDomain(userID)

		// Add new policies
		newPolicies := authorization.GetPersonalTeamPolicies(domain)
		if _, err := s.enforcer.AddPolicies(newPolicies); err != nil {
			log.Printf("   ⚠️  Failed to add policies for user %s: %v", userID, err)
			continue
		}

		// Ensure user has account_admin role
		hasRole := s.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain)
		if !hasRole {
			if _, err := s.enforcer.AddRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain); err != nil {
				log.Printf("   ⚠️  Failed to assign account_admin role for user %s: %v", userID, err)
			}
		}

		userUpdated++
	}

	log.Printf("   ✅ Updated personal team policies for %d users", userUpdated)
	return nil
}

// updateInstitutionTeamPolicies updates institution team policies for all institutions
func (s *permissionSeeder) updateInstitutionTeamPolicies() error {
	var institutionIDs []string
	if err := s.db.Table("institutions").Pluck("id", &institutionIDs).Error; err != nil {
		return fmt.Errorf("failed to get institution IDs: %w", err)
	}

	if len(institutionIDs) == 0 {
		log.Println("   No institutions found to update institution team policies")
		return nil
	}

	instUpdated := 0
	for _, instID := range institutionIDs {
		domain := authdomain.InstitutionTeamDomain(instID)

		newPolicies := authorization.GetInstitutionTeamPolicies(domain)
		if _, err := s.enforcer.AddPolicies(newPolicies); err != nil {
			log.Printf("   ⚠️  Failed to add policies for institution %s: %v", instID, err)
			continue
		}
		instUpdated++
	}

	log.Printf("   ✅ Updated institution team policies for %d institutions", instUpdated)
	return nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// IsSeeded checks if policies have been seeded
func IsSeeded(db *gorm.DB) (bool, error) {
	var count int64
	if err := db.Model(&PolicyVersion{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check policy versions: %w", err)
	}
	return count > 0, nil
}

// GetCurrentVersion returns the current policy version
func GetCurrentVersion(db *gorm.DB) (string, error) {
	var version PolicyVersion
	err := db.Order("applied_at DESC").First(&version).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get current version: %w", err)
	}
	return version.Version, nil
}

// ============================================================
// TEAM POLICY SEEDING (for individual teams)
// ============================================================

// SeedPersonalTeamPolicies adds policies for a personal team (user's own team)
// Domain: personal:team:{user_id}
func SeedPersonalTeamPolicies(db *gorm.DB, userID string) error {
	log.Printf("📦 Adding personal team policies for user: %s", userID)

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()

	policyManager := authorization.NewPolicyManager(enforcer)
	roleManager := authorization.NewRoleManager(enforcer)

	scope := authdomain.NewPersonalTeamScope(userID)

	// Add policies
	if err := policyManager.AddTeamPolicies(ctx, scope); err != nil {
		return fmt.Errorf("failed to add personal team policies: %w", err)
	}

	// Assign account admin role
	if err := roleManager.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String()); err != nil {
		return fmt.Errorf("failed to assign account admin role: %w", err)
	}

	log.Printf("✅ Personal team policies seeded for user: %s", userID)
	return nil
}

// SeedInstitutionTeamPolicies adds policies for an institution team
// Domain: institution:team:{institution_id}
func SeedInstitutionTeamPolicies(db *gorm.DB, institutionID string) error {
	log.Printf("📦 Adding institution team policies for institution: %s", institutionID)

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()

	policyManager := authorization.NewPolicyManager(enforcer)

	scope := authdomain.NewInstitutionTeamScope(institutionID)

	if err := policyManager.AddTeamPolicies(ctx, scope); err != nil {
		return fmt.Errorf("failed to add institution team policies: %w", err)
	}

	log.Printf("✅ Institution team policies seeded for institution: %s", institutionID)
	return nil
}

// ============================================================
// BULK SEEDING FUNCTIONS (for migrations)
// ============================================================

// SeedPersonalTeamPoliciesForAllUsers adds personal team policies for all users
func SeedPersonalTeamPoliciesForAllUsers(db *gorm.DB) error {
	log.Println("📦 Adding personal team policies for all users...")

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()

	policyManager := authorization.NewPolicyManager(enforcer)
	roleManager := authorization.NewRoleManager(enforcer)

	var userIDs []string
	if err := db.Table("users").Pluck("id", &userIDs).Error; err != nil {
		return fmt.Errorf("failed to get user IDs: %w", err)
	}

	if len(userIDs) == 0 {
		log.Println("No users found to seed personal team policies for")
		return nil
	}

	successCount := 0
	for _, userID := range userIDs {
		scope := authdomain.NewPersonalTeamScope(userID)

		roles, err := roleManager.GetUserRoles(ctx, userID, scope)
		if err != nil {
			log.Printf("⚠️  Failed to get roles for user %s: %v", userID, err)
			continue
		}

		if len(roles) > 0 {
			continue
		}

		if err := policyManager.AddTeamPolicies(ctx, scope); err != nil {
			log.Printf("⚠️  Failed to seed personal team policies for user %s: %v", userID, err)
			continue
		}

		if err := roleManager.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String()); err != nil {
			log.Printf("⚠️  Failed to assign account admin role for user %s: %v", userID, err)
			continue
		}

		successCount++
	}

	log.Printf("✅ Seeded personal team policies for %d out of %d users", successCount, len(userIDs))
	return nil
}

// SeedInstitutionTeamPoliciesForAllInstitutions adds institution team policies for all institutions
func SeedInstitutionTeamPoliciesForAllInstitutions(db *gorm.DB) error {
	log.Println("📦 Adding institution team policies for all institutions...")

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()

	policyManager := authorization.NewPolicyManager(enforcer)

	var institutionIDs []string
	if err := db.Table("institutions").Pluck("id", &institutionIDs).Error; err != nil {
		return fmt.Errorf("failed to get institution IDs: %w", err)
	}

	if len(institutionIDs) == 0 {
		log.Println("No institutions found to seed institution team policies for")
		return nil
	}

	successCount := 0
	for _, institutionID := range institutionIDs {
		scope := authdomain.NewInstitutionTeamScope(institutionID)

		policies, err := enforcer.GetFilteredPolicy(1, scope.Domain())
		if err != nil {
			log.Printf("⚠️  Failed to check policies for institution %s: %v", institutionID, err)
			continue
		}

		if len(policies) > 0 {
			continue
		}

		if err := policyManager.AddTeamPolicies(ctx, scope); err != nil {
			log.Printf("⚠️  Failed to seed institution team policies for institution %s: %v", institutionID, err)
			continue
		}

		successCount++
	}

	log.Printf("✅ Seeded institution team policies for %d out of %d institutions", successCount, len(institutionIDs))
	return nil
}

// ============================================================
// ASSIGN ADMIN ROLE TO INSTITUTION
// ============================================================

// AssignInstitutionAdmin assigns account_admin role to a user in an institution
func AssignInstitutionAdmin(db *gorm.DB, institutionID, userID string) error {
	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()
	roleManager := authorization.NewRoleManager(enforcer)

	scope := authdomain.NewInstitutionTeamScope(institutionID)

	if err := roleManager.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String()); err != nil {
		return fmt.Errorf("failed to assign institution admin role: %w", err)
	}

	log.Printf("✅ Assigned account_admin role for institution %s to user %s", institutionID, userID)
	return nil
}