// internal/database/seeders/media_seeders/media_types.go

package mediaseeder

import (
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"gorm.io/gorm"
)

// SeedMediaTypes seeds the media types into the database
func SeedMediaTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding media types...")

	for _, mt := range constants.MediaTypes {
		// Check if type already exists
		var existing models.MediaType
		err := db.Where("slug = ?", mt.Slug).First(&existing).Error
		if err == nil {
			// Update existing
			existing.Slug = mt.Slug
			existing.Name = mt.Name
			existing.DisplayName = mt.DisplayName
			existing.Description = mt.Description
			existing.Bucket = mt.Bucket
			existing.Icon = mt.Icon
			existing.SortOrder = mt.SortOrder
			existing.MaxFileSize = mt.MaxFileSize
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated media type: %s", mt.Name)
			continue
		}

		// Create new
		mediaType := &models.MediaType{
			Slug:        mt.Slug,
			Name:        mt.Name,
			DisplayName: mt.DisplayName,
			Description: mt.Description,
			Bucket:      mt.Bucket,
			Icon:        mt.Icon,
			SortOrder:   mt.SortOrder,
			MaxFileSize: mt.MaxFileSize,
			IsActive:    true,
		}
		if err := db.Create(mediaType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created media type: %s", mt.Name)
	}

	log.Printf("✅ Media types seeded: %d types", len(constants.MediaTypes))
	return nil
}