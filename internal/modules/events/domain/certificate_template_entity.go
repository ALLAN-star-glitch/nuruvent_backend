// internal/modules/events/domain/certificate_template_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertificateTemplate is a domain entity stored in the database
type CertificateTemplate struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	PreviewURL  string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewCertificateTemplate creates a new CertificateTemplate entity
func NewCertificateTemplate(slug, name, displayName, description, previewURL string) (*CertificateTemplate, error) {
	if slug == "" {
		return nil, ErrInvalidCertificateTemplate
	}
	if name == "" {
		return nil, ErrInvalidCertificateTemplate
	}
	if displayName == "" {
		return nil, ErrInvalidCertificateTemplate
	}

	now := time.Now()
	return &CertificateTemplate{
		ID:          generateCertificateTemplateID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		PreviewURL:  previewURL,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Domain behaviors...
func (c *CertificateTemplate) IsValid() bool {
	return c.ID != "" && c.Slug != "" && c.Name != ""
}

func (c *CertificateTemplate) IsActiveType() bool {
	return c.IsActive
}

func (c *CertificateTemplate) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

func (c *CertificateTemplate) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

func (c *CertificateTemplate) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidCertificateTemplate
	}
	c.DisplayName = displayName
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CertificateTemplate) UpdateDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

func (c *CertificateTemplate) UpdatePreviewURL(previewURL string) {
	c.PreviewURL = previewURL
	c.UpdatedAt = time.Now()
}

func generateCertificateTemplateID() string {
	return uuid.New().String()
}