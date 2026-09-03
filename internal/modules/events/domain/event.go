// internal/modules/events/domain/event.go

package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Visibility string

const (
	// VisibilityPublic - Anyone can view the event
	VisibilityPublic Visibility = "public"
	
	// VisibilityPrivate - Only team members can view the event
	VisibilityPrivate Visibility = "private"
	
	// VisibilityUnlisted - Anyone with the direct link can view
	VisibilityUnlisted Visibility = "unlisted"
)

// ============================================================
// EVENT - Aggregate Root
// ============================================================

// Event is the main aggregate root for the events module
type Event struct {
	// ============================================================
	// Core Identity
	// ============================================================
	ID               string
	Slug             string
	Name             string
	DisplayName      string
	Description      string
	ShortDescription string
	Tags             []string
	Language         string

	// ============================================================
	// Relations (Seeded Lookup Data)
	// ============================================================
	EventTypeID          string
	EventStatusID        string
	CategoryID           *string
	Category             *Category 
	EventType		      *EventType
	EventStatus				*EventStatus
	EventFormatID        *string
	CertificateTemplateID *string

	// ============================================================
	// Ownership
	// ============================================================
	InstitutionID *string // NULL for personal events
	CreatedBy     string  // User ID who created this event
	Creator       *UserInfo
	Organizer     *OrganizerInfo 

	// ============================================================
	// Schedule & Venue
	// ============================================================
	StartDate   time.Time
	EndDate     *time.Time
	IsMultiDay  bool
	IsRecurring bool

	// Recurrence
	RecurrencePatternID   *string
	RecurrenceInterval    int
	RecurrenceEndsOn      *time.Time
	RecurrenceOccurrences *int
	RecurrenceDaysOfWeek  []string
	RecurrenceDayOfMonth  *int
	RecurrenceWeekOfMonth *string

	// Venue
	VenueName          string
	VenueAddress       string
	VenueCity          string
	VenueCountry       string
	VenueCoordinates   map[string]float64 // JSONB
	IsVirtual          bool
	IsHybrid           bool
	VirtualPlatform    string
	VirtualPlatformURL string
	InPersonLocation   string
	ZoomLink           string
	MeetLink           string

	// ============================================================
	// Ticketing & Capacity
	// ============================================================
	IsFreeEvent        bool // Renamed to avoid conflict with IsFree() method
	Capacity           *int
	CurrentAttendees   int
	WaitlistEnabled    bool
	WaitlistCapacity   *int
	TicketSalesStart   *time.Time
	TicketSalesEnd     *time.Time
	MinTicketsPerOrder int
	MaxTicketsPerOrder *int

	// ============================================================
	// Access & Privacy
	// ============================================================
	Visibility          string // "public", "private", "unlisted"
	Password            *string
	InviteOnly          bool
	InvitedEmails       []string
	RequiresApproval    bool
	ApprovalRequiredFor []string // ["everyone", "new_users", "unverified"]

	// ============================================================
	// Monetization & Add-ons
	// ============================================================
	IsFeatured          bool
	FeaturedUntil       *time.Time
	CertificateEnabled  bool
	CertificatePrice    float64
	EarlyBirdDiscountPercentage *int
	GroupDiscountPercentage *int
	GroupMinAttendees   *int

	// ============================================================
	// Child Entities (Value Objects - will be loaded separately)
	// ============================================================
	Schedules   []EventSchedule   // Multi-day schedules
	Tickets     []EventTicket     // Ticket types
	Speakers    []EventSpeaker    // Speakers/presenters
	Materials   []EventMaterial   // Event materials
	Invitations []EventInvitation // Invite-only tracking

	// ============================================================
	// SEO & Marketing
	// ============================================================
	SEO struct {
		Title         string
		Description   string
		Keywords      []string
		CanonicalURL  string
		Robots        string
		NoIndex       bool
	}
	OpenGraph struct {
		Title       string
		Description string
		ImageURL    string
		Type        string // "event"
	}
	Twitter struct {
		Card        string // "summary_large_image"
		Title       string
		Description string
		ImageURL    string
	}
	SchemaOrg map[string]interface{} // JSON-LD

	// ============================================================
	// Media
	// ============================================================
	ImageURL     string
	ThumbnailURL string

	// ============================================================
	// Social & Engagement
	// ============================================================
	SocialLinks        map[string]string // LinkedIn, Twitter, etc.
	HasLivestream      bool
	LivestreamURL      string
	RecordingAvailable bool
	RecordingURL       string

	// ============================================================
	// Metadata & Versioning
	// ============================================================
	Metadata           map[string]any
	Version            int
	PublishedAt        *time.Time
	ScheduledPublishAt *time.Time
	LastPublishedAt    *time.Time

	// ============================================================
	// Audit Fields
	// ============================================================
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time

	// Soft delete
	DeletedAt  *time.Time
	DeletedBy  *string
	RestoredAt *time.Time
	RestoredBy *string
}

