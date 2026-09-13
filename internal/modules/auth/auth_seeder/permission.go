// internal/modules/auth/auth_seeder/permission.go

package authseeder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"gorm.io/gorm"
)

// ============================================================
// CONSTANTS
// ============================================================

const CURRENT_POLICY_VERSION = "v15"

// ============================================================
// POLICY VERSION TRACKING
// ============================================================

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

func SeedPermissions(db *gorm.DB) error {
	log.Println("🌱 Seeding platform permissions...")

	if err := db.AutoMigrate(&PolicyVersion{}); err != nil {
		return fmt.Errorf("failed to create policy_versions table: %w", err)
	}

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

	var oldVersion PolicyVersion
	err = db.Where("version != ?", CURRENT_POLICY_VERSION).First(&oldVersion).Error
	if err == nil {
		log.Printf("⚠️  Detected old policy version: %s", oldVersion.Version)
		log.Printf("🔄 Running migration to update policies...")

		if err := seeder.migratePolicies(); err != nil {
			return fmt.Errorf("failed to migrate policies: %w", err)
		}
	} else if err == gorm.ErrRecordNotFound {
		log.Println("📦 No previous policy versions found. Performing fresh seed...")

		if err := seeder.freshSeed(); err != nil {
			return fmt.Errorf("failed to seed policies: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check policy versions: %w", err)
	}

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

func (s *permissionSeeder) cleanAllPolicies() error {
	log.Println("🧹 Cleaning all existing policies...")

	if err := s.cleanPlatformPolicies(); err != nil {
		return err
	}
	if err := s.cleanAccountPolicies(); err != nil {
		return err
	}
	if err := s.cleanTeamDomainPolicies(); err != nil {
		return err
	}

	log.Println("   ✅ All policies cleaned")
	return nil
}

func (s *permissionSeeder) cleanPlatformPolicies() error {
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

func (s *permissionSeeder) cleanAccountPolicies() error {
	allPolicies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all policies: %w", err)
	}

	var accountPolicies [][]string
	for _, policy := range allPolicies {
		if len(policy) >= 2 {
			domain := policy[1]
			if strings.HasPrefix(domain, "account:") {
				accountPolicies = append(accountPolicies, policy)
			}
		}
	}

	if len(accountPolicies) > 0 {
		if _, err := s.enforcer.RemovePolicies(accountPolicies); err != nil {
			return fmt.Errorf("failed to remove account policies: %w", err)
		}
		log.Printf("   ✅ Removed %d account policies", len(accountPolicies))
	}

	return nil
}

func (s *permissionSeeder) cleanTeamDomainPolicies() error {
	allPolicies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all policies: %w", err)
	}

	var teamPolicies [][]string
	for _, policy := range allPolicies {
		if len(policy) >= 2 {
			domain := policy[1]
			if strings.HasPrefix(domain, "personal:team:") ||
				strings.HasPrefix(domain, "institution:team:") {
				teamPolicies = append(teamPolicies, policy)
			}
		}
	}

	if len(teamPolicies) > 0 {
		if _, err := s.enforcer.RemovePolicies(teamPolicies); err != nil {
			return fmt.Errorf("failed to remove team-domain policies: %w", err)
		}
		log.Printf("   ✅ Removed %d team-domain policies", len(teamPolicies))
	}

	allGrouping, err := s.enforcer.GetGroupingPolicy()
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	var teamGrouping [][]string
	for _, rule := range allGrouping {
		if len(rule) >= 3 {
			domain := rule[2]
			if strings.HasPrefix(domain, "personal:team:") ||
				strings.HasPrefix(domain, "institution:team:") {
				teamGrouping = append(teamGrouping, rule)
			}
		}
	}

	if len(teamGrouping) > 0 {
		if _, err := s.enforcer.RemoveGroupingPolicies(teamGrouping); err != nil {
			return fmt.Errorf("failed to remove team-domain grouping: %w", err)
		}
		log.Printf("   ✅ Removed %d team-domain grouping rules", len(teamGrouping))
	}

	return nil
}

// ============================================================
// SEED METHODS
// ============================================================

func (s *permissionSeeder) freshSeed() error {
	log.Println("🌱 Performing fresh seed...")

	if err := s.cleanAllPolicies(); err != nil {
		return err
	}
	if err := s.seedPlatformPolicies(); err != nil {
		return err
	}
	if err := s.seedPlatformRoleHierarchy(); err != nil {
		return err
	}
	if err := s.seedAccountPolicies(); err != nil {
		return err
	}
	if err := s.seedAccountRoleHierarchy(); err != nil {
		return err
	}

	log.Println("✅ Fresh seed completed")
	return nil
}

func (s *permissionSeeder) migratePolicies() error {
	log.Println("🔄 Migrating policies to account-only model...")

	if err := s.cleanAllPolicies(); err != nil {
		return err
	}
	if err := s.seedPlatformPolicies(); err != nil {
		return err
	}
	if err := s.seedPlatformRoleHierarchy(); err != nil {
		return err
	}
	if err := s.seedAccountPolicies(); err != nil {
		return err
	}
	if err := s.seedAccountRoleHierarchy(); err != nil {
		return err
	}

	log.Println("✅ Policy migration completed successfully")
	return nil
}

func (s *permissionSeeder) seedPlatformPolicies() error {
	policies := authorization.GetPlatformPolicies()
	if _, err := s.enforcer.AddPolicies(policies); err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}
	log.Printf("   ✅ Seeded %d platform policies", len(policies))
	return nil
}

