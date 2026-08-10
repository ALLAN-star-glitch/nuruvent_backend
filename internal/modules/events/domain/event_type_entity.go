package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// EVENT TYPE - Domain Entity (Database Representation)
// ============================================================

// EventType is a domain entity stored in the database
type EventType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	SupportsCertificate bool
	MinDuration int
	MaxDuration int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewEventType creates a new EventType entity
func NewEventType(slug, name, displayName, description, icon, color string, sortOrder int, supportsCertificate bool, minDuration, maxDuration int) (*EventType, error) {
	if slug == "" {
		return nil, ErrInvalidEventType
	}
	if name == "" {
		return nil, ErrInvalidEventType
	}
	if displayName == "" {
		return nil, ErrInvalidEventType
	}

	now := time.Now()
	return &EventType{
		ID:                  generateEventTypeID(),
		Slug:                slug,
		Name:                name,
		DisplayName:         displayName,
		Description:         description,
		Icon:                icon,
		Color:               color,
		SortOrder:           sortOrder,
		SupportsCertificate: supportsCertificate,
		MinDuration:         minDuration,
		MaxDuration:         maxDuration,
		IsActive:            true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// NewEventTypeFromRepository recreates EventType from repository data
func NewEventTypeFromRepository(
	id, slug, name, displayName, description, icon, color string,
	sortOrder int, supportsCertificate bool, minDuration, maxDuration int, isActive bool,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *EventType {
	return &EventType{
		ID:                  id,
		Slug:                slug,
		Name:                name,
		DisplayName:         displayName,
		Description:         description,
		Icon:                icon,
		Color:               color,
		SortOrder:           sortOrder,
		SupportsCertificate: supportsCertificate,
		MinDuration:         minDuration,
		MaxDuration:         maxDuration,
		IsActive:            isActive,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		DeletedAt:           deletedAt,
	}
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (e *EventType) IsValid() bool {
	return e.ID != ""
}

func (e *EventType) IsActiveType() bool {
	return e.IsActive
}

func (e *EventType) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

func (e *EventType) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidEventType
	}
	e.Name = name
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EventType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidEventType
	}
	e.DisplayName = displayName
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EventType) UpdateDescription(description string) {
	e.Description = description
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateIcon(icon string) {
	e.Icon = icon
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateColor(color string) {
	e.Color = color
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateSortOrder(sortOrder int) {
	e.SortOrder = sortOrder
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateSupportsCertificate(supportsCertificate bool) {
	e.SupportsCertificate = supportsCertificate
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateMinDuration(minDuration int) {
	e.MinDuration = minDuration
	e.UpdatedAt = time.Now()
}

func (e *EventType) UpdateMaxDuration(maxDuration int) {
	e.MaxDuration = maxDuration
	e.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateEventTypeID() string {
	return uuid.New().String()
}