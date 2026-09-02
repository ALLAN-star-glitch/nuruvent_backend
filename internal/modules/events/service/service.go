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
	// READ - One method to rule them all
	// ============================================================

	// GetEventByID retrieves a single event by ID
	GetEventByID(ctx context.Context, id string) (*domain.Event, error)

	// GetEventBySlug retrieves a single event by slug
	GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error)

	// ListEvents is the PRIMARY query method - handles ALL list/filter scenarios
	// Use TeamFilter to filter by team (personal or institution)
	// Use EventTypeID, EventStatusID, CategoryID for additional filters
	// Use IncludeCreator to populate creator info
	// Use IncludeDeleted/OnlyDeleted for soft-delete filtering
	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error)

	// GetEventsByType retrieves events by event type slug
	GetEventsByType(ctx context.Context, eventTypeSlug string, page, pageSize int) ([]*domain.Event, int64, error)

	// GetUpcomingEvents is a convenience method for homepage/upcoming section
	GetUpcomingEvents(ctx context.Context, team domain.TeamFilter, limit int) ([]*domain.Event, error)

	// GetPastEvents is a convenience method for archives/past section
	GetPastEvents(ctx context.Context, team domain.TeamFilter, limit int) ([]*domain.Event, error)

	// SearchEvents for full-text search
	SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error)

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

	// DeleteEvents soft deletes multiple events by IDs
	DeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error)

	// PermanentlyDeleteEvents hard deletes multiple events by IDs
	PermanentlyDeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error)

	// RestoreEvents restores multiple soft-deleted events by IDs
	RestoreEvents(ctx context.Context, ids []string, restoredBy string) (*BulkRestoreResult, error)

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

	UploadEventImage(ctx context.Context, cmd UploadEventImageCommand) (*MediaInfo, error)
	UploadCertificateTemplate(ctx context.Context, cmd UploadCertificateCommand) (*MediaInfo, error)

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
	// EVENT TYPES & STATUSES (Value Objects)
	// ============================================================

	GetEventTypes(ctx context.Context) ([]*domain.EventType, error)
	GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error)
}

// ============================================================
// COMMANDS - CREATE DRAFT
// ============================================================

type CreateDraftCommand struct {
	// Basic Information
	Name             string
	Description      string
	ShortDescription string
	EventTypeID      string
	CategoryID       *string
	Tags             []string
	Language         string

	// Ownership - determines the scope
	CreatedBy     string  // User ID (required)
	OwnerType     string  // "personal" or "institution" (required)
	InstitutionID *string // Set if OwnerType is "institution"

	// Schedule
	Schedules   []ScheduleInput
	IsMultiDay  bool
	IsRecurring bool
	Recurrence  *RecurrenceInput

	// Venue
	IsVirtual          bool
	IsHybrid           bool
	InPersonLocation   string
	VirtualPlatform    string
	VirtualPlatformURL string
	ZoomLink           string
	MeetLink           string
	VenueName          string
	VenueAddress       string
	VenueCity          string
	VenueCountry       string

	// Tickets
	IsFree   bool
	Capacity *int
	Tickets  []TicketInput

	// Access & Privacy
	Visibility    string
	Password      *string
	InviteOnly    bool
	InvitedEmails []string

	// Monetization
	IsFeatured          bool
	CertificateEnabled  bool
	CertificatePrice    float64
	CertificateTemplateID *string

	// Speakers
	Speakers []SpeakerInput

	// Materials
	Materials []MaterialInput

	// SEO
	SEO *SEOInput
}

// ============================================================
// COMMANDS - CREATE PUBLISHED EVENT
// ============================================================

type CreateEventCommand struct {
	// Basic Information
	Name             string
	Description      string
	ShortDescription string
	EventTypeID      string
	CategoryID       *string
	Tags             []string
	Language         string

	// Ownership - determines the scope
	CreatedBy     string  // User ID (required)
	OwnerType     string  // "personal" or "institution" (required)
	InstitutionID *string // Set if OwnerType is "institution"

	// Schedule - Required for published events
	Schedules   []ScheduleInput
	IsMultiDay  bool
	IsRecurring bool
	Recurrence  *RecurrenceInput

	// Tickets - Required for published events
	IsFree   bool
	Capacity *int
	Waitlist bool
	Tickets  []TicketInput

	// Venue - Required for published events
	IsVirtual          bool
	IsHybrid           bool
	InPersonLocation   string
	VirtualPlatform    string
	VirtualPlatformURL string
	ZoomLink           string
	MeetLink           string
	VenueName          string
	VenueAddress       string
	VenueCity          string
	VenueCountry       string

	// Access & Privacy
	Visibility    string
	Password      *string
	InviteOnly    bool
	InvitedEmails []string

	// Monetization
	IsFeatured          bool
	CertificateEnabled  bool
	CertificatePrice    float64
	CertificateTemplateID *string

	// Speakers
	Speakers []SpeakerInput

	// Materials
	Materials []MaterialInput

	// SEO
	SEO *SEOInput
}

