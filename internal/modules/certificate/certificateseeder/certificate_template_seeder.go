// internal/modules/certificate/certificateseeder/certificate_template_seeder.go

package certificateseeder

import (
	"log"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/postgres"
)

// SeedCertificateTemplates seeds the Nuruvent-provided system
// certificate templates.
//
// Idempotent: upserts by slug. On existing rows, only the metadata
// columns are updated — design and is_system are preserved so an
// accidentally-modified system template doesn't get silently reset.
//
// SystemSlug is always kept in sync with the seeder's slug. This is
// what makes the tier invariant hold: system templates derive their
// price and design restrictions from system_slug, and the unique
// index on system_slug enforces one template per tier.
func SeedCertificateTemplates(db *gorm.DB) error {
	log.Println("🌱 Seeding certificate templates...")

	infos := domain.AllCertificateTemplateInfos()

	for _, info := range infos {
		var existing postgres.CertificateTemplateModel
		err := db.Where("slug = ?", info.Slug).First(&existing).Error

		if err == nil {
			// Update metadata only — leave design, is_system, is_default intact.
			existing.Name = info.Name
			existing.DisplayName = info.DisplayName
			existing.Description = info.Description
			existing.PreviewURL = info.PreviewURL
			existing.IsActive = info.IsActive
			// Keep system_slug aligned with the canonical slug so the
			// unique-per-tier invariant never drifts.
			existing.SystemSlug = info.Slug
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("✅ Updated certificate template: %s", info.Name)
			continue
		}

		// Create new — explicitly mark as a system template with its tier slug.
		tmpl := &postgres.CertificateTemplateModel{
			Slug:        info.Slug,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			PreviewURL:  info.PreviewURL,
			IsActive:    info.IsActive,
			IsSystem:    true,
			SystemSlug:  info.Slug,
		}
		if err := db.Create(tmpl).Error; err != nil {
			return err
		}
		log.Printf("✅ Created certificate template: %s", info.Name)
	}

	log.Printf("✅ Certificate templates seeded: %d templates", len(infos))
	return nil
}