package authseeder

import (
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/postgres"
	"gorm.io/gorm"
)

// SeedAccountTypes seeds the account types from domain constants
func SeedAccountTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding account types...")

	// Get account types from domain constants
	infos := authdomain.AllAccountTypeInfos()

	for _, info := range infos {
		var existing postgres.AccountTypeModel
		// ✅ info.Slug is already a string - no need for .String()
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.Icon = info.Icon
			existing.Color = info.Color
			existing.SortOrder = info.SortOrder
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated account type: %s", info.Name)
			continue
		}

		// Create new
		accountType := &postgres.AccountTypeModel{
			Slug:        info.Slug, // ✅ info.Slug is already a string
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			Color:       info.Color,
			SortOrder:   info.SortOrder,
			IsActive:    info.IsActive,
		}
		if err := db.Create(accountType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created account type: %s", info.Name)
	}

	log.Printf("✅ Account types seeded: %d types", len(infos))
	return nil
}