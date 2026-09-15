// internal/modules/registration/infrastructure/seeder/registration_status_seeder.go

package registrationseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// SeedRegistrationStatuses seeds the registration statuses from domain constants
func SeedRegistrationStatuses(db *gorm.DB) error {
	log.Println("🌱 Seeding registration statuses...")

	infos := registrationdomain.AllRegistrationStatusInfos()

	for _, info := range infos {
		var existing postgres.RegistrationStatusModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
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
			log.Printf("✅ Updated registration status: %s", info.Name)
			continue
		}

		status := &postgres.RegistrationStatusModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Color:       info.Color,
			Icon:        info.Icon,
			SortOrder:   info.SortOrder,
			IsFinal:     info.IsFinal,
			IsActive:    info.IsActive,
		}
		if err := db.Create(status).Error; err != nil {
			return err
		}
		log.Printf("✅ Created registration status: %s", info.Name)
	}

	log.Printf("✅ Registration statuses seeded: %d statuses", len(infos))
	return nil
}