// ============================================================
// CHILD ENTITY DEFINITIONS
// ============================================================

// EventSchedule represents a schedule for multi-day events
type EventSchedule struct {
	ID           string
	EventID      string
	SessionName  string
	SessionNumber int
	StartDate    time.Time
	EndDate      *time.Time
	StartTime    string
	EndTime      string
	Timezone     string
	Location     string
	IsVirtual    bool
	ZoomLink     string
	MeetLink     string
	MaxAttendees *int
}

// EventTicket represents a ticket type for an event
type EventTicket struct {
	ID                 string
	EventID            string
	TicketTypeID       string
	Name               string
	Description        string
	Price              float64
	Quantity           int
	MaxPerPerson       *int
	EarlyBirdDeadline  *time.Time
	GroupMinAttendees  *int
	GroupDiscount      *float64
	SortOrder          int
	IsActive           bool
}

// EventSpeaker represents a speaker for an event
type EventSpeaker struct {
	ID          string
	EventID     string
	Name        string
	Title       string
	Bio         string
	PhotoURL    string
	SocialLinks map[string]string
	SortOrder   int
	IsKeynote   bool
}

// EventMaterial represents a material for an event
type EventMaterial struct {
	ID             string
	EventID        string
	MaterialTypeID string
	Title          string
	Description    string
	URL            string
	IsPreEvent     bool
	SortOrder      int
	FileSize       *int64
	MimeType       string
}

// EventInvitation represents an invitation for invite-only events
type EventInvitation struct {
	ID         string
	EventID    string
	Email      string
	Name       string
	InvitedBy  *string
	InvitedAt  time.Time
	AcceptedAt *time.Time
	DeclinedAt *time.Time
	Token      string
	ExpiresAt  *time.Time
}

// ============================================================
// FACTORY METHODS
// ============================================================

