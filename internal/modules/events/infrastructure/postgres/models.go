// internal/modules/events/infrastructure/postgres/models.go

package postgres

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// JSONB TYPE FOR GORM
// ============================================================
//
// JSONB stores a raw JSON payload as a byte slice. It handles any
// JSON shape — objects, arrays, primitives, null — because the
// serializer doesn't try to coerce the data into a Go map.
//
// This replaces an earlier `map[string]any` implementation that
// silently dropped arrays and non-object values. json.Unmarshal
// of `["a","b"]` into `map[string]any` fails, so the previous
// Value() returned nil for any list-shaped payload — the DB
// stored NULL for RecurrenceDaysOfWeek, Tags, InvitedEmails, etc.
type JSONB []byte

// Value implements driver.Valuer.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan implements sql.Scanner.
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = v
	case string:
		*j = []byte(v)
	}
	return nil
}

// ============================================================
// EVENT MODEL
// ============================================================

type EventModel struct {
	// ============================================================
	// Core Identity
	// ============================================================
	ID               string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug             string `gorm:"uniqueIndex"`
	Name             string
	DisplayName      string
	Description      string
	ShortDescription string
	Tags             JSONB  `gorm:"type:jsonb;default:'[]'"`
	Language         string `gorm:"default:'en'"`

	// ============================================================
	// Relations (Seeded Lookup Data)
	// ============================================================
	EventTypeID           string  `gorm:"index"`
	EventStatusID         string  `gorm:"index"`
	CategoryID            *string `gorm:"index"`
	EventFormatID         *string `gorm:"index"`
	CertificateTemplateID *string `gorm:"index"`

	// ============================================================
	// Ownership
	// ============================================================
	TeamID    string `gorm:"index;not null"`
	CreatedBy string `gorm:"index"`

	// ============================================================
	// Schedule & Venue
	// ============================================================
	//
	// StartDate, EndDate, Date, Time, Duration, IsMultiDay, IsVirtual,
	// IsHybrid, VirtualPlatform, VirtualPlatformURL, InPersonLocation,
	// ZoomLink, MeetLink, and the Venue* fields are DERIVED from
	// schedules. They remain in the schema as read-optimized caches for
	// the discovery feed, event cards, sorting, and public pages — but
	// the service writes them from the schedules and ignores anything
	// the client sends.
	StartDate          time.Time  `gorm:"index"`
	EndDate            *time.Time `gorm:"index"`
	Date               *time.Time `gorm:"index"`                       // date portion of StartDate
	Time               string     `gorm:"type:varchar(8);default:''"`  // "HH:MM:SS" of first schedule
	Duration           int        `gorm:"default:0"`                   // total minutes across schedules
	IsMultiDay         bool
	IsRecurring        bool `gorm:"index"`

	// Recurrence
	RecurrencePatternID   *string `gorm:"index"`
	RecurrenceInterval    int     `gorm:"default:1"`
	RecurrenceEndsOn      *time.Time
	RecurrenceOccurrences *int
	RecurrenceDaysOfWeek  JSONB `gorm:"type:jsonb;default:'[]'"`
	RecurrenceDayOfMonth  *int
	RecurrenceWeekOfMonth *string

	// Venue (derived)
	VenueName          string
	VenueAddress       string
	VenueCity          string
	VenueCountry       string
	VenueCoordinates   JSONB `gorm:"type:jsonb"`
	IsVirtual          bool  `gorm:"index"`
	IsHybrid           bool  `gorm:"index"`
	VirtualPlatform    string
	VirtualPlatformURL string
	InPersonLocation   string
	ZoomLink           string
	MeetLink           string

	// ============================================================
	// Ticketing & Capacity
	// ============================================================
	IsFree             bool       `gorm:"index"`
	Capacity           *int
	CurrentAttendees   int
	WaitlistEnabled    bool       `gorm:"index"`
	WaitlistCapacity   *int
	TicketSalesStart   *time.Time `gorm:"index"`
	TicketSalesEnd     *time.Time `gorm:"index"`
	MinTicketsPerOrder int        `gorm:"default:1"`
	MaxTicketsPerOrder *int

	// ============================================================
	// Access & Privacy
	// ============================================================
	Visibility          string `gorm:"default:'public';index"`
	Password            *string
	InviteOnly          bool  `gorm:"index"`
	InvitedEmails       JSONB `gorm:"type:jsonb;default:'[]'"`
	RequiresApproval    bool
	ApprovalRequiredFor JSONB `gorm:"type:jsonb;default:'[]'"`

	// ============================================================
	// Monetization & Add-ons
	// ============================================================
	IsFeatured                  bool       `gorm:"default:false;index"`
	FeaturedUntil               *time.Time `gorm:"index"`
	CertificateEnabled          bool       `gorm:"index"`
	CertificatePrice            float64
	EarlyBirdDiscountPercentage *int
	GroupDiscountPercentage     *int
	GroupMinAttendees           *int

	// ============================================================
	// SEO & Marketing
	// ============================================================
	SEOTitle        string
	SEODescription  string
	SEOKeywords     JSONB `gorm:"type:jsonb;default:'[]'"`
	SEOCanonicalURL string
	SEORobots       string
	SEONoIndex      bool `gorm:"index"`

	OGTitle       string
	OGDescription string
	OGImageURL    string
	OGType        string `gorm:"default:'event'"`

	TwitterCard        string `gorm:"default:'summary_large_image'"`
	TwitterTitle       string
	TwitterDescription string
	TwitterImageURL    string

	SchemaOrg JSONB `gorm:"type:jsonb"`

	// ============================================================
	// Media
	// ============================================================
	ImageURL     string
	ThumbnailURL string

	// ============================================================
	// Social & Engagement
	// ============================================================
	SocialLinks        JSONB `gorm:"type:jsonb;default:'{}'"`
	HasLivestream      bool
	LivestreamURL      string
	RecordingAvailable bool
	RecordingURL       string

	// ============================================================
	// Metadata & Versioning
	// ============================================================
	Metadata           JSONB      `gorm:"type:jsonb;default:'{}'"`
	Version            int        `gorm:"default:1;index"`
	PublishedAt        *time.Time `gorm:"index"`
	ScheduledPublishAt *time.Time `gorm:"index"`
	LastPublishedAt    *time.Time

	// ============================================================
	// Audit Fields
	// ============================================================
	IsActive  bool           `gorm:"default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Soft delete tracking
	DeletedBy  *string    `gorm:"index"`
	RestoredAt *time.Time `gorm:"index"`
	RestoredBy *string    `gorm:"index"`

	// ============================================================
	// Creator Info (populated via JOIN for display purposes only)
	// ============================================================
	CreatorName        string `gorm:"-"`
	CreatorDisplayName string `gorm:"-"`
	CreatorEmail       string `gorm:"-"`
	CreatorPhone       string `gorm:"-"`
	CreatorAccountType string `gorm:"-"`
	CreatorAccountName string `gorm:"-"`
}

func (EventModel) TableName() string {
	return "events"
}

// ============================================================
// EVENT SCHEDULE MODEL
// ============================================================

type EventScheduleModel struct {
	ID            string         `gorm:"primaryKey;default:gen_random_uuid()"`
	EventID       string         `gorm:"index;not null"`
	SessionName   string
	SessionNumber int            `gorm:"default:1"`
	StartDate     time.Time      `gorm:"not null"`
	EndDate       *time.Time
	StartTime     string         `gorm:"not null"`
	EndTime       string         `gorm:"not null"`
	Timezone      string         `gorm:"default:'Africa/Nairobi'"`
	Location      string
	IsVirtual     bool           `gorm:"default:false"`
	ZoomLink      string
	MeetLink      string

	// VideoMeetingID references video_meetings.id when the schedule's
	// meeting was created automatically by the video module. Null for
	// manually-pasted links or in-person sessions.
	VideoMeetingID *string `gorm:"index"`

	MaxAttendees *int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (EventScheduleModel) TableName() string {
	return "event_schedules"
}

// ============================================================
// EVENT TICKET MODEL
// ============================================================

type EventTicketModel struct {
	ID                string         `gorm:"primaryKey;default:gen_random_uuid()"`
	EventID           string         `gorm:"index;not null"`
	TicketTypeID      string         `gorm:"index;not null"`
	Name              string         `gorm:"not null"`
	Description       string
	Price             float64        `gorm:"not null;default:0"`
	Quantity          int            `gorm:"not null;default:0"`
	MaxPerPerson      *int
	EarlyBirdDeadline *time.Time
	GroupMinAttendees *int
	GroupDiscount     *float64
	SortOrder         int            `gorm:"default:0"`
	IsActive          bool           `gorm:"default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`

	TicketType *TicketTypeModel `gorm:"foreignKey:TicketTypeID"`
}

