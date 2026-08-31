// internal/modules/events/handler/dtos.go

package eventhandler

// ============================================================
// REQUEST DTOS
// ============================================================

// CreateDraftRequest - All fields optional for drafts (application/json)
// NOTE: owner_type and institution_id are NOT in the request body
// They are derived from the URL by the handler
type CreateDraftRequest struct {
	// Basic Information
	Name             string   `json:"name"`
	DisplayName      string   `json:"display_name"`
	Description      string   `json:"description"`
	ShortDescription string   `json:"short_description"`
	EventTypeID      string   `json:"event_type_id"`
	CategoryID       string   `json:"category_id"`
	Tags             []string `json:"tags"`
	Language         string   `json:"language"`

	// ❌ REMOVED: OwnerType     string `json:"owner_type"`
	// ❌ REMOVED: InstitutionID string `json:"institution_id"`

	// Schedule
	Schedules   []ScheduleRequest  `json:"schedules"`
	IsMultiDay  bool               `json:"is_multi_day"`
	IsRecurring bool               `json:"is_recurring"`
	Recurrence  *RecurrenceRequest `json:"recurrence"`

	// Venue
	IsVirtual          bool   `json:"is_virtual"`
	IsHybrid           bool   `json:"is_hybrid"`
	InPersonLocation   string `json:"in_person_location"`
	VirtualPlatform    string `json:"virtual_platform"`
	VirtualPlatformURL string `json:"virtual_platform_url"`
	ZoomLink           string `json:"zoom_link"`
	MeetLink           string `json:"meet_link"`
	VenueName          string `json:"venue_name"`
	VenueAddress       string `json:"venue_address"`
	VenueCity          string `json:"venue_city"`
	VenueCountry       string `json:"venue_country"`

	// Tickets
	IsFree   bool            `json:"is_free"`
	Capacity int             `json:"capacity"`
	Tickets  []TicketRequest `json:"tickets"`

	// Access & Privacy
	Visibility    string   `json:"visibility"` // "public", "private", "unlisted"
	Password      string   `json:"password"`
	InviteOnly    bool     `json:"invite_only"`
	InvitedEmails []string `json:"invited_emails"`

	// Monetization
	IsFeatured          bool    `json:"is_featured"`
	CertificateEnabled  bool    `json:"certificate_enabled"`
	CertificatePrice    float64 `json:"certificate_price"`
	CertificateTemplateID string `json:"certificate_template_id"`

	// Speakers
	Speakers []SpeakerRequest `json:"speakers"`

	// Materials
	Materials []MaterialRequest `json:"materials"`

	// SEO
	SEO *SEORequest `json:"seo"`

	// ❌ REMOVED: Image URL (handled by separate media endpoints)
	// ImageURL string `json:"image_url"`
}

// CreateEventRequest - All fields required for published events (application/json)
// NOTE: owner_type and institution_id are NOT in the request body
// They are derived from the URL by the handler
type CreateEventRequest struct {
	// Basic Information - Required
	Name             string   `json:"name" binding:"required"`
	DisplayName      string   `json:"display_name"`
	Description      string   `json:"description" binding:"required"`
	ShortDescription string   `json:"short_description"`
	EventTypeID      string   `json:"event_type_id" binding:"required"`
	CategoryID       string   `json:"category_id"`
	Tags             []string `json:"tags"`
	Language         string   `json:"language"`

	// ❌ REMOVED: OwnerType     string `json:"owner_type" binding:"required,oneof=personal institution"`
	// ❌ REMOVED: InstitutionID string `json:"institution_id"`

	// Schedule - Required
	Schedules   []ScheduleRequest  `json:"schedules" binding:"required,min=1"`
	IsMultiDay  bool               `json:"is_multi_day"`
	IsRecurring bool               `json:"is_recurring"`
	Recurrence  *RecurrenceRequest `json:"recurrence"`

	// Venue - Required
	IsVirtual          bool   `json:"is_virtual"`
	IsHybrid           bool   `json:"is_hybrid"`
	InPersonLocation   string `json:"in_person_location"`
	VirtualPlatform    string `json:"virtual_platform"`
	VirtualPlatformURL string `json:"virtual_platform_url"`
	ZoomLink           string `json:"zoom_link"`
	MeetLink           string `json:"meet_link"`
	VenueName          string `json:"venue_name"`
	VenueAddress       string `json:"venue_address"`
	VenueCity          string `json:"venue_city"`
	VenueCountry       string `json:"venue_country"`

	// Tickets - Required
	IsFree   bool            `json:"is_free"`
	Capacity int             `json:"capacity"`
	Waitlist bool            `json:"waitlist_enabled"`
	Tickets  []TicketRequest `json:"tickets" binding:"required,min=1"`

	// Access & Privacy - Required
	Visibility    string   `json:"visibility" binding:"required,oneof=public private unlisted"`
	Password      string   `json:"password"`
	InviteOnly    bool     `json:"invite_only"`
	InvitedEmails []string `json:"invited_emails"`

	// Monetization
	IsFeatured          bool    `json:"is_featured"`
	CertificateEnabled  bool    `json:"certificate_enabled"`
	CertificatePrice    float64 `json:"certificate_price"`
	CertificateTemplateID string `json:"certificate_template_id"`

	// Speakers
	Speakers []SpeakerRequest `json:"speakers"`

	// Materials
	Materials []MaterialRequest `json:"materials"`

	// SEO
	SEO *SEORequest `json:"seo"`

	// ❌ REMOVED: Image URL (handled by separate media endpoints)
	// ImageURL string `json:"image_url"`
}

