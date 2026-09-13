package authseeder

import (
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/postgres"
	"gorm.io/gorm"
)

// SeedProfessionalTypes seeds the professional types from domain constants
func SeedProfessionalTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding professional types...")

	infos := authdomain.AllProfessionalTypeInfos()

	for _, info := range infos {
		var existing postgres.ProfessionalTypeModel
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
			log.Printf("✅ Updated professional type: %s", info.Name)
			continue
		}

		professionalType := &postgres.ProfessionalTypeModel{
			Slug:        info.Slug, // ✅ info.Slug is already a string
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			Color:       info.Color,
			SortOrder:   info.SortOrder,
			IsActive:    info.IsActive,
		}
		if err := db.Create(professionalType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created professional type: %s", info.Name)
	}

	log.Printf("✅ Professional types seeded: %d types", len(infos))
	return nil
}