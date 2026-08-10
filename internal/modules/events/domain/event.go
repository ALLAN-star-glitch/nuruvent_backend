package domain

import (
	"errors"
	"time"
	"github.com/google/uuid"
)

type Event struct {
	ID               string
	Slug             string
	Name             string
	DisplayName      string
	Description      string
	EventTypeID      string
	EventStatusID    string
	ImageURL         string
	ThumbnailURL     string
	Date             time.Time
	Time             string
	Duration         int
	Price            float64
	CertificatePrice float64
	Location         string
	IsVirtual        bool
	ZoomLink         string
	MeetLink         string
	MaxAttendees     int
	CurrentAttendees int
	AccountID        string   // ← Changed from InstitutionID
	CreatedBy        string
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewEvent(name, description, eventTypeID, accountID, createdBy string) (*Event, error) {
	if name == "" {
		return nil, errors.New("event name is required")
	}
	if eventTypeID == "" {
		return nil, errors.New("event type is required")
	}
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	now := time.Now()
	return &Event{
		ID:               uuid.New().String(),
		Slug:             generateSlug(name),
		Name:             name,
		DisplayName:      name,
		Description:      description,
		EventTypeID:      eventTypeID,
		AccountID:        accountID,
		CreatedBy:        createdBy,
		IsActive:         true,
		CurrentAttendees: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func generateSlug(name string) string {
	return "event-" + uuid.New().String()[:8]
}

func (e *Event) CanPublish() bool {
	return e.EventStatusID != "" && e.Date.After(time.Now())
}

func (e *Event) IsFull() bool {
	return e.MaxAttendees > 0 && e.CurrentAttendees >= e.MaxAttendees
}

func (e *Event) Publish() error {
	if e.EventStatusID == "" {
		return errors.New("cannot publish event without status")
	}
	if e.Date.Before(time.Now()) {
		return errors.New("cannot publish past event")
	}
	return nil
}

func (e *Event) Cancel() error {
	if e.EventStatusID == "" {
		return errors.New("cannot cancel event without status")
	}
	return nil
}

func (e *Event) Complete() error {
	if e.EventStatusID == "" {
		return errors.New("cannot complete event without status")
	}
	if e.Date.After(time.Now()) {
		return errors.New("cannot complete future event")
	}
	return nil
}