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
	// EVENT TYPES & STATUSES
	// ============================================================

	GetEventTypes(ctx context.Context) ([]*domain.EventType, error)
	GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error)
}

// ============================================================
// COMMANDS - CREATE DRAFT
// ============================================================

type CreateDraftCommand struct {
	// Required - User Input
	Name string // User types this - raw input

	// System fields (auto-generated)
	CreatedBy     string
	OwnerType     string // "personal" or "institution"
	InstitutionID *string

	// Optional Fields
	Description      string
	ShortDescription string
	EventTypeID      string
	CategoryID       *string
	Tags             []string
	Language         string

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
	Visibility    string // "public", "private", "unlisted"
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

	ImageURL string 
}

// ============================================================
// COMMANDS - CREATE PUBLISHED EVENT
// ============================================================

type CreateEventCommand struct {
	// Required - User Input
	Name string // User types this - raw input

	// System fields (auto-generated)
	CreatedBy     string
	OwnerType     string // "personal" or "institution"
	InstitutionID *string
	EventTypeID   string
	Description   string

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

	// Optional Fields
	ShortDescription string
	CategoryID       *string
	Tags             []string
	Language         string

	// Access & Privacy
	Visibility    string // "public", "private", "unlisted"
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
	DisplayName      *string // If provided, update display name
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

	ImageURL string 
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
	UserID         string // Filter by user (creator)
	EventTypeID    string
	EventStatusID  string
	CategoryID     string // Filter by category
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	SortBy         string // Field to sort by (e.g., "created_at", "start_date", "name")
	SortOrder      string // "asc" or "desc"
	Visibility     string // ✅ ADDED: Filter by visibility ("public", "private", "unlisted")
}

type SearchFilters struct {
	InstitutionID  string // Filter by institution
	UserID         string // Filter by user (creator)
	EventTypeID    string
	CategoryID     string // Filter by category
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	Visibility     string // ✅ ADDED: Filter by visibility ("public", "private", "unlisted")
}

// ============================================================
// INPUT TYPES (Shared across commands)
// ============================================================

type ScheduleInput struct {
	ID           *string // For updates
	StartDate    string
	EndDate      *string
	StartTime    string
	EndTime      string
	Timezone     string
	SessionName  string
	SessionNumber int
	Location     string
	IsVirtual    bool
	ZoomLink     string
	MeetLink     string
	MaxAttendees *int
}

type RecurrenceInput struct {
	Pattern     string // "daily", "weekly", "monthly", "custom"
	Interval    int
	DaysOfWeek  []string
	DayOfMonth  *int
	WeekOfMonth *string
	EndsOn      *string
	Occurrences *int
}

type TicketInput struct {
	ID                 *string // For updates
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
	ID          *string // For updates
	Name        string
	Title       string
	Bio         string
	PhotoURL    string
	SocialLinks map[string]string
	IsKeynote   bool
	SortOrder   int
}

type MaterialInput struct {
	ID          *string // For updates
	Title       string
	MaterialTypeID string // "pdf", "video", "link", "document"
	URL         string
	Description string
	IsPreEvent  bool
	SortOrder   int
}

type SEOInput struct {
	MetaTitle       string
	MetaDescription string
	MetaKeywords    []string
	CanonicalURL    string
	Robots          string
	NoIndex         bool
	OGTitle         string
	OGDescription   string
	OGImageURL      string
	OGType          string
	TwitterCard     string
	TwitterTitle    string
	TwitterDescription string
	TwitterImageURL string
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