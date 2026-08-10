package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// ============================================================
	// CREATE
	// ============================================================

	CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error)
	CreateEventWithImage(ctx context.Context, cmd CreateEventWithImageCommand) (*domain.Event, error)

	// ============================================================
	// READ
	// ============================================================

	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error)
	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error)
	GetEventsByType(ctx context.Context, eventTypeSlug string, page, pageSize int) ([]*domain.Event, int64, error)
	GetEventsByAccount(ctx context.Context, accountID string, page, pageSize int) ([]*domain.Event, int64, error)
	GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error)

	// ============================================================
	// UPDATE
	// ============================================================

	UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error)

	// ============================================================
	// DELETE
	// ============================================================

	DeleteEvent(ctx context.Context, id, deletedBy string) error

	// ============================================================
	// STATUS
	// ============================================================

	PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error)
	CancelEvent(ctx context.Context, id, cancelledBy string) (*domain.Event, error)
	CompleteEvent(ctx context.Context, id string) (*domain.Event, error)

	// ============================================================
	// MEDIA
	// ============================================================

	UploadEventImage(ctx context.Context, cmd UploadEventImageCommand) (*domain.MediaInfo, error)
	UploadCertificateTemplate(ctx context.Context, cmd UploadCertificateCommand) (*domain.MediaInfo, error)

	// ============================================================
	// EVENT TYPES & STATUSES (Value objects)
	// ============================================================

	GetEventTypes(ctx context.Context) ([]*domain.EventType, error)
	GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error)

}

// ============================================================
// COMMANDS
// ============================================================

type CreateEventCommand struct {
	Name             string
	Description      string
	EventTypeID      string
	AccountID        string
	CreatedBy        string
	Date             string
	Time             string
	Duration         int
	Price            float64
	CertificatePrice float64
	Location         string
	IsVirtual        bool
	ZoomLink         string
	MeetLink         string
	MaxAttendees     int
}

type CreateEventWithImageCommand struct {
	CreateEventCommand
	ImageFile   interface{}
	ImageHeader interface{}
}

type UpdateEventCommand struct {
	ID               string
	Name             string
	Description      string
	EventTypeID      string
	EventStatusID    string
	Date             string
	Time             string
	Duration         int
	Price            float64
	CertificatePrice float64
	Location         string
	IsVirtual        bool
	ZoomLink         string
	MeetLink         string
	MaxAttendees     int
	UpdatedBy        string
}

type UploadEventImageCommand struct {
	EventID     string
	ImageFile   interface{}
	ImageHeader interface{}
	UploadedBy  string
}

type UploadCertificateCommand struct {
	EventID         string
	CertificateFile   interface{}
	CertificateHeader interface{}
	UploadedBy      string
}

// ============================================================
// FILTERS
// ============================================================

type ListEventsFilters struct {
	AccountID     string
	EventTypeID   string
	EventStatusID string
	Limit         int
	Offset        int
}

type SearchFilters struct {
	AccountID   string
	EventTypeID string
	Limit       int
	Offset      int
}