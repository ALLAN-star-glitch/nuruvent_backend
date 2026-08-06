// internal/database/seeders/core_seeders/account_types.go

package coreseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedAccountTypes seeds the account types into the database
func SeedAccountTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding account types...")

	for _, at := range constants.AccountTypes {
		// Check if type already exists
		var existing models.AccountType
		err := db.Where("slug = ?", at.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = at.Slug
			existing.Name = at.Name
			existing.DisplayName = at.DisplayName
			existing.Description = at.Description
			existing.Icon = at.Icon
			existing.Color = at.Color
			existing.SortOrder = at.SortOrder
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated account type: %s", at.Name)
			continue
		}

		// Create new
		accountType := &models.AccountType{
			Slug:        at.Slug,
			Name:        at.Name,
			DisplayName: at.DisplayName,
			Description: at.Description,
			Icon:        at.Icon,
			Color:       at.Color,
			SortOrder:   at.SortOrder,
			IsActive:    true,
		}
		if err := db.Create(accountType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created account type: %s", at.Name)
	}

	log.Printf("✅ Account types seeded: %d types", len(constants.AccountTypes))
	return nil
}