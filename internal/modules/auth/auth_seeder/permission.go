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

const CURRENT_POLICY_VERSION = "v13"

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

	if err := s.cleanTeamPolicies(); err != nil {
		return err
	}

	if err := s.cleanAccountPolicies(); err != nil {
		return err
	}

	if err := s.cleanGroupingPolicies(); err != nil {
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

func (s *permissionSeeder) cleanTeamPolicies() error {
	allPolicies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all policies: %w", err)
	}

	var teamPolicies [][]string
	for _, policy := range allPolicies {
		if len(policy) >= 4 {
			domain := policy[1]
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

func (s *permissionSeeder) cleanAccountPolicies() error {
	allPolicies, err := s.enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get all policies: %w", err)
	}

	var accountPolicies [][]string
	for _, policy := range allPolicies {
		if len(policy) >= 4 {
			domain := policy[1]
			if authdomain.IsAccountDomain(domain) {
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

func (s *permissionSeeder) cleanGroupingPolicies() error {
	log.Println("   ℹ️  Skipping grouping policy cleanup to preserve user roles")
	log.Println("   ℹ️  User roles will be managed separately via team_members table")
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

	log.Println("✅ Fresh seed completed")
	return nil
}

func (s *permissionSeeder) migratePolicies() error {
	log.Println("🔄 Migrating policies to new schema...")

	if err := s.cleanAllPolicies(); err != nil {
		return err
	}

	if err := s.seedPlatformPolicies(); err != nil {
		return err
	}

	if err := s.seedPlatformRoleHierarchy(); err != nil {
		return err
	}

	if err := s.updateTeamPolicies(); err != nil {
		return err
	}

	if err := s.updateAccountPolicies(); err != nil {
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

// ============================================================
// TEAM POLICY UPDATE METHODS
// ============================================================

func (s *permissionSeeder) updateTeamPolicies() error {
	log.Println("📝 Updating team policies for existing users and institutions...")

	if err := s.updatePersonalTeamPolicies(); err != nil {
		return err
	}

	if err := s.updateInstitutionTeamPolicies(); err != nil {
		return err
	}

	return nil
}

func (s *permissionSeeder) updatePersonalTeamPolicies() error {
	// Personal teams live in the `teams` table with type = 'personal'
	var teamIDs []string
	if err := s.db.Table("teams").
		Where("type = ? AND deleted_at IS NULL", "personal").
		Pluck("id", &teamIDs).Error; err != nil {
		return fmt.Errorf("failed to get personal team IDs: %w", err)
	}

	if len(teamIDs) == 0 {
		log.Println("   No personal teams found to update personal team policies")
		return nil
	}

	updated := 0
	for _, teamID := range teamIDs {
		domain := authdomain.PersonalTeamDomain(teamID)

		newPolicies := authorization.GetPersonalTeamPolicies(domain)
		if _, err := s.enforcer.AddPolicies(newPolicies); err != nil {
			log.Printf("   ⚠️  Failed to add policies for team %s: %v", teamID, err)
			continue
		}
		updated++
	}

	log.Printf("   ✅ Updated personal team policies for %d teams", updated)
	return nil
}

func (s *permissionSeeder) updateInstitutionTeamPolicies() error {
	// Institution teams live in the `teams` table with type = 'institution'
	var teamIDs []string
	if err := s.db.Table("teams").
		Where("type = ? AND deleted_at IS NULL", "institution").
		Pluck("id", &teamIDs).Error; err != nil {
		return fmt.Errorf("failed to get institution team IDs: %w", err)
	}

	if len(teamIDs) == 0 {
		log.Println("   No institution teams found to update institution team policies")
		return nil
	}

	updated := 0
	for _, teamID := range teamIDs {
		domain := authdomain.InstitutionTeamDomain(teamID)

		newPolicies := authorization.GetInstitutionTeamPolicies(domain)
		if _, err := s.enforcer.AddPolicies(newPolicies); err != nil {
			log.Printf("   ⚠️  Failed to add policies for team %s: %v", teamID, err)
			continue
		}
		updated++
	}

	log.Printf("   ✅ Updated institution team policies for %d teams", updated)
	return nil
}

// ============================================================
// ACCOUNT POLICY UPDATE METHODS (NEW)
// ============================================================

func (s *permissionSeeder) updateAccountPolicies() error {
	log.Println("📝 Updating account policies for existing accounts...")

	var accountIDs []string
	if err := s.db.Table("accounts").Pluck("id", &accountIDs).Error; err != nil {
		log.Printf("   ℹ️  No accounts found to update account policies (table may not exist yet)")
		return nil
	}

	if len(accountIDs) == 0 {
		log.Println("   No accounts found to update account policies")
		return nil
	}

	accountUpdated := 0
	for _, accountID := range accountIDs {
		domain := authdomain.AccountDomain(accountID)

		newPolicies := authorization.GetAccountPolicies(domain)
		if _, err := s.enforcer.AddPolicies(newPolicies); err != nil {
			log.Printf("   ⚠️  Failed to add policies for account %s: %v", accountID, err)
			continue
		}
		accountUpdated++
	}

	log.Printf("   ✅ Updated account policies for %d accounts", accountUpdated)
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
// TEAM POLICY SEEDING (for individual teams)
// ============================================================

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

	domain := authdomain.PersonalTeamDomain(userID)

	if err := policyManager.AddTeamPolicies(ctx, domain); err != nil {
		return fmt.Errorf("failed to add personal team policies: %w", err)
	}

	if err := roleManager.AssignRole(ctx, domain, userID, authdomain.RoleAccountAdmin.String()); err != nil {
		return fmt.Errorf("failed to assign account admin role: %w", err)
	}

	log.Printf("✅ Personal team policies seeded for user: %s", userID)
	return nil
}

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

	domain := authdomain.InstitutionTeamDomain(institutionID)

	if err := policyManager.AddTeamPolicies(ctx, domain); err != nil {
		return fmt.Errorf("failed to add institution team policies: %w", err)
	}

	log.Printf("✅ Institution team policies seeded for institution: %s", institutionID)
	return nil
}

// ============================================================
// ACCOUNT POLICY SEEDING (NEW)
// ============================================================

func SeedAccountPolicies(db *gorm.DB, accountID string) error {
	log.Printf("📦 Adding account policies for account: %s", accountID)

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	ctx := context.Background()
	policyManager := authorization.NewPolicyManager(enforcer)

	domain := authdomain.AccountDomain(accountID)

	if err := policyManager.AddAccountPolicies(ctx, domain); err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	log.Printf("✅ Account policies seeded for account: %s", accountID)
	return nil
}

// ============================================================
// BULK SEEDING FUNCTIONS
// ============================================================

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
		domain := authdomain.PersonalTeamDomain(userID)

		roles, err := roleManager.GetUserRoles(ctx, userID, domain)
		if err != nil {
			log.Printf("⚠️  Failed to get roles for user %s: %v", userID, err)
			continue
		}

		if len(roles) > 0 {
			continue
		}

		if err := policyManager.AddTeamPolicies(ctx, domain); err != nil {
			log.Printf("⚠️  Failed to seed personal team policies for user %s: %v", userID, err)
			continue
		}

		if err := roleManager.AssignRole(ctx, domain, userID, authdomain.RoleAccountAdmin.String()); err != nil {
			log.Printf("⚠️  Failed to assign account admin role for user %s: %v", userID, err)
			continue
		}

		successCount++
	}

	log.Printf("✅ Seeded personal team policies for %d out of %d users", successCount, len(userIDs))
	return nil
}

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
		domain := authdomain.InstitutionTeamDomain(institutionID)

		policies, err := enforcer.GetFilteredPolicy(1, domain)
		if err != nil {
			log.Printf("⚠️  Failed to check policies for institution %s: %v", institutionID, err)
			continue
		}

		if len(policies) > 0 {
			continue
		}

		if err := policyManager.AddTeamPolicies(ctx, domain); err != nil {
			log.Printf("⚠️  Failed to seed institution team policies for institution %s: %v", institutionID, err)
			continue
		}

		successCount++
	}

	log.Printf("✅ Seeded institution team policies for %d out of %d institutions", successCount, len(institutionIDs))
	return nil
}

// ============================================================
// ASSIGN ROLE FUNCTIONS
// ============================================================

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

// Deprecated: Use AssignAccountAdmin instead
func AssignInstitutionAdmin(db *gorm.DB, institutionID, userID string) error {
	return AssignAccountAdmin(db, institutionID, userID)
}