// internal/database/seeders/institution_seeders/institution_types.go

package institutionseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedInstitutionTypes seeds the institution types into the database
func SeedInstitutionTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding institution types...")

	for _, it := range constants.InstitutionTypes {
		// Check if type already exists
		var existing models.InstitutionType
		err := db.Where("slug = ?", it.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = it.Slug
			existing.Name = it.Name
			existing.DisplayName = it.DisplayName
			existing.Description = it.Description
			existing.Icon = it.Icon
			existing.Color = it.Color
			existing.SortOrder = it.SortOrder
			existing.MetaTitle = it.MetaTitle
			existing.MetaDescription = it.MetaDescription
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated institution type: %s", it.Name)
			continue
		}

		// Create new
		institutionType := &models.InstitutionType{
			Slug:            it.Slug,
			Name:            it.Name,
			DisplayName:     it.DisplayName,
			Description:     it.Description,
			Icon:            it.Icon,
			Color:           it.Color,
			SortOrder:       it.SortOrder,
			MetaTitle:       it.MetaTitle,
			MetaDescription: it.MetaDescription,
			IsActive:        true,
		}
		if err := db.Create(institutionType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created institution type: %s", it.Name)
	}

	log.Printf("✅ Institution types seeded: %d types", len(constants.InstitutionTypes))
	return nil
}