// UpdateEventRequest - All fields optional for updates (application/json)
type UpdateEventRequest struct {
	// Basic Information
	Name             *string   `json:"name,omitempty"`
	DisplayName      *string   `json:"display_name,omitempty"`
	Description      *string   `json:"description,omitempty"`
	ShortDescription *string   `json:"short_description,omitempty"`
	EventTypeID      *string   `json:"event_type_id,omitempty"`
	CategoryID       *string   `json:"category_id,omitempty"`
	Tags             []string  `json:"tags,omitempty"`
	Language         *string   `json:"language,omitempty"`

	// Schedule
	Schedules   []ScheduleRequest  `json:"schedules,omitempty"`
	IsMultiDay  *bool              `json:"is_multi_day,omitempty"`
	IsRecurring *bool              `json:"is_recurring,omitempty"`
	Recurrence  *RecurrenceRequest `json:"recurrence,omitempty"`

	// Venue
	IsVirtual          *bool   `json:"is_virtual,omitempty"`
	IsHybrid           *bool   `json:"is_hybrid,omitempty"`
	InPersonLocation   *string `json:"in_person_location,omitempty"`
	VirtualPlatform    *string `json:"virtual_platform,omitempty"`
	VirtualPlatformURL *string `json:"virtual_platform_url,omitempty"`
	ZoomLink           *string `json:"zoom_link,omitempty"`
	MeetLink           *string `json:"meet_link,omitempty"`
	VenueName          *string `json:"venue_name,omitempty"`
	VenueAddress       *string `json:"venue_address,omitempty"`
	VenueCity          *string `json:"venue_city,omitempty"`
	VenueCountry       *string `json:"venue_country,omitempty"`

	// Tickets
	IsFree   *bool           `json:"is_free,omitempty"`
	Capacity *int            `json:"capacity,omitempty"`
	Waitlist *bool           `json:"waitlist_enabled,omitempty"`
	Tickets  []TicketRequest `json:"tickets,omitempty"`

	// Access & Privacy
	Visibility    *string  `json:"visibility,omitempty"`
	Password      *string  `json:"password,omitempty"`
	InviteOnly    *bool    `json:"invite_only,omitempty"`
	InvitedEmails []string `json:"invited_emails,omitempty"`

	// Monetization
	IsFeatured          *bool    `json:"is_featured,omitempty"`
	CertificateEnabled  *bool    `json:"certificate_enabled,omitempty"`
	CertificatePrice    *float64 `json:"certificate_price,omitempty"`
	CertificateTemplateID *string `json:"certificate_template_id,omitempty"`

	// Speakers
	Speakers []SpeakerRequest `json:"speakers,omitempty"`

	// Materials
	Materials []MaterialRequest `json:"materials,omitempty"`

	// SEO
	SEO *SEORequest `json:"seo,omitempty"`

	// ❌ REMOVED: Image (handled by separate media endpoints)
	// ImageURL *string `json:"image_url,omitempty"`
}

// ============================================================
// NESTED REQUEST TYPES
// ============================================================

