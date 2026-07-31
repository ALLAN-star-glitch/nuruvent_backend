// internal/database/seeders/core_seeders/permissions.go

package coreseeder

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"gorm.io/gorm"
)

// ============================================================
// PUBLIC ENTRY FUNCTION
// ============================================================

// SeedPermissions seeds the platform permissions
// This is the main entry point for the seeder
func SeedPermissions(db *gorm.DB) error {
	log.Println("🌱 Seeding platform permissions...")

	cfg := config.Load()

	// Initialize authorization module
	permModule, err := authorization.NewModule(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init authorization module: %w", err)
	}
	defer permModule.Close()

	enforcer := permModule.GetEnforcer()
	service := permModule.GetService()

	// Create seeder
	seeder := &permissionSeeder{
		enforcer: enforcer,
		service:  service,
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
	log.Println("ℹ️  Business policies will be added when businesses are created")
	return nil
}

// ============================================================
// INTERNAL SEEDER
// ============================================================

// permissionSeeder handles the actual seeding
type permissionSeeder struct {
	enforcer *authorization.Enforcer
	service  *authorization.Service
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

				if strings.Contains(dom, "{{.BusinessID}}") {
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
		log.Printf("ℹ️  Skipped %d business policies (handled in code)", skippedCount)
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

				if strings.Contains(domain, "{{.BusinessID}}") {
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
		log.Printf("ℹ️  Skipped %d business role hierarchies (handled in code)", skippedCount)
	}

	return nil
}

// ============================================================
// HELPER FUNCTIONS (Public)
// ============================================================

// IsSeeded checks if the database has been seeded
func IsSeeded(db *gorm.DB) (bool, error) {
	cfg := config.Load()

	permModule, err := authorization.NewModule(db, cfg)
	if err != nil {
		return false, fmt.Errorf("failed to init authorization module: %w", err)
	}
	defer permModule.Close()

	enforcer := permModule.GetEnforcer()
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return false, fmt.Errorf("failed to get policies: %w", err)
	}
	return len(policies) > 0, nil
}

// SeedBusinessPolicies adds business policies for an existing business
// This is useful for backfilling businesses created before policies were added
func SeedBusinessPolicies(db *gorm.DB, businessID string) error {
	log.Printf("📦 Adding business policies for business: %s", businessID)

	cfg := config.Load()

	permModule, err := authorization.NewModule(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init authorization module: %w", err)
	}
	defer permModule.Close()

	service := permModule.GetService()
	ctx := context.Background()

	return service.AddBusinessPolicies(ctx, businessID)
}

// SeedBusinessPoliciesForAllBusinesses adds business policies for all existing businesses
func SeedBusinessPoliciesForAllBusinesses(db *gorm.DB) error {
	log.Println("📦 Adding business policies for all existing businesses...")

	cfg := config.Load()

	permModule, err := authorization.NewModule(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to init authorization module: %w", err)
	}
	defer permModule.Close()

	service := permModule.GetService()
	ctx := context.Background()

	// Get all business IDs
	var businessIDs []string
	if err := db.Table("businesses").Pluck("id", &businessIDs).Error; err != nil {
		return fmt.Errorf("failed to get business IDs: %w", err)
	}

	if len(businessIDs) == 0 {
		log.Println("No businesses found to seed policies for")
		return nil
	}

	successCount := 0
	for _, businessID := range businessIDs {
		if err := service.AddBusinessPolicies(ctx, businessID); err != nil {
			log.Printf("⚠️  Failed to seed policies for business %s: %v", businessID, err)
			continue
		}
		successCount++
	}

	log.Printf("✅ Seeded business policies for %d out of %d businesses", successCount, len(businessIDs))
	return nil
}