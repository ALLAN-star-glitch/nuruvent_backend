// internal/database/seeders/event_seeders/event_statuses.go

package eventseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedEventStatuses seeds the event statuses into the database
func SeedEventStatuses(db *gorm.DB) error {
	log.Println("🌱 Seeding event statuses...")

	for _, es := range constants.EventStatuses {
		// Check if status already exists
		var existing models.EventStatus
		err := db.Where("slug = ?", es.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = es.Slug
			existing.Name = es.Name
			existing.DisplayName = es.DisplayName
			existing.Description = es.Description
			existing.Color = es.Color
			existing.Icon = es.Icon
			existing.SortOrder = es.SortOrder
			existing.IsFinal = es.IsFinal
			existing.CanEdit = es.CanEdit
			existing.CanRegister = es.CanRegister
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated event status: %s", es.Name)
			continue
		}

		// Create new
		eventStatus := &models.EventStatus{
			Slug:        es.Slug,
			Name:        es.Name,
			DisplayName: es.DisplayName,
			Description: es.Description,
			Color:       es.Color,
			Icon:        es.Icon,
			SortOrder:   es.SortOrder,
			IsFinal:     es.IsFinal,
			CanEdit:     es.CanEdit,
			CanRegister: es.CanRegister,
			IsActive:    true,
		}
		if err := db.Create(eventStatus).Error; err != nil {
			return err
		}
		log.Printf("✅ Created event status: %s", es.Name)
	}

	log.Printf("✅ Event statuses seeded: %d statuses", len(constants.EventStatuses))
	return nil
}