func (EventTicketModel) TableName() string {
	return "event_tickets"
}

// ============================================================
// EVENT SPEAKER MODEL
// ============================================================

type EventSpeakerModel struct {
	ID          string         `gorm:"primaryKey;default:gen_random_uuid()"`
	EventID     string         `gorm:"index;not null"`
	Name        string         `gorm:"not null"`
	Title       string
	Bio         string
	PhotoURL    string
	SocialLinks JSONB          `gorm:"type:jsonb;default:'{}'"`
	SortOrder   int            `gorm:"default:0"`
	IsKeynote   bool           `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (EventSpeakerModel) TableName() string {
	return "event_speakers"
}

// ============================================================
// EVENT MATERIAL MODEL
// ============================================================

type EventMaterialModel struct {
	ID             string         `gorm:"primaryKey;default:gen_random_uuid()"`
	EventID        string         `gorm:"index;not null"`
	MaterialTypeID string         `gorm:"index;not null"`
	Title          string         `gorm:"not null"`
	Description    string
	URL            string         `gorm:"not null"`
	IsPreEvent     bool           `gorm:"default:false"`
	SortOrder      int            `gorm:"default:0"`
	FileSize       *int64
	MimeType       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (EventMaterialModel) TableName() string {
	return "event_materials"
}

// ============================================================
// EVENT TYPE MODEL
// ============================================================

type EventTypeModel struct {
	ID                  string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug                string `gorm:"uniqueIndex;not null"`
	Name                string `gorm:"not null"`
	DisplayName         string
	Description         string
	Icon                string
	Color               string
	SortOrder           int  `gorm:"default:0"`
	SupportsCertificate bool `gorm:"default:true"`
	MinDuration         int  `gorm:"default:60"`
	MaxDuration         int  `gorm:"default:480"`
	IsActive            bool `gorm:"default:true"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

func (EventTypeModel) TableName() string {
	return "event_types"
}

// ============================================================
// EVENT STATUS MODEL
// ============================================================

type EventStatusModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int  `gorm:"default:0"`
	IsFinal     bool `gorm:"default:false"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (EventStatusModel) TableName() string {
	return "event_statuses"
}

// ============================================================
// EVENT FORMAT MODEL
// ============================================================

type EventFormatModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (EventFormatModel) TableName() string {
	return "event_formats"
}

// ============================================================
// CATEGORY MODEL
// ============================================================

type CategoryModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int  `gorm:"default:0"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string {
	return "categories"
}

// ============================================================
// RECURRENCE PATTERN MODEL
// ============================================================

type RecurrencePatternModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (RecurrencePatternModel) TableName() string {
	return "recurrence_patterns"
}

// ============================================================
// CERTIFICATE TEMPLATE MODEL
// ============================================================

type CertificateTemplateModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	PreviewURL  string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CertificateTemplateModel) TableName() string {
	return "certificate_templates"
}

