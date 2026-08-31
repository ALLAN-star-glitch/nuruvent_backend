// internal/modules/events/infrastructure/seeder/certificate_type_seeder.go

package eventseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/postgres"
)

// SeedCertificateTypes seeds the certificate types from domain constants
func SeedCertificateTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding certificate types...")

	infos := domain.AllCertificateTypeInfos()

	for _, info := range infos {
		var existing postgres.CertificateTypeModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update existing
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated certificate type: %s", info.Name)
			continue
		}

		// Create new
		certificateType := &postgres.CertificateTypeModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			IsActive:    info.IsActive,
		}
		if err := db.Create(certificateType).Error; err != nil {
			return err
		}
		log.Printf("✅ Created certificate type: %s", info.Name)
	}

	log.Printf("✅ Certificate types seeded: %d types", len(infos))
	return nil
}