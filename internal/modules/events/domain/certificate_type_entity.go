// internal/modules/events/domain/certificate_type_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertificateType is a domain entity stored in the database
type CertificateType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewCertificateType creates a new CertificateType entity
func NewCertificateType(slug, name, displayName string) (*CertificateType, error) {
	if slug == "" {
		return nil, ErrInvalidCertificateType
	}
	if name == "" {
		return nil, ErrInvalidCertificateType
	}
	if displayName == "" {
		return nil, ErrInvalidCertificateType
	}

	now := time.Now()
	return &CertificateType{
		ID:          generateCertificateTypeID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Domain behaviors...
func (c *CertificateType) IsValid() bool {
	return c.ID != "" && c.Slug != "" && c.Name != ""
}

func (c *CertificateType) IsActiveType() bool {
	return c.IsActive
}

func (c *CertificateType) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

func (c *CertificateType) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

func (c *CertificateType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidCertificateType
	}
	c.DisplayName = displayName
	c.UpdatedAt = time.Now()
	return nil
}

func generateCertificateTypeID() string {
	return uuid.New().String()
}