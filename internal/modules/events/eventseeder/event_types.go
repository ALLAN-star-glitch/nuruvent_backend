// internal/modules/events/infrastructure/seeder/event_type_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedEventTypes seeds the event types from domain constants
func SeedEventTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding event types...")

	// Get event types from domain constants
	infos := domain.AllEventTypeInfos()

	for _, info := range infos {
		var existing postgres.EventTypeModel
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
			existing.SupportsCertificate = info.SupportsCertificate
			existing.MinDuration = info.MinDuration
			existing.MaxDuration = info.MaxDuration
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated event type: %s", info.Name)
			continue
		}

		// Create new
		eventType := &postgres.EventTypeModel{
			Slug:                info.Slug, // ✅ info.Slug is already a string
			Name:                info.Name,
			DisplayName:         info.DisplayName,
			Description:         info.Description,
			Icon:                info.Icon,
			Color:               info.Color,
			SortOrder:           info.SortOrder,
			SupportsCertificate: info.SupportsCertificate,
			MinDuration:         info.MinDuration,
			MaxDuration:         info.MaxDuration,
			IsActive:            info.IsActive,
		}
		if err := db.Create(eventType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created event type: %s", info.Name)
	}

	log.Printf("✅ Event types seeded: %d types", len(infos))
	return nil
}