package accseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/postgres"
)

// SeedInstitutionTypes seeds the institution types from domain constants
func SeedInstitutionTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding institution types...")

	infos := domain.AllInstitutionTypeInfos()

	for _, info := range infos {
		var existing postgres.InstitutionTypeModel
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
			log.Printf("✅ Updated institution type: %s", info.Name)
			continue
		}

		institutionType := &postgres.InstitutionTypeModel{
			Slug:        info.Slug, // ✅ info.Slug is already a string
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			Color:       info.Color,
			SortOrder:   info.SortOrder,
			IsActive:    info.IsActive,
		}
		if err := db.Create(institutionType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created institution type: %s", info.Name)
	}

	log.Printf("✅ Institution types seeded: %d types", len(infos))
	return nil
}