// NewEvent creates a new event with required fields
func NewEvent(
	name, displayName, description string,
	eventTypeID, createdBy string,
) (*Event, error) {
	if name == "" {
		return nil, ErrInvalidEventName
	}
	if displayName == "" {
		return nil, errors.New("event display name is required")
	}
	if eventTypeID == "" {
		return nil, ErrInvalidEventType
	}
	if createdBy == "" {
		return nil, errors.New("created by is required")
	}

	now := time.Now()
	slug := generateSlug(displayName)

	return &Event{
		ID:          uuid.New().String(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		EventTypeID: eventTypeID,
		EventStatusID: "", // Default to draft
		CreatedBy:   createdBy,
		IsVirtual:   true,
		Visibility:  "public",
		Language:    "en",
		Version:     1,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        []string{},
		InvitedEmails: []string{},
		SocialLinks: make(map[string]string),
		Metadata:    make(map[string]interface{}),
		SchemaOrg:   make(map[string]interface{}),
		OpenGraph: struct {
			Title       string
			Description string
			ImageURL    string
			Type        string
		}{
			Type: "event",
		},
		Twitter: struct {
			Card        string
			Title       string
			Description string
			ImageURL    string
		}{
			Card: "summary_large_image",
		},
		Schedules:   []EventSchedule{},
		Tickets:     []EventTicket{},
		Speakers:    []EventSpeaker{},
		Materials:   []EventMaterial{},
		Invitations: []EventInvitation{},
	}, nil
}

// ============================================================
// VALIDATION METHODS
// ============================================================

// ValidateForPublish validates event before publishing
func (e *Event) ValidateForPublish() error {
	var validationErrors []string

	// ============================================================
	// 1. BASIC CHECKS
	// ============================================================
	if e.IsDeleted() {
		validationErrors = append(validationErrors, "cannot publish a deleted event")
	}
	if e.Name == "" {
		validationErrors = append(validationErrors, ErrInvalidEventName.Error())
	}
	if e.Description == "" {
		validationErrors = append(validationErrors, "description is required for published events")
	}
	if e.EventTypeID == "" {
		validationErrors = append(validationErrors, "event type is required for published events")
	}

	// ============================================================
	// 2. DATE & TIME CHECKS
	// ============================================================
	if e.StartDate.IsZero() {
		validationErrors = append(validationErrors, "start date is required for published events")
	}
	if !e.StartDate.IsZero() && e.StartDate.Before(time.Now()) {
		validationErrors = append(validationErrors, "start date cannot be in the past")
	}

	// ============================================================
	// 3. SCHEDULE CHECKS
	// ============================================================
	if len(e.Schedules) == 0 {
		validationErrors = append(validationErrors, "at least one schedule is required for published events")
	} else {
		// Validate each schedule
		for i, schedule := range e.Schedules {
			if schedule.StartDate.IsZero() {
				validationErrors = append(validationErrors, fmt.Sprintf("schedule %d: start date is required", i+1))
			}
			if schedule.StartTime == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("schedule %d: start time is required", i+1))
			}
			if schedule.EndTime == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("schedule %d: end time is required", i+1))
			}
		}
	}

	// ============================================================
	// 4. VENUE & MEETING LINK CHECKS (IMPROVED)
	// ============================================================
	
	// Check event-level meeting links
	hasEventLevelLink := e.VirtualPlatformURL != "" || e.ZoomLink != "" || e.MeetLink != ""
	
	// Check schedule-level meeting links
	hasScheduleLevelLink := false
	if len(e.Schedules) > 0 {
		for _, schedule := range e.Schedules {
			if schedule.ZoomLink != "" || schedule.MeetLink != "" {
				hasScheduleLevelLink = true
				break
			}
		}
	}
	
	// Determine if meeting link exists at any level
	hasMeetingLink := hasEventLevelLink || hasScheduleLevelLink

	if e.IsVirtual && !hasMeetingLink {
		validationErrors = append(validationErrors, 
			"meeting link is required for virtual events. Provide zoom_link or meet_link at event level or in at least one schedule")
	}

	// For in-person events (non-virtual, non-hybrid)
	if !e.IsVirtual && !e.IsHybrid && e.InPersonLocation == "" {
		validationErrors = append(validationErrors, "location is required for in-person events")
	}

	// For hybrid events
	if e.IsHybrid {
		// Need both virtual and in-person
		if !hasMeetingLink {
			validationErrors = append(validationErrors, 
				"meeting link is required for hybrid events (virtual component)")
		}
		if e.InPersonLocation == "" {
			validationErrors = append(validationErrors, 
				"in-person location is required for hybrid events")
		}
	}

	// ============================================================
	// 5. TICKET CHECKS
	// ============================================================
	if len(e.Tickets) == 0 {
		validationErrors = append(validationErrors, "at least one ticket is required for published events")
	} else {
		totalQuantity := 0
		for i, ticket := range e.Tickets {
			if ticket.TicketTypeID == "" {
				validationErrors = append(validationErrors, fmt.Sprintf("ticket %d: ticket type is required", i+1))
			}
			if ticket.Quantity <= 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("ticket %d: quantity must be greater than 0", i+1))
			}
			if ticket.Price < 0 {
				validationErrors = append(validationErrors, fmt.Sprintf("ticket %d: price cannot be negative", i+1))
			}
			totalQuantity += ticket.Quantity
		}
		
		// Validate capacity against total tickets
		if e.Capacity != nil && *e.Capacity > 0 && totalQuantity > *e.Capacity {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("total ticket quantity (%d) exceeds capacity (%d)", totalQuantity, *e.Capacity))
		}
	}

	// ============================================================
	// 6. CAPACITY CHECKS
	// ============================================================
	if e.Capacity != nil && *e.Capacity < 0 {
		validationErrors = append(validationErrors, "capacity cannot be negative")
	}

	// ============================================================
	// 7. WAITLIST CHECKS
	// ============================================================
	if e.WaitlistEnabled && (e.Capacity == nil || *e.Capacity == 0) {
		validationErrors = append(validationErrors, "waitlist can only be enabled if capacity is set")
	}
	if e.WaitlistCapacity != nil && e.WaitlistEnabled && *e.WaitlistCapacity < 0 {
		validationErrors = append(validationErrors, "waitlist capacity cannot be negative")
	}

	// ============================================================
	// 8. PRICE CHECKS
	// ============================================================
	if e.CertificatePrice < 0 {
		validationErrors = append(validationErrors, "certificate price cannot be negative")
	}
	if e.CertificateEnabled && e.CertificatePrice < 0 {
		validationErrors = append(validationErrors, "certificate price cannot be negative when certificate is enabled")
	}

	// ============================================================
	// 9. RECURRENCE CHECKS
	// ============================================================
	if e.IsRecurring {
		if e.RecurrencePatternID == nil || *e.RecurrencePatternID == "" {
			validationErrors = append(validationErrors, "recurrence pattern is required for recurring events")
		}
		
		// Validate based on pattern
		if e.RecurrencePatternID != nil {
			switch *e.RecurrencePatternID {
			case "weekly":
				if len(e.RecurrenceDaysOfWeek) == 0 {
					validationErrors = append(validationErrors, "days of week are required for weekly recurrence")
				}
			case "monthly":
				if e.RecurrenceDayOfMonth == nil && e.RecurrenceWeekOfMonth == nil {
					validationErrors = append(validationErrors, "day of month or week of month is required for monthly recurrence")
				}
			}
		}
		
		// Check recurrence end
		if e.RecurrenceEndsOn == nil && e.RecurrenceOccurrences == nil {
			validationErrors = append(validationErrors, "recurrence must have an end date or number of occurrences")
		}
	}

	// ============================================================
	// 10. ACCESS & PRIVACY CHECKS
	// ============================================================
	if e.Visibility == "" {
		validationErrors = append(validationErrors, "visibility is required (public, private, unlisted)")
	}
	if e.InviteOnly && len(e.InvitedEmails) == 0 {
		validationErrors = append(validationErrors, "invited emails are required for invite-only events")
	}

	// ============================================================
	// 11. RETURN ERRORS
	// ============================================================
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}
	return nil
}

