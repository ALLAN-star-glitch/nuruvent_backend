// internal/modules/events/infrastructure/seeder/category_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedCategories seeds the categories from domain constants
func SeedCategories(db *gorm.DB) error {
	log.Println("🌱 Seeding categories...")

	infos := domain.AllCategoryInfos()

	for _, info := range infos {
		var existing postgres.CategoryModel
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
			log.Printf("✅ Updated category: %s", info.Name)
			continue
		}

		// Create new
		category := &postgres.CategoryModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			Color:       info.Color,
			SortOrder:   info.SortOrder,
			IsActive:    info.IsActive,
		}
		if err := db.Create(category).Error; err != nil {
			return err
		}
		log.Printf("✅ Created category: %s", info.Name)
	}

	log.Printf("✅ Categories seeded: %d categories", len(infos))
	return nil
}