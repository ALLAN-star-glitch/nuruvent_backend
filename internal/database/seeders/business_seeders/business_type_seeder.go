package bizseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedBusinessTypes seeds the business types into the database
func SeedBusinessTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding business types...")

	for _, bt := range constants.BusinessTypes {
		// Check if type already exists
		var existing models.BusinessType
		err := db.Where("name = ?", bt.Name).First(&existing).Error
		if err == nil {
			// Update existing
			existing.DisplayName = bt.DisplayName
			existing.Description = bt.Description
			existing.Icon = bt.Icon
			existing.SortOrder = bt.SortOrder
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated business type: %s", bt.DisplayName)
			continue
		}

		// Create new
		businessType := &models.BusinessType{
			Name:        bt.Name,
			DisplayName: bt.DisplayName,
			Description: bt.Description,
			Icon:        bt.Icon,
			SortOrder:   bt.SortOrder,
			IsActive:    true,
		}
		if err := db.Create(businessType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created business type: %s", bt.DisplayName)
	}

	log.Printf("✅ Business types seeded: %d types", len(constants.BusinessTypes))
	return nil
}