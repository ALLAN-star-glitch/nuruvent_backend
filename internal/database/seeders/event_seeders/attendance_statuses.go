// internal/database/seeders/event_seeders/attendance_statuses.go

package eventseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedAttendanceStatuses seeds the attendance statuses into the database
func SeedAttendanceStatuses(db *gorm.DB) error {
	log.Println("🌱 Seeding attendance statuses...")

	for _, as := range constants.AttendanceStatuses {
		// Check if status already exists
		var existing models.AttendanceStatus
		err := db.Where("slug = ?", as.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = as.Slug
			existing.Name = as.Name
			existing.DisplayName = as.DisplayName
			existing.Description = as.Description
			existing.Color = as.Color
			existing.Icon = as.Icon
			existing.SortOrder = as.SortOrder
			existing.CanIssueCertificate = as.CanIssueCertificate
			existing.IsFinal = as.IsFinal
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated attendance status: %s", as.Name)
			continue
		}

		// Create new
		attendanceStatus := &models.AttendanceStatus{
			Slug:                as.Slug,
			Name:                as.Name,
			DisplayName:         as.DisplayName,
			Description:         as.Description,
			Color:               as.Color,
			Icon:                as.Icon,
			SortOrder:           as.SortOrder,
			CanIssueCertificate: as.CanIssueCertificate,
			IsFinal:             as.IsFinal,
			IsActive:            true,
		}
		if err := db.Create(attendanceStatus).Error; err != nil {
			return err
		}
		log.Printf("✅ Created attendance status: %s", as.Name)
	}

	log.Printf("✅ Attendance statuses seeded: %d statuses", len(constants.AttendanceStatuses))
	return nil
}