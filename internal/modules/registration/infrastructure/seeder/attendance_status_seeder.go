// internal/modules/registration/infrastructure/seeder/attendance_status_seeder.go

package registrationseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// SeedAttendanceStatuses seeds the attendance statuses from domain constants
func SeedAttendanceStatuses(db *gorm.DB) error {
	log.Println("🌱 Seeding attendance statuses...")

	infos := registrationdomain.AllAttendanceStatusInfos()

	for _, info := range infos {
		var existing postgres.AttendanceStatusModel
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
			existing.CountsAsAttended = info.CountsAsAttended
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated attendance status: %s", info.Name)
			continue
		}

		status := &postgres.AttendanceStatusModel{
			Slug:             info.Slug,
			Name:             info.Name,
			DisplayName:      info.DisplayName,
			Description:      info.Description,
			Color:            info.Color,
			Icon:             info.Icon,
			SortOrder:        info.SortOrder,
			IsFinal:          info.IsFinal,
			IsActive:         info.IsActive,
			CountsAsAttended: info.CountsAsAttended,
		}
		if err := db.Create(status).Error; err != nil {
			return err
		}
		log.Printf("✅ Created attendance status: %s", info.Name)
	}

	log.Printf("✅ Attendance statuses seeded: %d statuses", len(infos))
	return nil
}