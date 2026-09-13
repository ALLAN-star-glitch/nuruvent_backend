// internal/modules/events/domain/event_format_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// EVENT FORMAT - Domain Entity (Database Representation)
// ============================================================

// EventFormat is a domain entity stored in the database
type EventFormat struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewEventFormat creates a new EventFormat entity
func NewEventFormat(slug, name, displayName, description, icon string) (*EventFormat, error) {
	if slug == "" {
		return nil, ErrInvalidEventFormat
	}
	if name == "" {
		return nil, ErrInvalidEventFormat
	}
	if displayName == "" {
		return nil, ErrInvalidEventFormat
	}

	now := time.Now()
	return &EventFormat{
		ID:          generateEventFormatID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// IsValid checks if the event format entity is valid
func (e *EventFormat) IsValid() bool {
	return e.ID != "" && e.Slug != "" && e.Name != ""
}

// IsActiveType returns whether the format is active
func (e *EventFormat) IsActiveType() bool {
	return e.IsActive
}

// Activate activates the event format
func (e *EventFormat) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

// Deactivate deactivates the event format
func (e *EventFormat) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

// UpdateName updates the internal name
func (e *EventFormat) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidEventFormat
	}
	e.Name = name
	e.UpdatedAt = time.Now()
	return nil
}

// UpdateDisplayName updates the display name
func (e *EventFormat) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidEventFormat
	}
	e.DisplayName = displayName
	e.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (e *EventFormat) UpdateDescription(description string) {
	e.Description = description
	e.UpdatedAt = time.Now()
}

// UpdateIcon updates the icon
func (e *EventFormat) UpdateIcon(icon string) {
	e.Icon = icon
	e.UpdatedAt = time.Now()
}

// UpdateSlug updates the slug (only if not in use - validation should be done at application level)
func (e *EventFormat) UpdateSlug(slug string) error {
	if slug == "" {
		return ErrInvalidEventFormat
	}
	e.Slug = slug
	e.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// HELPERS
// ============================================================

func generateEventFormatID() string {
	return uuid.New().String()
}