// ValidateForDraft validates draft event
func (e *Event) ValidateForDraft() error {
	if e.IsDeleted() {
		return errors.New("cannot modify a deleted event")
	}
	if len(e.Name) > 255 {
		return errors.New("event name must be less than 255 characters")
	}
	if len(e.DisplayName) > 150 {
		return errors.New("display name must be less than 150 characters")
	}
	return nil
}

// ValidateForUpdate validates event before update
func (e *Event) ValidateForUpdate() error {
	if e.IsDeleted() {
		return errors.New("cannot update a deleted event")
	}
	if e.Name == "" {
		return ErrInvalidEventName
	}
	if e.DisplayName == "" {
		return errors.New("display name is required")
	}
	return nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// Publish validates and publishes the event
func (e *Event) Publish() error {
	if err := e.ValidateForPublish(); err != nil {
		return err
	}
	now := time.Now()
	e.PublishedAt = &now
	e.LastPublishedAt = &now
	e.UpdatedAt = now
	return nil
}

// Unpublish unpublishes the event
func (e *Event) Unpublish() error {
	if e.IsDeleted() {
		return errors.New("cannot unpublish a deleted event")
	}
	e.PublishedAt = nil
	e.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the event
func (e *Event) Cancel() error {
	if e.IsDeleted() {
		return errors.New("cannot cancel a deleted event")
	}
	if e.IsPast() {
		return errors.New("cannot cancel a past event")
	}
	e.UpdatedAt = time.Now()
	return nil
}

// Complete marks the event as completed
func (e *Event) Complete() error {
	if e.IsDeleted() {
		return errors.New("cannot complete a deleted event")
	}
	if !e.IsPast() {
		return errors.New("cannot complete a future event")
	}
	e.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// SOFT DELETE METHODS
// ============================================================

func (e *Event) SoftDelete(deletedBy string) error {
	if e.IsDeleted() {
		return errors.New("event is already deleted")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}
	now := time.Now()
	e.DeletedAt = &now
	e.DeletedBy = &deletedBy
	e.UpdatedAt = now
	e.IsActive = false
	return nil
}

func (e *Event) Restore(restoredBy string) error {
	if !e.IsDeleted() {
		return errors.New("event is not deleted")
	}
	if restoredBy == "" {
		return errors.New("restored by is required")
	}
	now := time.Now()
	e.DeletedAt = nil
	e.DeletedBy = nil
	e.RestoredAt = &now
	e.RestoredBy = &restoredBy
	e.UpdatedAt = now
	e.IsActive = true
	return nil
}

// ============================================================
// CAPACITY METHODS
// ============================================================

func (e *Event) IsFull() bool {
	if e.Capacity == nil || *e.Capacity == 0 {
		return false
	}
	return e.CurrentAttendees >= *e.Capacity
}

func (e *Event) HasCapacity() bool {
	if e.Capacity == nil || *e.Capacity == 0 {
		return true
	}
	return e.CurrentAttendees < *e.Capacity
}

func (e *Event) AvailableSpots() int {
	if e.Capacity == nil || *e.Capacity == 0 {
		return -1 // Unlimited
	}
	available := *e.Capacity - e.CurrentAttendees
	if available < 0 {
		return 0
	}
	return available
}

// ============================================================
// TICKET METHODS
// ============================================================

func (e *Event) AddTicket(ticket EventTicket) {
	e.Tickets = append(e.Tickets, ticket)
	e.UpdatedAt = time.Now()
}

func (e *Event) RemoveTicket(ticketID string) {
	for i, t := range e.Tickets {
		if t.ID == ticketID {
			e.Tickets = append(e.Tickets[:i], e.Tickets[i+1:]...)
			e.UpdatedAt = time.Now()
			return
		}
	}
}

// ============================================================
// SCHEDULE METHODS
// ============================================================

func (e *Event) AddSchedule(schedule EventSchedule) {
	e.Schedules = append(e.Schedules, schedule)
	e.UpdatedAt = time.Now()
}

func (e *Event) RemoveSchedule(scheduleID string) {
	for i, s := range e.Schedules {
		if s.ID == scheduleID {
			e.Schedules = append(e.Schedules[:i], e.Schedules[i+1:]...)
			e.UpdatedAt = time.Now()
			return
		}
	}
}

// ============================================================
// SPEAKER METHODS
// ============================================================

func (e *Event) AddSpeaker(speaker EventSpeaker) {
	e.Speakers = append(e.Speakers, speaker)
	e.UpdatedAt = time.Now()
}

func (e *Event) RemoveSpeaker(speakerID string) {
	for i, s := range e.Speakers {
		if s.ID == speakerID {
			e.Speakers = append(e.Speakers[:i], e.Speakers[i+1:]...)
			e.UpdatedAt = time.Now()
			return
		}
	}
}

// ============================================================
// MATERIAL METHODS
// ============================================================

func (e *Event) AddMaterial(material EventMaterial) {
	e.Materials = append(e.Materials, material)
	e.UpdatedAt = time.Now()
}

func (e *Event) RemoveMaterial(materialID string) error {
    // Validate state
    if e.IsDeleted() {
        return errors.New("cannot modify a deleted event")
    }
    
    // Find and remove
    for i, m := range e.Materials {
        if m.ID == materialID {
            e.Materials = append(e.Materials[:i], e.Materials[i+1:]...)
            e.UpdatedAt = time.Now()
            return nil
        }
    }
    
    return fmt.Errorf("material with ID %s not found", materialID)
}

// ============================================================
// BOOLEAN CHECKS
// ============================================================

func (e *Event) IsDeleted() bool {
	return e.DeletedAt != nil && !e.DeletedAt.IsZero()
}

func (e *Event) IsDraft() bool {
	return e.PublishedAt == nil
}

func (e *Event) IsPublished() bool {
	return e.PublishedAt != nil && !e.PublishedAt.IsZero()
}

func (e *Event) IsPast() bool {
	return !e.StartDate.IsZero() && e.StartDate.Before(time.Now())
}

func (e *Event) IsUpcoming() bool {
	return !e.StartDate.IsZero() && e.StartDate.After(time.Now())
}

// ============================================================
// VISIBILITY HELPER METHODS ON EVENT
// ============================================================

// String returns the string representation of Visibility
func (v Visibility) String() string {
    return string(v)
}
// IsPublic checks if the event is public
func (e *Event) IsPublic() bool {
    return e.Visibility == string(VisibilityPublic)
}

// IsPrivate checks if the event is private
func (e *Event) IsPrivate() bool {
    return e.Visibility == string(VisibilityPrivate)
}

// IsUnlisted checks if the event is unlisted
func (e *Event) IsUnlisted() bool {
    return e.Visibility == string(VisibilityUnlisted)
}

// GetVisibility returns the visibility as a Visibility type
func (e *Event) GetVisibility() Visibility {
    return Visibility(e.Visibility)
}


func (e *Event) IsFree() bool {
	return e.IsFreeEvent
}

// ============================================================
// PERMISSION METHODS
// ============================================================

func (e *Event) CanEdit(userID string) bool {
	return e.CreatedBy == userID || (e.InstitutionID != nil && *e.InstitutionID == userID)
}

func (e *Event) CanDelete(userID string) bool {
	return e.CreatedBy == userID
}

// ============================================================
// HELPERS
// ============================================================

func generateSlug(displayName string) string {
	slug := strings.ToLower(displayName)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove special characters
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	// Add timestamp for uniqueness
	if slug == "" {
		slug = "event"
	}
	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}

