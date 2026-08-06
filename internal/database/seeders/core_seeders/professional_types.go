// internal/database/seeders/core_seeders/professional_types.go

package coreseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedProfessionalTypes seeds the professional types into the database
func SeedProfessionalTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding professional types...")

	for _, pt := range constants.ProfessionalTypes {
		// Check if type already exists
		var existing models.ProfessionalType
		err := db.Where("slug = ?", pt.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = pt.Slug
			existing.Name = pt.Name
			existing.DisplayName = pt.DisplayName
			existing.Description = pt.Description
			existing.Icon = pt.Icon
			existing.Color = pt.Color
			existing.SortOrder = pt.SortOrder
			existing.CanHost = pt.CanHost
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated professional type: %s", pt.Name)
			continue
		}

		// Create new
		professionalType := &models.ProfessionalType{
			Slug:        pt.Slug,
			Name:        pt.Name,
			DisplayName: pt.DisplayName,
			Description: pt.Description,
			Icon:        pt.Icon,
			Color:       pt.Color,
			SortOrder:   pt.SortOrder,
			CanHost:     pt.CanHost,
			IsActive:    true,
		}
		if err := db.Create(professionalType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created professional type: %s", pt.Name)
	}

	log.Printf("✅ Professional types seeded: %d types", len(constants.ProfessionalTypes))
	return nil
}