// internal/modules/events/infrastructure/seeder/material_type_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedMaterialTypes seeds the material types from domain constants
func SeedMaterialTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding material types...")

	infos := domain.AllMaterialTypeInfos()

	for _, info := range infos {
		var existing postgres.MaterialTypeModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.Icon = info.Icon
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated material type: %s", info.Name)
			continue
		}

		// Create new
		materialType := &postgres.MaterialTypeModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			IsActive:    info.IsActive,
		}
		if err := db.Create(materialType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created material type: %s", info.Name)
	}

	log.Printf("✅ Material types seeded: %d types", len(infos))
	return nil
}