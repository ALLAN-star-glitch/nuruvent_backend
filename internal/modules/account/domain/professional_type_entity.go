package domain

import "time"

// ============================================================
// PROFESSIONAL TYPE - Domain Entity (Database Representation)
// ============================================================

// ProfessionalType is a domain entity stored in the database
type ProfessionalType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	CanHost     bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewProfessionalType creates a new ProfessionalType entity
func NewProfessionalType(slug, name, displayName, description, icon, color string, sortOrder int, canHost bool) (*ProfessionalType, error) {
	if slug == "" {
		return nil, ErrInvalidProfessionalType
	}
	if name == "" {
		return nil, ErrInvalidProfessionalType
	}
	if displayName == "" {
		return nil, ErrInvalidProfessionalType
	}

	now := time.Now()
	return &ProfessionalType{
		ID:          generateID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		CanHost:     canHost,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewProfessionalTypeFromRepository recreates ProfessionalType from repository data
func NewProfessionalTypeFromRepository(
	id, slug, name, displayName, description, icon, color string,
	sortOrder int, canHost, isActive bool,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *ProfessionalType {
	return &ProfessionalType{
		ID:          id,
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		CanHost:     canHost,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	}
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (p *ProfessionalType) IsValid() bool {
	return p.ID != ""
}

func (p *ProfessionalType) IsActiveType() bool {
	return p.IsActive
}

func (p *ProfessionalType) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidProfessionalType
	}
	p.Name = name
	p.UpdatedAt = time.Now()
	return nil
}

func (p *ProfessionalType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidProfessionalType
	}
	p.DisplayName = displayName
	p.UpdatedAt = time.Now()
	return nil
}

func (p *ProfessionalType) UpdateDescription(description string) {
	p.Description = description
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) UpdateIcon(icon string) {
	p.Icon = icon
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) UpdateColor(color string) {
	p.Color = color
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) UpdateSortOrder(sortOrder int) {
	p.SortOrder = sortOrder
	p.UpdatedAt = time.Now()
}

func (p *ProfessionalType) UpdateCanHost(canHost bool) {
	p.CanHost = canHost
	p.UpdatedAt = time.Now()
}