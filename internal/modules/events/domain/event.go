// internal/modules/events/domain/event.go

package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Default UUID for drafts when no event type is provided
const DefaultEventTypeID = "00000000-0000-0000-0000-000000000000"

// Event entity
type Event struct {
	// Core fields
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string

	// Relations
	EventTypeID   string
	EventStatusID string
	InstitutionID string // ✅ NEW: Institution this event belongs to
	CreatedBy     string // ✅ User ID who created this event
	Creator       *AccountInfo

	// Media
	ImageURL     string
	ThumbnailURL string

	// Event details
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

	// Audit fields
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time

	// Soft delete fields
	DeletedAt  *time.Time
	DeletedBy  string
	RestoredAt *time.Time
	RestoredBy string

	// Feature flags
	IsFeatured bool
	IsPrivate  bool

	// ✅ DEPRECATED - Use InstitutionID instead
	// AccountID string // Keep for backward compatibility during migration
}

// ============================================================
// FACTORY METHODS
// ============================================================

// NewEvent creates a new event with all required fields
func NewEvent(
	name, displayName, slug, description,
	eventTypeID, institutionID, createdBy string,
) (*Event, error) {
	if name == "" {
		return nil, errors.New("event name is required")
	}
	if displayName == "" {
		return nil, errors.New("event display name is required")
	}
	if slug == "" {
		return nil, errors.New("event slug is required")
	}
	if institutionID == "" {
		return nil, errors.New("institution ID is required")
	}
	if createdBy == "" {
		return nil, errors.New("created by is required")
	}

	now := time.Now()

	return &Event{
		ID:               uuid.New().String(),
		Slug:             slug,
		Name:             name,
		DisplayName:      displayName,
		Description:      description,
		EventTypeID:      eventTypeID,
		EventStatusID:    "",
		InstitutionID:    institutionID,
		CreatedBy:        createdBy,
		ImageURL:         "",
		ThumbnailURL:     "",
		Date:             time.Time{},
		Time:             "",
		Duration:         0,
		Price:            0,
		CertificatePrice: 0,
		Location:         "",
		IsVirtual:        true,
		ZoomLink:         "",
		MeetLink:         "",
		MaxAttendees:     0,
		CurrentAttendees: 0,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
		DeletedAt:        nil,
		DeletedBy:        "",
		RestoredAt:       nil,
		RestoredBy:       "",
		IsFeatured:       false,
		IsPrivate:        false,
	}, nil
}

// ============================================================
// VALIDATION METHODS
// ============================================================

// ValidateForPublish - Validate event before publishing
func (e *Event) ValidateForPublish() error {
	var validationErrors []string

	if e.IsDeleted() {
		validationErrors = append(validationErrors, ErrEventValidationFailed.Error()+": cannot publish a deleted event")
	}
	if e.Name == "" {
		validationErrors = append(validationErrors, ErrEventNameRequired.Error())
	}
	if e.DisplayName == "" {
		validationErrors = append(validationErrors, ErrEventDisplayNameRequired.Error())
	}
	if e.Slug == "" {
		validationErrors = append(validationErrors, ErrEventSlugRequired.Error())
	}
	if e.EventTypeID == "" || e.EventTypeID == DefaultEventTypeID {
		validationErrors = append(validationErrors, ErrEventTypeRequired.Error())
	}
	if e.InstitutionID == "" {
		validationErrors = append(validationErrors, "institution ID is required")
	}
	if e.CreatedBy == "" {
		validationErrors = append(validationErrors, "created by is required")
	}
	if e.Date.IsZero() {
		validationErrors = append(validationErrors, ErrEventDateRequired.Error())
	}
	if !e.Date.IsZero() && e.Date.Before(time.Now()) {
		validationErrors = append(validationErrors, ErrEventPastDate.Error())
	}
	if e.Time == "" {
		validationErrors = append(validationErrors, ErrEventTimeRequired.Error())
	}
	if e.Duration <= 0 {
		validationErrors = append(validationErrors, ErrEventDurationRequired.Error())
	}
	if e.Duration > 0 && e.Duration < 15 {
		validationErrors = append(validationErrors, ErrEventDurationTooShort.Error())
	}
	if e.Duration > 1440 {
		validationErrors = append(validationErrors, ErrEventDurationTooLong.Error())
	}
	if !e.IsVirtual && e.Location == "" {
		validationErrors = append(validationErrors, ErrEventLocationRequired.Error())
	}
	if e.IsVirtual && e.ZoomLink == "" && e.MeetLink == "" {
		validationErrors = append(validationErrors, ErrEventMeetingLinkRequired.Error())
	}
	if e.Price < 0 {
		validationErrors = append(validationErrors, "price cannot be negative")
	}
	if e.CertificatePrice < 0 {
		validationErrors = append(validationErrors, "certificate price cannot be negative")
	}
	if e.MaxAttendees < 0 {
		validationErrors = append(validationErrors, "maximum attendees cannot be negative")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}
	return nil
}