func (s *permissionSeeder) seedPlatformRoleHierarchy() error {
	hierarchy := authorization.GetPlatformRoleHierarchy()
	if _, err := s.enforcer.AddGroupingPolicies(hierarchy); err != nil {
		return fmt.Errorf("failed to add platform role hierarchy: %w", err)
	}
	log.Printf("   ✅ Seeded %d platform role hierarchy entries", len(hierarchy))
	return nil
}

func (s *permissionSeeder) seedAccountPolicies() error {
	policies := authorization.GetAccountPolicies()
	if _, err := s.enforcer.AddPolicies(policies); err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}
	log.Printf("   ✅ Seeded %d account policies", len(policies))
	return nil
}

func (s *permissionSeeder) seedAccountRoleHierarchy() error {
	hierarchy := authorization.GetAccountRoleHierarchy()
	if _, err := s.enforcer.AddGroupingPolicies(hierarchy); err != nil {
		return fmt.Errorf("failed to add account role hierarchy: %w", err)
	}
	log.Printf("   ✅ Seeded %d account role hierarchy entries", len(hierarchy))
	return nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func IsSeeded(db *gorm.DB) (bool, error) {
	var count int64
	if err := db.Model(&PolicyVersion{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check policy versions: %w", err)
	}
	return count > 0, nil
}

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
// ACCOUNT ROLE ASSIGNMENT HELPERS
// ============================================================

// AssignAccountAdmin assigns the account_admin role to a user in an account.
func AssignAccountAdmin(db *gorm.DB, accountID, userID string) error {
	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()
	roleManager := authorization.NewRoleManager(enforcer)

	domain := authdomain.AccountDomain(accountID)

	if err := roleManager.AssignRole(ctx, domain, userID, authdomain.RoleAccountAdmin.String()); err != nil {
		return fmt.Errorf("failed to assign account admin role: %w", err)
	}

	log.Printf("✅ Assigned account_admin role for account %s to user %s", accountID, userID)
	return nil
}

// AssignTrainer assigns the trainer role to a user in an account.
func AssignTrainer(db *gorm.DB, accountID, userID string) error {
	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()
	roleManager := authorization.NewRoleManager(enforcer)

	domain := authdomain.AccountDomain(accountID)

	if err := roleManager.AssignRole(ctx, domain, userID, authdomain.RoleTrainer.String()); err != nil {
		return fmt.Errorf("failed to assign trainer role: %w", err)
	}

	log.Printf("✅ Assigned trainer role for account %s to user %s", accountID, userID)
	return nil
}