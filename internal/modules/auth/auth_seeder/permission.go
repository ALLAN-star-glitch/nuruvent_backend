package authseeder

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"

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

	// Create service
	svc := authorization.NewService(enforcer)

	// Create seeder
	seeder := &permissionSeeder{
		enforcer: enforcer,
		service:  svc,
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

	// Seed platform policies
	if err := seeder.seedPlatformPolicies(); err != nil {
		return err
	}

	// Seed platform role hierarchy
	if err := seeder.seedPlatformRoleHierarchy(); err != nil {
		return err
	}

	log.Println("✅ Platform permissions seeded successfully")
	log.Println("ℹ️  Personal and Institution team policies will be added when users/institutions are created")
	return nil
}

// ============================================================
// INTERNAL SEEDER
// ============================================================

// permissionSeeder handles the actual seeding
type permissionSeeder struct {
	enforcer *authorization.Enforcer
	service  authdomain.PermissionService
}

// seedPlatformPolicies seeds platform policies from default_policies.csv
func (s *permissionSeeder) seedPlatformPolicies() error {
    file, err := os.Open("configs/casbin/policies/default_policies.csv")
    if err != nil {
        return fmt.Errorf("failed to open policies CSV: %w", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comment = '#'
    reader.TrimLeadingSpace = true
    reader.FieldsPerRecord = -1

    var platformPolicies [][]string
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

                // Skip team-specific policies (handled in code)
                if strings.Contains(dom, "{{.TeamID}}") ||
                    strings.Contains(dom, "{team_id}") ||
                    strings.Contains(dom, "{{.UserID}}") ||
                    strings.Contains(dom, "{user_id}") ||
                    strings.Contains(dom, "{{.InstitutionID}}") ||
                    strings.Contains(dom, "{institution_id}") ||
                    strings.Contains(dom, "personal:team:{") ||    // ← ADD THIS
                    strings.Contains(dom, "institution:team:{") {  // ← ADD THIS
                    skippedCount++
                    continue
                }

                // Platform policies (domain = "platform")
                if dom == "platform" {
                    platformPolicies = append(platformPolicies, []string{sub, dom, obj, act})
                } else {
                    // Skip any other policies
                    skippedCount++
                }
            }
        }
    }

    // Add platform policies
    if len(platformPolicies) > 0 {
        _, err := s.enforcer.AddPolicies(platformPolicies)
        if err != nil {
            return fmt.Errorf("failed to add platform policies: %w", err)
        }
        log.Printf("✅ Seeded %d platform policies from CSV", len(platformPolicies))
    }

    if skippedCount > 0 {
        log.Printf("ℹ️  Skipped %d team policy templates (handled in code)", skippedCount)
    }

    return nil
}