// ============================================================
// CERTIFICATE TYPE MODEL
// ============================================================

type CertificateTypeModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CertificateTypeModel) TableName() string {
	return "certificate_types"
}

// ============================================================
// MATERIAL TYPE MODEL
// ============================================================

type MaterialTypeModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	Icon        string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (MaterialTypeModel) TableName() string {
	return "material_types"
}

// ============================================================
// TICKET TYPE MODEL
// ============================================================

type TicketTypeModel struct {
	ID          string `gorm:"primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	SortOrder   int  `gorm:"default:0"`
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (TicketTypeModel) TableName() string {
	return "ticket_types"
}

// ============================================================
// USER MODEL (for creator info)
// ============================================================

type UserModel struct {
	ID            string  `gorm:"primaryKey"`
	Name          string
	DisplayName   string
	Email         string
	Phone         string
	AccountTypeID string  `gorm:"index"`
	AccountID     *string `gorm:"index"`
	AccountName   string  `gorm:"-"`
	IsActive      bool    `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}

// ============================================================
// ACCOUNT MODEL
// ============================================================

type AccountModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	DisplayName string
	Slug        string
	Type        string `gorm:"index"`
	Email       string
	Phone       string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

// ============================================================
// ACCOUNT TYPE MODEL
// ============================================================

type AccountTypeModel struct {
	ID          string `gorm:"primaryKey"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Description string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AccountTypeModel) TableName() string {
	return "account_types"
}

// ============================================================
// TEAM MODEL
// ============================================================

type TeamModel struct {
	ID          string `gorm:"primaryKey"`
	AccountID   string `gorm:"index;not null"`
	Name        string `gorm:"not null"`
	DisplayName string
	Slug        string `gorm:"uniqueIndex;not null"`
	Type        string `gorm:"index"`
	Description string
	LogoURL     string
	IsActive    bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (TeamModel) TableName() string {
	return "teams"
}