// ============================================================
// COMMANDS - UPDATE EVENT
// ============================================================

type UpdateEventCommand struct {
	ID        string
	UpdatedBy string

	// Basic Information (pointers for optional updates)
	Name             *string
	DisplayName      *string
	Description      *string
	ShortDescription *string
	EventTypeID      *string
	CategoryID       *string
	Tags             []string
	Language         *string

	// Schedule
	Schedules   []ScheduleInput
	IsMultiDay  *bool
	IsRecurring *bool
	Recurrence  *RecurrenceInput

	// Venue
	IsVirtual          *bool
	IsHybrid           *bool
	InPersonLocation   *string
	VirtualPlatform    *string
	VirtualPlatformURL *string
	ZoomLink           *string
	MeetLink           *string
	VenueName          *string
	VenueAddress       *string
	VenueCity          *string
	VenueCountry       *string

	// Tickets
	IsFree   *bool
	Capacity *int
	Waitlist *bool
	Tickets  []TicketInput

	// Access & Privacy
	Visibility    *string
	Password      *string
	InviteOnly    *bool
	InvitedEmails []string

	// Monetization
	IsFeatured          *bool
	CertificateEnabled  *bool
	CertificatePrice    *float64
	CertificateTemplateID *string

	// Speakers
	Speakers []SpeakerInput

	// Materials
	Materials []MaterialInput

	// SEO
	SEO *SEOInput
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

// ListEventsFilters provides comprehensive filtering for ListEvents
type ListEventsFilters struct {
	// TeamFilter filters events by team (personal or institution)
	Team domain.TeamFilter

	// UserID filters events by the creator (created_by)
	UserID string

	// EventTypeID filters events by their type
	EventTypeID string

	// EventStatusID filters events by their status
	EventStatusID string

	// CategoryID filters events by their category
	CategoryID string

	// IncludeDeleted controls whether soft-deleted events are included
	IncludeDeleted bool

	// OnlyDeleted controls whether ONLY soft-deleted events are returned
	OnlyDeleted bool

	// IncludeCreator controls whether creator user info is populated
	IncludeCreator bool

	// Limit controls the maximum number of events returned
	Limit int

	// Offset controls pagination offset
	Offset int

	// SortBy specifies the field to sort by
	SortBy string

	// SortOrder specifies the sort direction
	SortOrder string

	// Visibility filters events by their visibility level
	Visibility string
}

// SearchFilters provides filtering for the SearchEvents method
type SearchFilters struct {
	// TeamFilter filters search results by team
	Team domain.TeamFilter

	// UserID filters search results by creator
	UserID string

	// EventTypeID filters search results by event type
	EventTypeID string

	// CategoryID filters search results by category
	CategoryID string

	// IncludeDeleted controls whether soft-deleted events are included in search
	IncludeDeleted bool

	// OnlyDeleted controls whether ONLY soft-deleted events are returned in search
	OnlyDeleted bool

	// Limit controls the maximum number of search results
	Limit int

	// Offset controls pagination offset for search results
	Offset int

	// Visibility filters search results by visibility level
	Visibility string
}

// ============================================================
// INPUT TYPES (Shared across commands)
// ============================================================

type ScheduleInput struct {
	ID            *string
	StartDate     string
	EndDate       *string
	StartTime     string
	EndTime       string
	Timezone      string
	SessionName   string
	SessionNumber int
	Location      string
	IsVirtual     bool
	ZoomLink      string
	MeetLink      string
	MaxAttendees  *int
}

type RecurrenceInput struct {
	Pattern     string
	Interval    int
	DaysOfWeek  []string
	DayOfMonth  *int
	WeekOfMonth *string
	EndsOn      *string
	Occurrences *int
}

type TicketInput struct {
	ID                 *string
	TicketTypeID       string
	Name               string
	Description        string
	Price              float64
	Quantity           int
	MaxPerPerson       *int
	EarlyBirdDeadline  *string
	GroupMinAttendees  *int
	GroupDiscount      *float64
}

type SpeakerInput struct {
	ID          *string
	Name        string
	Title       string
	Bio         string
	PhotoURL    string
	SocialLinks map[string]string
	IsKeynote   bool
	SortOrder   int
}

type MaterialInput struct {
	ID             *string
	Title          string
	MaterialTypeID string
	URL            string
	Description    string
	IsPreEvent     bool
	SortOrder      int
}

type SEOInput struct {
	MetaTitle          string
	MetaDescription    string
	MetaKeywords       []string
	CanonicalURL       string
	Robots             string
	NoIndex            bool
	OGTitle            string
	OGDescription      string
	OGImageURL         string
	OGType             string
	TwitterCard        string
	TwitterTitle       string
	TwitterDescription string
	TwitterImageURL    string
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

// ============================================================
// MEDIA INFO (Response Type)
// ============================================================

type MediaInfo struct {
	ID          string
	URL         string
	Filename    string
	Size        int64
	ContentType string
	UploadedAt  string
}