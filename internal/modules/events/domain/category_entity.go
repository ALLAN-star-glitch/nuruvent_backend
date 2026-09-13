// internal/modules/events/domain/category_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// CATEGORY - Domain Entity (Database Representation)
// ============================================================

// Category is a domain entity stored in the database
type Category struct {
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

// NewCategory creates a new Category entity
func NewCategory(slug, name, displayName, description, icon, color string, sortOrder int) (*Category, error) {
	if slug == "" {
		return nil, ErrInvalidCategory
	}
	if name == "" {
		return nil, ErrInvalidCategory
	}
	if displayName == "" {
		return nil, ErrInvalidCategory
	}

	now := time.Now()
	return &Category{
		ID:          generateCategoryID(),
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

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (c *Category) IsValid() bool {
	return c.ID != "" && c.Slug != "" && c.Name != ""
}

func (c *Category) IsActiveType() bool {
	return c.IsActive
}

func (c *Category) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

func (c *Category) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

func (c *Category) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidCategory
	}
	c.Name = name
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Category) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidCategory
	}
	c.DisplayName = displayName
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Category) UpdateDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

func (c *Category) UpdateIcon(icon string) {
	c.Icon = icon
	c.UpdatedAt = time.Now()
}

func (c *Category) UpdateColor(color string) {
	c.Color = color
	c.UpdatedAt = time.Now()
}

func (c *Category) UpdateSortOrder(sortOrder int) {
	c.SortOrder = sortOrder
	c.UpdatedAt = time.Now()
}

func (c *Category) UpdateSlug(slug string) error {
	if slug == "" {
		return ErrInvalidCategory
	}
	c.Slug = slug
	c.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// HELPERS
// ============================================================

func generateCategoryID() string {
	return uuid.New().String()
}