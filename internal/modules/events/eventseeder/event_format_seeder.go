// internal/modules/events/infrastructure/seeder/event_format_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedEventFormats seeds the event formats from domain constants
func SeedEventFormats(db *gorm.DB) error {
	log.Println("🌱 Seeding event formats...")

	infos := domain.AllEventFormatInfos()

	for _, info := range infos {
		var existing postgres.EventFormatModel
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
			log.Printf("✅ Updated event format: %s", info.Name)
			continue
		}

		// Create new
		eventFormat := &postgres.EventFormatModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Icon:        info.Icon,
			IsActive:    info.IsActive,
		}
		if err := db.Create(eventFormat).Error; err != nil {
			return err
		}
		log.Printf("✅ Created event format: %s", info.Name)
	}

	log.Printf("✅ Event formats seeded: %d formats", len(infos))
	return nil
}