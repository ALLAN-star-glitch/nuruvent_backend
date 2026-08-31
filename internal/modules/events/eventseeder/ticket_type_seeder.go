// internal/modules/events/infrastructure/seeder/ticket_type_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedTicketTypes seeds the ticket types from domain constants
func SeedTicketTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding ticket types...")

	infos := domain.AllTicketTypeInfos()

	for _, info := range infos {
		var existing postgres.TicketTypeModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.SortOrder = info.SortOrder
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated ticket type: %s", info.Name)
			continue
		}

		// Create new
		ticketType := &postgres.TicketTypeModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			SortOrder:   info.SortOrder,
			IsActive:    info.IsActive,
		}
		if err := db.Create(ticketType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created ticket type: %s", info.Name)
	}

	log.Printf("✅ Ticket types seeded: %d types", len(infos))
	return nil
}