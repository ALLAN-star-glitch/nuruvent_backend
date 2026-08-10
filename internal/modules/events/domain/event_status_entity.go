package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// EVENT STATUS - Domain Entity (Database Representation)
// ============================================================

// EventStatus is a domain entity stored in the database
type EventStatus struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int
	IsFinal     bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewEventStatus creates a new EventStatus entity
func NewEventStatus(slug, name, displayName, description, color, icon string, sortOrder int, isFinal bool) (*EventStatus, error) {
	if slug == "" {
		return nil, ErrInvalidEventStatus
	}
	if name == "" {
		return nil, ErrInvalidEventStatus
	}
	if displayName == "" {
		return nil, ErrInvalidEventStatus
	}

	now := time.Now()
	return &EventStatus{
		ID:          generateEventStatusID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Color:       color,
		Icon:        icon,
		SortOrder:   sortOrder,
		IsFinal:     isFinal,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewEventStatusFromRepository recreates EventStatus from repository data
func NewEventStatusFromRepository(
	id, slug, name, displayName, description, color, icon string,
	sortOrder int, isFinal, isActive bool,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *EventStatus {
	return &EventStatus{
		ID:          id,
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Color:       color,
		Icon:        icon,
		SortOrder:   sortOrder,
		IsFinal:     isFinal,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	}
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (e *EventStatus) IsValid() bool {
	return e.ID != ""
}

func (e *EventStatus) IsActiveType() bool {
	return e.IsActive
}

func (e *EventStatus) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidEventStatus
	}
	e.Name = name
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EventStatus) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidEventStatus
	}
	e.DisplayName = displayName
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EventStatus) UpdateDescription(description string) {
	e.Description = description
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) UpdateColor(color string) {
	e.Color = color
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) UpdateIcon(icon string) {
	e.Icon = icon
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) UpdateSortOrder(sortOrder int) {
	e.SortOrder = sortOrder
	e.UpdatedAt = time.Now()
}

func (e *EventStatus) UpdateIsFinal(isFinal bool) {
	e.IsFinal = isFinal
	e.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateEventStatusID() string {
	return uuid.New().String()
}