// ScheduleRequest represents a schedule for multi-day events
type ScheduleRequest struct {
	ID            string  `json:"id,omitempty"`
	StartDate     string  `json:"start_date" binding:"required"`
	EndDate       string  `json:"end_date,omitempty"`
	StartTime     string  `json:"start_time" binding:"required"`
	EndTime       string  `json:"end_time" binding:"required"`
	Timezone      string  `json:"timezone,omitempty"`
	SessionName   string  `json:"session_name,omitempty"`
	SessionNumber int     `json:"session_number,omitempty"`
	Location      string  `json:"location,omitempty"`
	IsVirtual     bool    `json:"is_virtual"`
	ZoomLink      string  `json:"zoom_link,omitempty"`
	MeetLink      string  `json:"meet_link,omitempty"`
	MaxAttendees  *int    `json:"max_attendees,omitempty"`
}

// RecurrenceRequest represents recurrence configuration
type RecurrenceRequest struct {
	Pattern     string   `json:"pattern" binding:"required,oneof=daily weekly monthly custom"`
	Interval    int      `json:"interval,omitempty"`
	DaysOfWeek  []string `json:"days_of_week,omitempty"`
	DayOfMonth  *int     `json:"day_of_month,omitempty"`
	WeekOfMonth string   `json:"week_of_month,omitempty"`
	EndsOn      string   `json:"ends_on,omitempty"`
	Occurrences *int     `json:"occurrences,omitempty"`
}

// TicketRequest represents a ticket type
type TicketRequest struct {
	ID                 string   `json:"id,omitempty"`
	TicketTypeID       string   `json:"ticket_type_id" binding:"required"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	Price              float64  `json:"price" binding:"min=0"`
	Quantity           int      `json:"quantity" binding:"min=1"`
	MaxPerPerson       *int     `json:"max_per_person,omitempty"`
	EarlyBirdDeadline  string   `json:"early_bird_deadline,omitempty"`
	GroupMinAttendees  *int     `json:"group_min_attendees,omitempty"`
	GroupDiscount      *float64 `json:"group_discount,omitempty"`
}

// SpeakerRequest represents a speaker
type SpeakerRequest struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name" binding:"required"`
	Title       string            `json:"title,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	PhotoURL    string            `json:"photo_url,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	IsKeynote   bool              `json:"is_keynote,omitempty"`
	SortOrder   int               `json:"sort_order,omitempty"`
}

// MaterialRequest represents event materials
type MaterialRequest struct {
	ID             string `json:"id,omitempty"`
	Title          string `json:"title" binding:"required"`
	MaterialTypeID string `json:"material_type_id" binding:"required"`
	URL            string `json:"url" binding:"required"`
	Description    string `json:"description,omitempty"`
	IsPreEvent     bool   `json:"is_pre_event,omitempty"`
	SortOrder      int    `json:"sort_order,omitempty"`
}

// SEORequest represents SEO metadata
type SEORequest struct {
	MetaTitle       string   `json:"meta_title,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	MetaKeywords    []string `json:"meta_keywords,omitempty"`
	CanonicalURL    string   `json:"canonical_url,omitempty"`
	Robots          string   `json:"robots,omitempty"`
	NoIndex         bool     `json:"noindex,omitempty"`
	OGTitle         string   `json:"og_title,omitempty"`
	OGDescription   string   `json:"og_description,omitempty"`
	OGImageURL      string   `json:"og_image_url,omitempty"`
	OGType          string   `json:"og_type,omitempty"`
	TwitterCard     string   `json:"twitter_card,omitempty"`
	TwitterTitle    string   `json:"twitter_title,omitempty"`
	TwitterDescription string `json:"twitter_description,omitempty"`
	TwitterImageURL string   `json:"twitter_image_url,omitempty"`
}

// ============================================================
// BULK REQUEST TYPES
// ============================================================

// BulkIDsRequest - For bulk operations
type BulkIDsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// DuplicateEventRequest - For duplicating a single event
type DuplicateEventRequest struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	IsDraft bool   `json:"is_draft"`
}

// BulkDuplicateRequest - For duplicating multiple events
type BulkDuplicateRequest struct {
	IDs            []string `json:"ids" binding:"required,min=1"`
	NamePrefix     string   `json:"name_prefix"`
	DateOffsetDays int      `json:"date_offset_days"`
	IsDraft        bool     `json:"is_draft"`
}

// ============================================================
// RESPONSE DTOS
// ============================================================

// CreatorDTO - Creator information in API responses
type CreatorDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// EventTypeDTO - Event type information
type EventTypeDTO struct {
	ID                  string `json:"id"`
	Slug                string `json:"slug"`
	Name                string `json:"name"`
	DisplayName         string `json:"display_name"`
	Description         string `json:"description"`
	Icon                string `json:"icon"`
	Color               string `json:"color"`
	SupportsCertificate bool   `json:"supports_certificate"`
	MinDuration         int    `json:"min_duration"`
	MaxDuration         int    `json:"max_duration"`
}

// EventStatusDTO - Event status information
type EventStatusDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	IsFinal     bool   `json:"is_final"`
}

// CategoryDTO - Category information
type CategoryDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// EventFormatDTO - Event format information
type EventFormatDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// TicketTypeDTO - Ticket type information
type TicketTypeDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// CertificateTemplateDTO - Certificate template information
type CertificateTemplateDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
}

// ScheduleDTO - Schedule response
type ScheduleDTO struct {
	ID            string `json:"id"`
	SessionName   string `json:"session_name,omitempty"`
	SessionNumber int    `json:"session_number,omitempty"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date,omitempty"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Timezone      string `json:"timezone"`
	Location      string `json:"location,omitempty"`
	IsVirtual     bool   `json:"is_virtual"`
	ZoomLink      string `json:"zoom_link,omitempty"`
	MeetLink      string `json:"meet_link,omitempty"`
	MaxAttendees  int    `json:"max_attendees,omitempty"`
}

