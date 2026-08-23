package mediaseeder

import (
	"log"
	
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/postgres"
	"gorm.io/gorm"
)

// SeedMediaTypes seeds the media types from domain constants
func SeedMediaTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding media types...")

	infos := domain.AllMediaTypeInfos()

	for _, info := range infos {
		var existing postgres.MediaTypeModel
		// ✅ info.Slug is already a string, no need for .String()
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.Bucket = info.Bucket
			existing.Icon = info.Icon
			existing.SortOrder = info.SortOrder
			existing.MaxFileSize = info.MaxFileSize
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated media type: %s", info.Name)
			continue
		}

		// Create new
		mediaType := &postgres.MediaTypeModel{
			Slug:        info.Slug,        // ✅ Already a string
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Bucket:      info.Bucket,
			Icon:        info.Icon,
			SortOrder:   info.SortOrder,
			MaxFileSize: info.MaxFileSize,
			IsActive:    info.IsActive,
		}
		if err := db.Create(mediaType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created media type: %s", info.Name)
	}

	log.Printf("✅ Media types seeded: %d types", len(infos))
	return nil
}