// ValidateForUpdate - Validate event before update
func (e *Event) ValidateForUpdate() error {
	if e.IsDeleted() {
		return errors.New("cannot update a deleted event")
	}
	if e.Name == "" {
		return errors.New("event name is required")
	}
	if e.DisplayName == "" {
		return errors.New("event display name is required")
	}
	if e.Slug == "" {
		return errors.New("event slug is required")
	}
	if e.EventTypeID == "" || e.EventTypeID == DefaultEventTypeID {
		return errors.New("valid event type is required")
	}
	if e.InstitutionID == "" {
		return errors.New("institution ID is required")
	}
	if e.CreatedBy == "" {
		return errors.New("created by is required")
	}
	if e.Date.IsZero() {
		return errors.New("event date is required")
	}
	if e.Time == "" {
		return errors.New("event time is required")
	}
	if e.Duration <= 0 {
		return errors.New("event duration must be greater than 0")
	}
	if e.Duration < 15 {
		return errors.New("duration must be at least 15 minutes")
	}
	if e.Duration > 1440 {
		return errors.New("duration cannot exceed 1440 minutes (24 hours)")
	}
	if !e.IsVirtual && e.Location == "" {
		return errors.New("location is required for in-person events")
	}
	if e.IsVirtual && e.ZoomLink == "" && e.MeetLink == "" {
		return errors.New("at least one meeting link is required for virtual events")
	}
	if e.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if e.CertificatePrice < 0 {
		return errors.New("certificate price cannot be negative")
	}
	if e.MaxAttendees < 0 {
		return errors.New("maximum attendees cannot be negative")
	}
	return nil
}

// ValidateForDraft - Validate draft fields
func (e *Event) ValidateForDraft() error {
	if e.IsDeleted() {
		return errors.New("cannot modify a deleted event")
	}
	if len(e.Name) > 100 {
		return errors.New("event name must be less than 100 characters")
	}
	if len(e.DisplayName) > 200 {
		return errors.New("display name must be less than 200 characters")
	}
	if e.Duration < 0 {
		return errors.New("duration cannot be negative")
	}
	if e.Duration > 1440 {
		return errors.New("duration cannot exceed 1440 minutes (24 hours)")
	}
	if e.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if e.CertificatePrice < 0 {
		return errors.New("certificate price cannot be negative")
	}
	if e.MaxAttendees < 0 {
		return errors.New("maximum attendees cannot be negative")
	}
	return nil
}

// ============================================================
// STATUS METHODS
// ============================================================

// Publish - Validate and publish the event
func (e *Event) Publish() error {
	if err := e.ValidateForPublish(); err != nil {
		return err
	}
	e.UpdatedAt = time.Now()
	return nil
}

// Cancel - Cancel the event
func (e *Event) Cancel() error {
	if e.EventStatusID == "" {
		return errors.New("cannot cancel event without status")
	}
	if e.IsDeleted() {
		return errors.New("cannot cancel a deleted event")
	}
	if e.IsPublished() && e.Date.Before(time.Now()) {
		return errors.New("cannot cancel a past event")
	}
	e.UpdatedAt = time.Now()
	return nil
}

// Complete - Mark the event as completed
func (e *Event) Complete() error {
	if e.EventStatusID == "" {
		return errors.New("cannot complete event without status")
	}
	if e.Date.IsZero() || e.Date.After(time.Now()) {
		return errors.New("cannot complete future event")
	}
	if e.IsDeleted() {
		return errors.New("cannot complete a deleted event")
	}
	if !e.IsPublished() {
		return errors.New("only published events can be completed")
	}
	e.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// SOFT DELETE METHODS
// ============================================================

// SoftDelete - Soft delete the event
func (e *Event) SoftDelete(deletedBy string) error {
	if e.IsDeleted() {
		return errors.New("event is already deleted")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}
	now := time.Now()
	e.DeletedAt = &now
	e.DeletedBy = deletedBy
	e.UpdatedAt = now
	e.IsActive = false
	return nil
}

// Restore - Restore a soft-deleted event
func (e *Event) Restore(restoredBy string) error {
	if !e.IsDeleted() {
		return errors.New("event is not deleted")
	}
	if restoredBy == "" {
		return errors.New("restored by is required")
	}
	now := time.Now()
	e.DeletedAt = nil
	e.DeletedBy = ""
	e.RestoredAt = &now
	e.RestoredBy = restoredBy
	e.UpdatedAt = now
	e.IsActive = true
	return nil
}

// ============================================================
// BOOLEAN CHECKS
// ============================================================

func (e *Event) IsDeleted() bool {
	return e.DeletedAt != nil && !e.DeletedAt.IsZero()
}

func (e *Event) IsDraft() bool {
	return e.EventStatusID == ""
}

func (e *Event) IsPublished() bool {
	// Will be handled by service layer with status lookup
	return false
}

func (e *Event) IsCancelled() bool {
	return false // Will be handled by service
}

func (e *Event) IsCompleted() bool {
	return false // Will be handled by service
}

func (e *Event) IsFull() bool {
	return e.MaxAttendees > 0 && e.CurrentAttendees >= e.MaxAttendees
}

func (e *Event) HasCapacity() bool {
	if e.MaxAttendees == 0 {
		return true
	}
	return e.CurrentAttendees < e.MaxAttendees
}

func (e *Event) AvailableSpots() int {
	if e.MaxAttendees == 0 {
		return -1
	}
	available := e.MaxAttendees - e.CurrentAttendees
	if available < 0 {
		return 0
	}
	return available
}

func (e *Event) IsUpcoming() bool {
	return e.Date.After(time.Now())
}

func (e *Event) IsPast() bool {
	return e.Date.Before(time.Now())
}

func (e *Event) IsPublic() bool {
	return !e.IsPrivate
}

// IsVisibleToUser - Check if event is visible to a specific user
func (e *Event) IsVisibleToUser(userID string) bool {
	if !e.IsPrivate {
		return true
	}
	return e.CreatedBy == userID || e.InstitutionID == userID
}

// ============================================================
// SETTERS
// ============================================================

func (e *Event) SetFeatured(featured bool) {
	e.IsFeatured = featured
	e.UpdatedAt = time.Now()
}

func (e *Event) SetPrivate(private bool) {
	e.IsPrivate = private
	e.UpdatedAt = time.Now()
}

func (e *Event) UpdateTimestamps() {
	e.UpdatedAt = time.Now()
}

func (e *Event) WithCreator(creator *AccountInfo) *Event {
	e.Creator = creator
	return e
}