// TicketDTO - Ticket response
type TicketDTO struct {
	ID                 string          `json:"id"`
	TicketType         *TicketTypeDTO  `json:"ticket_type,omitempty"`
	Name               string          `json:"name"`
	Description        string          `json:"description,omitempty"`
	Price              float64         `json:"price"`
	Quantity           int             `json:"quantity"`
	MaxPerPerson       int             `json:"max_per_person,omitempty"`
	EarlyBirdDeadline  string          `json:"early_bird_deadline,omitempty"`
	GroupMinAttendees  int             `json:"group_min_attendees,omitempty"`
	GroupDiscount      float64         `json:"group_discount,omitempty"`
	SortOrder          int             `json:"sort_order"`
	IsActive           bool            `json:"is_active"`
}

// SpeakerDTO - Speaker response
type SpeakerDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Title       string            `json:"title,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	PhotoURL    string            `json:"photo_url,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	IsKeynote   bool              `json:"is_keynote"`
	SortOrder   int               `json:"sort_order"`
}

// MaterialDTO - Material response
type MaterialDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	IsPreEvent  bool   `json:"is_pre_event"`
	SortOrder   int    `json:"sort_order"`
}

// SEODTO - SEO response
type SEODTO struct {
	MetaTitle       string   `json:"meta_title,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	MetaKeywords    []string `json:"meta_keywords,omitempty"`
	CanonicalURL    string   `json:"canonical_url,omitempty"`
	Robots          string   `json:"robots,omitempty"`
	NoIndex         bool     `json:"noindex"`
	OGTitle         string   `json:"og_title,omitempty"`
	OGDescription   string   `json:"og_description,omitempty"`
	OGImageURL      string   `json:"og_image_url,omitempty"`
	OGType          string   `json:"og_type,omitempty"`
	TwitterCard     string   `json:"twitter_card,omitempty"`
	TwitterTitle    string   `json:"twitter_title,omitempty"`
	TwitterDescription string `json:"twitter_description,omitempty"`
	TwitterImageURL string   `json:"twitter_image_url,omitempty"`
}

// VenueDTO - Venue response
type VenueDTO struct {
	Name        string             `json:"name,omitempty"`
	Address     string             `json:"address,omitempty"`
	City        string             `json:"city,omitempty"`
	Country     string             `json:"country,omitempty"`
	Coordinates map[string]float64 `json:"coordinates,omitempty"`
}

// RecurrenceDTO - Recurrence response
type RecurrenceDTO struct {
	Pattern     string   `json:"pattern"`
	Interval    int      `json:"interval"`
	DaysOfWeek  []string `json:"days_of_week,omitempty"`
	DayOfMonth  int      `json:"day_of_month,omitempty"`
	WeekOfMonth string   `json:"week_of_month,omitempty"`
	EndsOn      string   `json:"ends_on,omitempty"`
	Occurrences int      `json:"occurrences,omitempty"`
}

// EventResponse - Complete event response
type EventResponse struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Language         string `json:"language"`

	// Relations
	EventType   *EventTypeDTO        `json:"event_type,omitempty"`
	EventStatus *EventStatusDTO      `json:"event_status,omitempty"`
	Category    *CategoryDTO         `json:"category,omitempty"`
	EventFormat *EventFormatDTO      `json:"event_format,omitempty"`
	CertificateTemplate *CertificateTemplateDTO `json:"certificate_template,omitempty"`

	// Ownership
	OwnerType     string `json:"owner_type"` // "personal" or "institution"
	InstitutionID string `json:"institution_id,omitempty"`
	Creator       *CreatorDTO `json:"creator,omitempty"`

	// Schedule
	StartDate   string         `json:"start_date,omitempty"`
	EndDate     string         `json:"end_date,omitempty"`
	IsMultiDay  bool           `json:"is_multi_day"`
	IsRecurring bool           `json:"is_recurring"`
	Schedules   []ScheduleDTO  `json:"schedules,omitempty"`
	Recurrence  *RecurrenceDTO `json:"recurrence,omitempty"`

	// Venue
	Venue              *VenueDTO `json:"venue,omitempty"`
	IsVirtual          bool      `json:"is_virtual"`
	IsHybrid           bool      `json:"is_hybrid"`
	VirtualPlatform    string    `json:"virtual_platform,omitempty"`
	VirtualPlatformURL string    `json:"virtual_platform_url,omitempty"`
	InPersonLocation   string    `json:"in_person_location,omitempty"`
	ZoomLink           string    `json:"zoom_link,omitempty"`
	MeetLink           string    `json:"meet_link,omitempty"`

	// Tickets
	IsFree             bool         `json:"is_free"`
	Capacity           int          `json:"capacity"`
	CurrentAttendees   int          `json:"current_attendees"`
	WaitlistEnabled    bool         `json:"waitlist_enabled"`
	WaitlistCapacity   int          `json:"waitlist_capacity,omitempty"`
	MinTicketsPerOrder int          `json:"min_tickets_per_order"`
	MaxTicketsPerOrder int          `json:"max_tickets_per_order"`
	Tickets            []TicketDTO  `json:"tickets,omitempty"`

	// Access & Privacy
	Visibility    string   `json:"visibility"`
	IsPrivate     bool     `json:"is_private"`
	InviteOnly    bool     `json:"invite_only"`
	InvitedEmails []string `json:"invited_emails,omitempty"`

	// Monetization
	IsFeatured         bool    `json:"is_featured"`
	CertificateEnabled bool    `json:"certificate_enabled"`
	CertificatePrice   float64 `json:"certificate_price,omitempty"`

	// Speakers & Materials
	Speakers  []SpeakerDTO  `json:"speakers,omitempty"`
	Materials []MaterialDTO `json:"materials,omitempty"`

	// SEO
	SEO       *SEODTO                `json:"seo,omitempty"`
	SchemaOrg map[string]interface{} `json:"schema_org,omitempty"`

	// Media
	ImageURL     string `json:"image_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`

	// Social
	SocialLinks        map[string]string `json:"social_links,omitempty"`
	HasLivestream      bool              `json:"has_livestream"`
	LivestreamURL      string            `json:"livestream_url,omitempty"`
	RecordingAvailable bool              `json:"recording_available"`
	RecordingURL       string            `json:"recording_url,omitempty"`

	// Metadata
	Version            int     `json:"version"`
	PublishedAt        *string `json:"published_at,omitempty"`
	ScheduledPublishAt *string `json:"scheduled_publish_at,omitempty"`
	LastPublishedAt    *string `json:"last_published_at,omitempty"`

	// Audit
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

// ============================================================
// BULK RESPONSE TYPES
// ============================================================

// BulkDeleteResultResponse
type BulkDeleteResultResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// BulkRestoreResultResponse
type BulkRestoreResultResponse struct {
	RestoredCount int      `json:"restored_count"`
	FailedIDs     []string `json:"failed_ids,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// BulkStatusResultResponse
type BulkStatusResultResponse struct {
	ProcessedCount int      `json:"processed_count"`
	FailedIDs      []string `json:"failed_ids,omitempty"`
	Errors         []string `json:"errors,omitempty"`
}

// BulkDuplicateResultResponse
type BulkDuplicateResultResponse struct {
	DuplicatedCount int `json:"duplicated_count"`
	CreatedEvents   []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"created_events"`
	FailedIDs []string `json:"failed_ids,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// MediaInfoResponse - Response for media data
type MediaInfoResponse struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	MediaType  string `json:"media_type"`
	EntityID   string `json:"entity_id"`
	UploadedBy string `json:"uploaded_by"`
	CreatedAt  string `json:"created_at"`
}