// internal/modules/events/infrastructure/seeder/recurrence_pattern_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedRecurrencePatterns seeds the recurrence patterns from domain constants
func SeedRecurrencePatterns(db *gorm.DB) error {
	log.Println("🌱 Seeding recurrence patterns...")

	infos := domain.AllRecurrencePatternInfos()

	for _, info := range infos {
		var existing postgres.RecurrencePatternModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated recurrence pattern: %s", info.Name)
			continue
		}

		// Create new
		recurrencePattern := &postgres.RecurrencePatternModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			IsActive:    info.IsActive,
		}
		if err := db.Create(recurrencePattern).Error; err != nil {
			return err
		}
		log.Printf("✅ Created recurrence pattern: %s", info.Name)
	}

	log.Printf("✅ Recurrence patterns seeded: %d patterns", len(infos))
	return nil
}