// internal/database/seeders/event_seeders/event_types.go

package eventseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedEventTypes seeds the event types into the database
func SeedEventTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding event types...")

	for _, et := range constants.EventTypes {
		// Check if type already exists
		var existing models.EventType
		err := db.Where("slug = ?", et.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = et.Slug
			existing.Name = et.Name
			existing.DisplayName = et.DisplayName
			existing.Description = et.Description
			existing.Icon = et.Icon
			existing.Color = et.Color
			existing.SortOrder = et.SortOrder
			existing.SupportsCertificate = et.SupportsCertificate
			existing.MinDuration = et.MinDuration
			existing.MaxDuration = et.MaxDuration
			existing.MetaTitle = et.MetaTitle
			existing.MetaDescription = et.MetaDescription
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated event type: %s", et.Name)
			continue
		}

		// Create new
		eventType := &models.EventType{
			Slug:                et.Slug,
			Name:                et.Name,
			DisplayName:         et.DisplayName,
			Description:         et.Description,
			Icon:                et.Icon,
			Color:               et.Color,
			SortOrder:           et.SortOrder,
			SupportsCertificate: et.SupportsCertificate,
			MinDuration:         et.MinDuration,
			MaxDuration:         et.MaxDuration,
			MetaTitle:           et.MetaTitle,
			MetaDescription:     et.MetaDescription,
			IsActive:            true,
		}
		if err := db.Create(eventType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created event type: %s", et.Name)
	}

	log.Printf("✅ Event types seeded: %d types", len(constants.EventTypes))
	return nil
}