package permseeder

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"os"
	"strings"

	"gorm.io/gorm"
)

// ============================================================
// PUBLIC ENTRY FUNCTION
// ============================================================

// SeedPermissions seeds the platform permissions
func SeedPermissions(db *gorm.DB) error {
	log.Println("🌱 Seeding platform permissions...")

	cfg := config.Load()

	// Initialize enforcer
	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	// ✅ Create service (returns domain.PermissionService interface)
	svc := authorization.NewService(enforcer)

	// ✅ Create seeder with interface
	seeder := &permissionSeeder{
		enforcer: enforcer,
		service:  svc, // ✅ domain.PermissionService interface
	}

	// Check if policies already exist
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get policies: %w", err)
	}

	if len(policies) > 0 {
		log.Printf("⚠️  Policies already exist (%d rules), skipping seed", len(policies))
		log.Printf("ℹ️  To re-seed, truncate casbin_rule table first")
		return nil
	}

	// Seed policies from CSV
	if err := seeder.seedPoliciesFromCSV(); err != nil {
		return err
	}

	// Seed role hierarchy from CSV
	if err := seeder.seedRoleHierarchyFromCSV(); err != nil {
		return err
	}

	log.Println("✅ Platform permissions seeded successfully")
	log.Println("ℹ️  Account policies will be added when accounts are created")
	return nil
}

// ============================================================
// INTERNAL SEEDER
// ============================================================

// permissionSeeder handles the actual seeding
type permissionSeeder struct {
	enforcer *authorization.Enforcer
	service  domain.PermissionService // ✅ Use interface
}

// seedPoliciesFromCSV seeds platform policies from default_policies.csv
func (s *permissionSeeder) seedPoliciesFromCSV() error {
	file, err := os.Open("configs/casbin/policies/default_policies.csv")
	if err != nil {
		return fmt.Errorf("failed to open policies CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var policies [][]string
	var skippedCount int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV: %w", err)
		}

		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		if len(record) >= 5 {
			policyType := strings.TrimSpace(record[0])
			if policyType == "p" {
				sub := strings.TrimSpace(record[1])
				dom := strings.TrimSpace(record[2])
				obj := strings.TrimSpace(record[3])
				act := strings.TrimSpace(record[4])

				if strings.Contains(dom, "{{.AccountID}}") {
					skippedCount++
					continue
				}

				policies = append(policies, []string{sub, dom, obj, act})
			}
		}
	}

	if len(policies) > 0 {
		_, err := s.enforcer.AddPolicies(policies)
		if err != nil {
			return fmt.Errorf("failed to add policies: %w", err)
		}
		log.Printf("✅ Seeded %d platform policies from CSV", len(policies))
	}

	if skippedCount > 0 {
		log.Printf("ℹ️  Skipped %d account policies (handled in code)", skippedCount)
	}

	return nil
}

// seedRoleHierarchyFromCSV seeds platform role hierarchy from role_hierarchy.csv
func (s *permissionSeeder) seedRoleHierarchyFromCSV() error {
	file, err := os.Open("configs/casbin/policies/role_hierarchy.csv")
	if err != nil {
		return fmt.Errorf("failed to open role hierarchy CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	var rules [][]string
	var skippedCount int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV: %w", err)
		}

		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		if len(record) >= 4 {
			policyType := strings.TrimSpace(record[0])
			if policyType == "g" {
				user := strings.TrimSpace(record[1])
				role := strings.TrimSpace(record[2])
				domain := strings.TrimSpace(record[3])

				if strings.Contains(domain, "{{.AccountID}}") {
					skippedCount++
					continue
				}

				rules = append(rules, []string{user, role, domain})
			}
		}
	}

	if len(rules) > 0 {
		_, err := s.enforcer.AddGroupingPolicies(rules)
		if err != nil {
			return fmt.Errorf("failed to add role hierarchy: %w", err)
		}
		log.Printf("✅ Seeded %d platform role hierarchy entries from CSV", len(rules))
	}

	if skippedCount > 0 {
		log.Printf("ℹ️  Skipped %d account role hierarchies (handled in code)", skippedCount)
	}

	return nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// IsSeeded checks if the database has been seeded
func IsSeeded(db *gorm.DB) (bool, error) {
	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return false, fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	policies, err := enforcer.GetPolicy()
	if err != nil {
		return false, fmt.Errorf("failed to get policies: %w", err)
	}
	return len(policies) > 0, nil
}

// SeedAccountPolicies adds account policies for an existing account
func SeedAccountPolicies(db *gorm.DB, accountID string) error {
	log.Printf("📦 Adding account policies for account: %s", accountID)

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	return service.AddAccountPolicies(ctx, accountID)
}

// SeedAccountPoliciesForAllAccounts adds account policies for all existing accounts
func SeedAccountPoliciesForAllAccounts(db *gorm.DB) error {
	log.Println("📦 Adding account policies for all existing accounts...")

	cfg := config.Load()

	enforcer, err := authorization.NewEnforcer(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init enforcer: %w", err)
	}
	defer enforcer.Close()

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	// Get all account IDs
	var accountIDs []string
	if err := db.Table("accounts").Pluck("id", &accountIDs).Error; err != nil {
		return fmt.Errorf("failed to get account IDs: %w", err)
	}

	if len(accountIDs) == 0 {
		log.Println("No accounts found to seed policies for")
		return nil
	}

	successCount := 0
	for _, accountID := range accountIDs {
		if err := service.AddAccountPolicies(ctx, accountID); err != nil {
			log.Printf("⚠️  Failed to seed policies for account %s: %v", accountID, err)
			continue
		}
		successCount++
	}

	log.Printf("✅ Seeded account policies for %d out of %d accounts", successCount, len(accountIDs))
	return nil
}