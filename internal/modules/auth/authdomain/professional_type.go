// internal/modules/auth/authdomain/professional_type.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ProfessionalType entity
type ProfessionalType struct {
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

// NewProfessionalType creates a validated professional type
func NewProfessionalType(slug, name, displayName, description, icon, color string, sortOrder int) (*ProfessionalType, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}

	now := time.Now()
	return &ProfessionalType{
		ID:          uuid.New().String(),
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

// Predefined professional types
func NewTrainerType() (*ProfessionalType, error) {
	return NewProfessionalType(
		"professional_type_trainer",
		"trainer",
		"Trainer",
		"Professional trainer who conducts training sessions",
		"graduation-cap",
		"#4F46E5",
		1,
	)
}

func NewCoachType() (*ProfessionalType, error) {
	return NewProfessionalType(
		"professional_type_coach",
		"coach",
		"Coach",
		"Professional coach providing guidance and mentorship",
		"user-tie",
		"#7C3AED",
		2,
	)
}

func NewConsultantType() (*ProfessionalType, error) {
	return NewProfessionalType(
		"professional_type_consultant",
		"consultant",
		"Consultant",
		"Professional consultant providing expert advice",
		"briefcase",
		"#0EA5E9",
		3,
	)
}

func NewFreelancerType() (*ProfessionalType, error) {
	return NewProfessionalType(
		"professional_type_freelancer",
		"freelancer",
		"Freelancer",
		"Independent professional offering services",
		"laptop",
		"#F59E0B",
		4,
	)
}

// Behaviors
func (pt *ProfessionalType) Deactivate() {
	pt.IsActive = false
	pt.UpdatedAt = time.Now()
}

func (pt *ProfessionalType) Activate() {
	pt.IsActive = true
	pt.UpdatedAt = time.Now()
}

func (pt *ProfessionalType) IsActiveType() bool {
	return pt.IsActive
}

func (pt *ProfessionalType) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	pt.Name = name
	pt.UpdatedAt = time.Now()
	return nil
}

func (pt *ProfessionalType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("display name cannot be empty")
	}
	pt.DisplayName = displayName
	pt.UpdatedAt = time.Now()
	return nil
}

func (pt *ProfessionalType) UpdateDescription(description string) {
	pt.Description = description
	pt.UpdatedAt = time.Now()
}

func (pt *ProfessionalType) UpdateIcon(icon string) {
	pt.Icon = icon
	pt.UpdatedAt = time.Now()
}

func (pt *ProfessionalType) UpdateColor(color string) {
	pt.Color = color
	pt.UpdatedAt = time.Now()
}

func (pt *ProfessionalType) UpdateSortOrder(order int) {
	pt.SortOrder = order
	pt.UpdatedAt = time.Now()
}

// Soft delete
func (pt *ProfessionalType) Delete() {
	now := time.Now()
	pt.DeletedAt = &now
	pt.IsActive = false
	pt.UpdatedAt = now
}