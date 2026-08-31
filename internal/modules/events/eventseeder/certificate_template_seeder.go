// internal/modules/events/infrastructure/seeder/certificate_template_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedCertificateTemplates seeds the certificate templates from domain constants
func SeedCertificateTemplates(db *gorm.DB) error {
	log.Println("🌱 Seeding certificate templates...")

	infos := domain.AllCertificateTemplateInfos()

	for _, info := range infos {
		var existing postgres.CertificateTemplateModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.PreviewURL = info.PreviewURL
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated certificate template: %s", info.Name)
			continue
		}

		// Create new
		certificateTemplate := &postgres.CertificateTemplateModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			PreviewURL:  info.PreviewURL,
			IsActive:    info.IsActive,
		}
		if err := db.Create(certificateTemplate).Error; err != nil {
			return err
		}
		log.Printf("✅ Created certificate template: %s", info.Name)
	}

	log.Printf("✅ Certificate templates seeded: %d templates", len(infos))
	return nil
}