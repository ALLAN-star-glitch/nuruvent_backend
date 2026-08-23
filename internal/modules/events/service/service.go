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

	// Create draft event (with optional image data)
	CreateDraft(ctx context.Context, cmd CreateDraftCommand) (*domain.Event, error)
	
	// Create published event (with optional image data)
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
	GetEventsByAccount(ctx context.Context, accountID string, page, pageSize int) ([]*domain.Event, int64, error)
	GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error)
	SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error)

	// ============================================================
	// READ - With Creator Info
	// ============================================================

	// GetUpcomingEventsWithCreator returns upcoming events with creator info populated
	GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*domain.Event, error)

	// GetEventBySlugWithCreator returns an event by slug with creator info populated
	GetEventBySlugWithCreator(ctx context.Context, slug string) (*domain.Event, error)

	// GetEventByIDWithCreator returns an event by ID with creator info populated
	GetEventByIDWithCreator(ctx context.Context, id string) (*domain.Event, error)

	// GetEventsByAccountWithCreator returns events for an account with creator info populated
	GetEventsByAccountWithCreator(ctx context.Context, accountID string, page, pageSize int) ([]*domain.Event, int64, error)

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
	DeleteEventsByAccount(ctx context.Context, accountID string, deletedBy string) (*BulkDeleteResult, error)
	PermanentlyDeleteEventsByAccount(ctx context.Context, accountID string, deletedBy string) (*BulkDeleteResult, error)

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
// COMMANDS - CREATE DRAFT (Pure Domain Types - NO JSON TAGS)
// ============================================================

type CreateDraftCommand struct {
	// Required
	AccountID string
	CreatedBy string

	// User input - what they type in the form
	Name string // ✅ This is the ONLY field from user

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
// COMMANDS - CREATE PUBLISHED EVENT (Pure Domain Types - NO JSON TAGS)
// ============================================================

type CreateEventCommand struct {
	// Required
	Name        string // ✅ User input - what they type
	AccountID   string
	CreatedBy   string
	Date        string
	Time        string
	Duration    int

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
// COMMANDS - UPDATE EVENT (Pure Domain Types - NO JSON TAGS)
// ============================================================

type UpdateEventCommand struct {
	ID       string
	UpdatedBy string

	// ✅ Use pointers for optional fields
	Name             *string // ✅ User input (optional)
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
// COMMANDS - DUPLICATE (Pure Domain Types - NO JSON TAGS)
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
// COMMANDS - MEDIA (Pure Domain Types - NO JSON TAGS)
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
// FILTERS (Pure Domain Types - NO JSON TAGS)
// ============================================================

type ListEventsFilters struct {
	AccountID      string
	EventTypeID    string
	EventStatusID  string
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
}

type SearchFilters struct {
	AccountID      string
	EventTypeID    string
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
}

// ============================================================
// BULK RESULT TYPES (NO JSON TAGS - These are for internal use)
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