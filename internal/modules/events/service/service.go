// internal/modules/events/service/service.go

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

	CreateDraft(ctx context.Context, cmd CreateDraftCommand) (*domain.Event, error)
	CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error)

	// ============================================================
	// DUPLICATE
	// ============================================================

	DuplicateEvent(ctx context.Context, id string, cmd DuplicateEventCommand) (*domain.Event, error)
	BulkDuplicateEvents(ctx context.Context, ids []string, cmd BulkDuplicateCommand) (*BulkDuplicateResult, error)

	// ============================================================
	// READ - Basic (No Creator Info)
	// ============================================================

	GetEventByID(ctx context.Context, id string) (*domain.Event, error)
	GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error)
	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error)
	GetEventsByType(ctx context.Context, eventTypeSlug string, page, pageSize int) ([]*domain.Event, int64, error)
	GetEventsByInstitution(ctx context.Context, institutionID string, page, pageSize int) ([]*domain.Event, int64, error)
	GetEventsByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.Event, int64, error)
	GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error)

	// ============================================================
	// READ - With Creator Info
	// ============================================================

	GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*domain.Event, error)
	GetEventBySlugWithCreator(ctx context.Context, slug string) (*domain.Event, error)
	GetEventByIDWithCreator(ctx context.Context, id string) (*domain.Event, error)
	GetEventsByInstitutionWithCreator(ctx context.Context, institutionID string, page, pageSize int) ([]*domain.Event, int64, error)
	GetEventsByUserWithCreator(ctx context.Context, userID string, page, pageSize int) ([]*domain.Event, int64, error)

	// ============================================================
	// UPDATE
	// ============================================================

	UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error)

	// ============================================================
	// DELETE - Single
	// ============================================================

	DeleteEvent(ctx context.Context, id, deletedBy string) error
	PermanentlyDeleteEvent(ctx context.Context, id, deletedBy string) error
	RestoreEvent(ctx context.Context, id, restoredBy string) (*domain.Event, error)

	// ============================================================
	// DELETE - Bulk
	// ============================================================

	DeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error)
	PermanentlyDeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error)
	RestoreEvents(ctx context.Context, ids []string, restoredBy string) (*BulkRestoreResult, error)
	DeleteEventsByInstitution(ctx context.Context, institutionID string, deletedBy string) (*BulkDeleteResult, error)
	PermanentlyDeleteEventsByInstitution(ctx context.Context, institutionID string, deletedBy string) (*BulkDeleteResult, error)

	// ============================================================
	// STATUS - Single
	// ============================================================

	PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error)
	CancelEvent(ctx context.Context, id, cancelledBy string) (*domain.Event, error)
	CompleteEvent(ctx context.Context, id string) (*domain.Event, error)

	// ============================================================
	// STATUS - Bulk
	// ============================================================

	BulkPublishEvents(ctx context.Context, ids []string, publishedBy string) (*BulkStatusResult, error)
	BulkCancelEvents(ctx context.Context, ids []string, cancelledBy string) (*BulkStatusResult, error)
	BulkCompleteEvents(ctx context.Context, ids []string) (*BulkStatusResult, error)

	// ============================================================
	// MEDIA - Upload
	// ============================================================

	UploadEventImage(ctx context.Context, cmd UploadEventImageCommand) (*domain.MediaInfo, error)
	UploadCertificateTemplate(ctx context.Context, cmd UploadCertificateCommand) (*domain.MediaInfo, error)

	// ============================================================
	// MEDIA - Delete Single
	// ============================================================

	DeleteEventImage(ctx context.Context, eventID string, deletedBy string) error
	DeleteEventCertificate(ctx context.Context, eventID string, deletedBy string) error
	DeleteAllEventMedia(ctx context.Context, eventID string, deletedBy string) error

	// ============================================================
	// MEDIA - Delete Bulk
	// ============================================================

	BulkDeleteEventMedia(ctx context.Context, eventIDs []string, deletedBy string) (*BulkDeleteResult, error)

	// ============================================================
	// EVENT TYPES & STATUSES
	// ============================================================

	GetEventTypes(ctx context.Context) ([]*domain.EventType, error)
	GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error)
}