// seedPlatformRoleHierarchy seeds platform role hierarchy from role_hierarchy.csv
func (s *permissionSeeder) seedPlatformRoleHierarchy() error {
    file, err := os.Open("configs/casbin/policies/role_hierarchy.csv")
    if err != nil {
        return fmt.Errorf("failed to open role hierarchy CSV: %w", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    reader.Comment = '#'
    reader.TrimLeadingSpace = true
    reader.FieldsPerRecord = -1

    var platformRules [][]string
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

                // Skip team-specific rules (handled in code)
                if strings.Contains(domain, "{{.TeamID}}") ||
                    strings.Contains(domain, "{team_id}") ||
                    strings.Contains(domain, "{{.UserID}}") ||
                    strings.Contains(domain, "{user_id}") ||
                    strings.Contains(domain, "{{.InstitutionID}}") ||
                    strings.Contains(domain, "{institution_id}") ||
                    strings.Contains(domain, "personal:team:{") ||    // ← ADD THIS
                    strings.Contains(domain, "institution:team:{") {  // ← ADD THIS
                    skippedCount++
                    continue
                }

                // Platform rules (domain = "platform")
                if domain == "platform" {
                    platformRules = append(platformRules, []string{user, role, domain})
                } else {
                    skippedCount++
                }
            }
        }
    }

    // Add platform role hierarchy
    if len(platformRules) > 0 {
        _, err := s.enforcer.AddGroupingPolicies(platformRules)
        if err != nil {
            return fmt.Errorf("failed to add platform role hierarchy: %w", err)
        }
        log.Printf("✅ Seeded %d platform role hierarchy entries from CSV", len(platformRules))
    }

    if skippedCount > 0 {
        log.Printf("ℹ️  Skipped %d team role hierarchy entries (handled in code)", skippedCount)
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

// ============================================================
// TEAM POLICY SEEDING
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

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	// Add personal team policies
	if err := service.AddPersonalTeamPolicies(ctx, userID); err != nil {
		return fmt.Errorf("failed to add personal team policies: %w", err)
	}

	// Add personal team role hierarchy
	domain := authdomain.PersonalTeamDomain(userID)
	if err := seedTeamRoleHierarchy(enforcer, domain); err != nil {
		return fmt.Errorf("failed to seed personal team role hierarchy: %w", err)
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

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	// Add institution team policies
	if err := service.AddInstitutionPolicies(ctx, institutionID); err != nil {
		return fmt.Errorf("failed to add institution policies: %w", err)
	}

	// Add institution team role hierarchy
	domain := authdomain.InstitutionTeamDomain(institutionID)
	if err := seedTeamRoleHierarchy(enforcer, domain); err != nil {
		return fmt.Errorf("failed to seed institution team role hierarchy: %w", err)
	}

	log.Printf("✅ Institution team policies seeded for institution: %s", institutionID)
	return nil
}

// seedTeamRoleHierarchy adds role hierarchy for a team domain
func seedTeamRoleHierarchy(enforcer *authorization.Enforcer, domain string) error {
	hierarchy := [][]string{
		// Account Admin inherits Event Manager
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleEventManager.String(), domain},
		// Account Admin inherits Team Member
		{authdomain.RoleAccountAdmin.String(), authdomain.RoleTeamMember.String(), domain},
		// Event Manager inherits Team Member
		{authdomain.RoleEventManager.String(), authdomain.RoleTeamMember.String(), domain},
	}

	_, err := enforcer.AddGroupingPolicies(hierarchy)
	if err != nil {
		return fmt.Errorf("failed to add team role hierarchy: %w", err)
	}

	log.Printf("✅ Team role hierarchy seeded for domain: %s", domain)
	return nil
}

// ============================================================
// BULK SEEDING FUNCTIONS
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

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	// Get all user IDs
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

		// Check if policies already exist
		roles := enforcer.GetRolesForUserInDomain(userID, domain)
		if len(roles) > 0 {
			log.Printf("⚠️  User %s already has personal team roles, skipping", userID)
			continue
		}

		if err := service.AddPersonalTeamPolicies(ctx, userID); err != nil {
			log.Printf("⚠️  Failed to seed personal team policies for user %s: %v", userID, err)
			continue
		}

		// Add role hierarchy
		if err := seedTeamRoleHierarchy(enforcer, domain); err != nil {
			log.Printf("⚠️  Failed to seed personal team role hierarchy for user %s: %v", userID, err)
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

	service := authorization.NewService(enforcer)
	ctx := context.Background()

	// Get all institution IDs
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

		// Check if policies already exist
		policies, err := enforcer.GetFilteredPolicy(1, domain)
		if err != nil {
			log.Printf("⚠️  Failed to check policies for institution %s: %v", institutionID, err)
			continue
		}

		if len(policies) > 0 {
			log.Printf("⚠️  Institution %s already has policies, skipping", institutionID)
			continue
		}

		if err := service.AddInstitutionPolicies(ctx, institutionID); err != nil {
			log.Printf("⚠️  Failed to seed institution team policies for institution %s: %v", institutionID, err)
			continue
		}

		// Add role hierarchy
		if err := seedTeamRoleHierarchy(enforcer, domain); err != nil {
			log.Printf("⚠️  Failed to seed institution team role hierarchy for institution %s: %v", institutionID, err)
		}

		successCount++
	}

	log.Printf("✅ Seeded institution team policies for %d out of %d institutions", successCount, len(institutionIDs))
	return nil
}