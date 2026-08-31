// internal/modules/events/domain/material_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// MATERIAL - Domain Entity (Database Representation)
// ============================================================

// Material represents an event material
type Material struct {
	ID           string
	EventID      string // Foreign key to Event
	Title        string
	Type         string // "pdf", "video", "link", "document"
	URL          string
	IsPreEvent   bool
	SortOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewMaterial creates a new Material entity
func NewMaterial(eventID, title, materialType, url string, isPreEvent bool, sortOrder int) (*Material, error) {
	if eventID == "" {
		return nil, ErrInvalidMaterial
	}
	if title == "" {
		return nil, ErrInvalidMaterial
	}
	if materialType == "" {
		return nil, ErrInvalidMaterial
	}
	if url == "" {
		return nil, ErrInvalidMaterial
	}

	// Validate material type
	if !IsMaterialTypeValid(MaterialTypeValue(materialType)) {
		return nil, ErrInvalidMaterial
	}

	now := time.Now()
	return &Material{
		ID:         generateMaterialID(),
		EventID:    eventID,
		Title:      title,
		Type:       materialType,
		URL:        url,
		IsPreEvent: isPreEvent,
		SortOrder:  sortOrder,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (m *Material) IsValid() bool {
	return m.ID != "" && m.EventID != "" && m.Title != "" && m.URL != ""
}

func (m *Material) UpdateTitle(title string) error {
	if title == "" {
		return ErrInvalidMaterial
	}
	m.Title = title
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Material) UpdateType(materialType string) error {
	if materialType == "" {
		return ErrInvalidMaterial
	}
	if !IsMaterialTypeValid(MaterialTypeValue(materialType)) {
		return ErrInvalidMaterial
	}
	m.Type = materialType
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Material) UpdateURL(url string) error {
	if url == "" {
		return ErrInvalidMaterial
	}
	m.URL = url
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Material) MarkAsPreEvent() {
	m.IsPreEvent = true
	m.UpdatedAt = time.Now()
}

func (m *Material) MarkAsPostEvent() {
	m.IsPreEvent = false
	m.UpdatedAt = time.Now()
}

func (m *Material) UpdateSortOrder(sortOrder int) {
	m.SortOrder = sortOrder
	m.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateMaterialID() string {
	return uuid.New().String()
}