// ============================================================
// COMMANDS - CREATE DRAFT
// ============================================================

type CreateDraftCommand struct {
	// Required
	InstitutionID *string // ✅ Can be NULL for personal events
	CreatedBy     string  // ✅ The user (human) creating the event
	TeamType      string  // ✅ "personal" or "institution"

	// User input - what they type in the form
	Name string

	// Optional
	Description      string
	EventTypeID      string
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
	IsFeatured       bool
	IsPrivate        bool

	ImageData   []byte
	ImageName   string
	ContentType string
}

// ============================================================
// COMMANDS - CREATE PUBLISHED EVENT
// ============================================================

type CreateEventCommand struct {
	// Required
	Name          string  // User input - what they type
	InstitutionID *string // ✅ Can be NULL for personal events
	CreatedBy     string  // ✅ The user (human) creating the event
	TeamType      string  // ✅ "personal" or "institution"
	Date          string
	Time          string
	Duration      int

	// Optional
	Description      string
	EventTypeID      string
	Price            float64
	CertificatePrice float64
	Location         string
	IsVirtual        bool
	ZoomLink         string
	MeetLink         string
	MaxAttendees     int
	IsFeatured       bool
	IsPrivate        bool

	ImageData   []byte
	ImageName   string
	ContentType string
}

// ============================================================
// COMMANDS - UPDATE EVENT
// ============================================================

type UpdateEventCommand struct {
	ID        string
	UpdatedBy string // ✅ The user (human) updating the event

	// Use pointers for optional fields
	Name             *string
	DisplayName      *string
	Description      *string
	EventTypeID      *string
	EventStatusID    *string
	Date             *string
	Time             *string
	Duration         *int
	Price            *float64
	CertificatePrice *float64
	Location         *string
	IsVirtual        *bool
	ZoomLink         *string
	MeetLink         *string
	MaxAttendees     *int
	IsFeatured       *bool
	IsPrivate        *bool
}

// ============================================================
// COMMANDS - DUPLICATE
// ============================================================

type DuplicateEventCommand struct {
	Name    string
	Date    string
	IsDraft bool
}

type BulkDuplicateCommand struct {
	NamePrefix     string
	DateOffsetDays int
	IsDraft        bool
}

// ============================================================
// COMMANDS - MEDIA
// ============================================================

type UploadEventImageCommand struct {
	EventID     string
	ImageData   []byte
	ImageName   string
	ContentType string
	UploadedBy  string
}

type UploadCertificateCommand struct {
	EventID         string
	CertificateData []byte
	CertificateName string
	ContentType     string
	UploadedBy      string
}

// ============================================================
// FILTERS
// ============================================================

type ListEventsFilters struct {
	InstitutionID  string // Filter by institution
	UserID         string // ✅ NEW: Filter by user (creator)
	EventTypeID    string
	EventStatusID  string
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
}

type SearchFilters struct {
	InstitutionID  string // Filter by institution
	UserID         string // ✅ NEW: Filter by user (creator)
	EventTypeID    string
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
}

// ============================================================
// BULK RESULT TYPES
// ============================================================

type BulkDeleteResult struct {
	DeletedCount int
	FailedIDs    []string
	Errors       []string
}

type BulkRestoreResult struct {
	RestoredCount int
	FailedIDs     []string
	Errors        []string
}

type BulkStatusResult struct {
	ProcessedCount int
	FailedIDs      []string
	Errors         []string
}

type BulkDuplicateResult struct {
	DuplicatedCount int
	CreatedEvents   []DuplicatedEvent
	FailedIDs       []string
	Errors          []string
}

type DuplicatedEvent struct {
	ID   string
	Name string
	Slug string
}