package authseeder

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// SeedSuperAdmins promotes a specific list of users to platform super_admin.
// Safe to run multiple times — uses ON CONFLICT DO NOTHING semantics.
//
// EDIT the list below to include the user IDs you want promoted.
func SeedSuperAdmins(db *gorm.DB) error {
	log.Println("🌱 Seeding platform super admins...")

	// 👇 Add real user IDs here.
	superAdminUserIDs := []string{
		"5887abbd-68ec-4815-9807-5c83f491c18f", // allanmathenge67@gmail.com
		// "another-uuid-here",
	}

	for _, userID := range superAdminUserIDs {
		// Check if the rule already exists
		var count int64
		db.Table("casbin_rule").
			Where("ptype = 'g' AND v0 = ? AND v1 = 'super_admin' AND v2 = 'platform'", userID).
			Count(&count)

		if count > 0 {
			log.Printf("⏭️  User %s already super_admin, skipping", userID)
			continue
		}

		err := db.Table("casbin_rule").Create(map[string]interface{}{
			"ptype": "g",
			"v0":    userID,
			"v1":    "super_admin",
			"v2":    "platform",
		}).Error
		if err != nil {
			return fmt.Errorf("failed to promote %s: %w", userID, err)
		}

		log.Printf("✅ Promoted %s to super_admin", userID)
	}

	log.Println("✅ Super admin seeder complete")
	return nil
}