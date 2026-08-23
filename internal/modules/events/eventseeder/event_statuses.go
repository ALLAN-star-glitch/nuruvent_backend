// internal/modules/events/infrastructure/seeder/event_status_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedEventStatuses seeds the event statuses from domain constants
func SeedEventStatuses(db *gorm.DB) error {
	log.Println("🌱 Seeding event statuses...")

	infos := domain.AllEventStatusInfos()

	for _, info := range infos {
		var existing postgres.EventStatusModel
		// ✅ info.Slug is already a string - no need for .String()
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.Color = info.Color
			existing.Icon = info.Icon
			existing.SortOrder = info.SortOrder
			existing.IsFinal = info.IsFinal
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated event status: %s", info.Name)
			continue
		}

		// Create new
		eventStatus := &postgres.EventStatusModel{
			Slug:        info.Slug, // ✅ info.Slug is already a string
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Color:       info.Color,
			Icon:        info.Icon,
			SortOrder:   info.SortOrder,
			IsFinal:     info.IsFinal,
			IsActive:    info.IsActive,
		}
		if err := db.Create(eventStatus).Error; err != nil {
			return err
		}
		log.Printf("✅ Created event status: %s", info.Name)
	}

	log.Printf("✅ Event statuses seeded: %d statuses", len(infos))
	return nil
}