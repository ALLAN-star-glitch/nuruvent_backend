package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Institution struct {
	ID                string
	Slug              string
	Name              string
	DisplayName       string
	Email             string
	Phone             string
	InstitutionTypeID string
	Description       string
	Logo              string
	Website           string
	Address           string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

func NewInstitution(name, email, phone, institutionTypeID string) (*Institution, error) {
	if name == "" {
		return nil, errors.New("institution name is required")
	}
	if email == "" {
		return nil, errors.New("institution email is required")
	}
	if phone == "" { // ✅ Add phone validation
		return nil, errors.New("institution phone is required")
	}
	if institutionTypeID == "" {
		return nil, errors.New("institution type is required")
	}

	now := time.Now()
	return &Institution{
		ID:                uuid.New().String(),
		Slug:              slugify(name),
		Name:              name,
		DisplayName:       name,
		Email:             email,
		Phone:             phone,
		InstitutionTypeID: institutionTypeID,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (i *Institution) Deactivate() {
	i.IsActive = false
	i.UpdatedAt = time.Now()
}

func (i *Institution) Activate() {
	i.IsActive = true
	i.UpdatedAt = time.Now()
}

func slugifyinstitution(name string) string {
	return "inst-" + uuid.New().String()[:8]
}