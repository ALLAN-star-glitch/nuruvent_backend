package domain

import "time"

// ============================================================
// INSTITUTION TYPE - Domain Entity (Database Representation)
// ============================================================

// InstitutionType is a domain entity stored in the database
type InstitutionType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewInstitutionType creates a new InstitutionType entity
func NewInstitutionType(slug, name, displayName, description, icon, color string, sortOrder int) (*InstitutionType, error) {
	if slug == "" {
		return nil, ErrInvalidInstitutionType
	}
	if name == "" {
		return nil, ErrInvalidInstitutionType
	}
	if displayName == "" {
		return nil, ErrInvalidInstitutionType
	}

	now := time.Now()
	return &InstitutionType{
		ID:          generateID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewInstitutionTypeFromRepository recreates InstitutionType from repository data
func NewInstitutionTypeFromRepository(
	id, slug, name, displayName, description, icon, color string,
	sortOrder int, isActive bool,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *InstitutionType {
	return &InstitutionType{
		ID:          id,
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	}
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// IsValid checks if the institution type has a valid ID
func (i *InstitutionType) IsValid() bool {
	return i.ID != ""
}

// IsActiveType checks if the institution type is active
func (i *InstitutionType) IsActiveType() bool {
	return i.IsActive
}

// Activate activates the institution type
func (i *InstitutionType) Activate() {
	i.IsActive = true
	i.UpdatedAt = time.Now()
}

// Deactivate deactivates the institution type
func (i *InstitutionType) Deactivate() {
	i.IsActive = false
	i.UpdatedAt = time.Now()
}

// UpdateName updates the name
func (i *InstitutionType) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidInstitutionType
	}
	i.Name = name
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateDisplayName updates the display name
func (i *InstitutionType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidInstitutionType
	}
	i.DisplayName = displayName
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (i *InstitutionType) UpdateDescription(description string) {
	i.Description = description
	i.UpdatedAt = time.Now()
}

// UpdateIcon updates the icon
func (i *InstitutionType) UpdateIcon(icon string) {
	i.Icon = icon
	i.UpdatedAt = time.Now()
}

// UpdateColor updates the color
func (i *InstitutionType) UpdateColor(color string) {
	i.Color = color
	i.UpdatedAt = time.Now()
}

// UpdateSortOrder updates the sort order
func (i *InstitutionType) UpdateSortOrder(sortOrder int) {
	i.SortOrder = sortOrder
	i.UpdatedAt = time.Now()
}
