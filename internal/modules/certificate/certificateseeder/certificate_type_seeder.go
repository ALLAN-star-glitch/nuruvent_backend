// internal/modules/certificate/certificateseeder/certificate_type_seeder.go

package certificateseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/postgres"
)

// SeedCertificateTypes seeds the certificate types lookup table.
//
// Idempotent: upserts by slug.
func SeedCertificateTypes(db *gorm.DB) error {
	log.Println("🌱 Seeding certificate types...")

	infos := domain.AllCertificateTypeInfos()

	for _, info := range infos {
		var existing postgres.CertificateTypeModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.IsActive = info.IsActive
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated certificate type: %s", info.Name)
			continue
		}

		ct := &postgres.CertificateTypeModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			IsActive:    info.IsActive,
		}
		if err := db.Create(ct).Error; err != nil {
			return err
		}
		log.Printf("✅ Created certificate type: %s", info.Name)
	}

	log.Printf("✅ Certificate types seeded: %d types", len(infos))
	return nil
}