// internal/modules/events/domain/event.go

package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ✅ Default UUID for drafts when no event type is provided
const DefaultEventTypeID = "00000000-0000-0000-0000-000000000000"

// ✅ Published Status ID - from your database
const PublishedStatusID = "68717108-031d-4ff5-89de-9475b148c25b"

// ✅ Draft Status ID - from your database
const DraftStatusID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"

// ✅ Cancelled Status ID - from your database
const CancelledStatusID = "0c8ad19d-0473-4dc8-ae7d-f9e33242e8bc"

// ✅ Completed Status ID - from your database
const CompletedStatusID = "791595dd-91ad-40cb-a2c1-08cb06a84eb0"

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
	AccountID     string
	CreatedBy     string

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

	// ✅ NEW: Feature flags
	IsFeatured bool
	IsPrivate  bool
}

// ✅ NewEvent - Creates a new event with default values
func NewEvent(name, description, eventTypeID, accountID, createdBy string) (*Event, error) {
	// Name: Default if empty (for drafts)
	if name == "" {
		name = "Untitled Event"
	}

	// EventTypeID: Use default UUID if empty (for drafts)
	if eventTypeID == "" {
		eventTypeID = DefaultEventTypeID
	}

	// AccountID and CreatedBy are required
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if createdBy == "" {
		return nil, errors.New("created by is required")
	}

	now := time.Now()
	return &Event{
		ID:               uuid.New().String(),
		Slug:             generateSlug(name),
		Name:             name,
		DisplayName:      name,
		Description:      description,
		EventTypeID:      eventTypeID,
		EventStatusID:    "",
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
		AccountID:        accountID,
		CreatedBy:        createdBy,
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

func generateSlug(name string) string {
	if name == "" || name == "Untitled Event" {
		return "event-" + uuid.New().String()[:8]
	}
	return "event-" + uuid.New().String()[:8]
}

// ============================================================
// STATUS METHODS
// ============================================================

// Publish - Validate and publish the event
func (e *Event) Publish() error {
	// Validate required fields before publishing
	if err := e.ValidateForPublish(); err != nil {
		return err
	}

	// Set status to PUBLISHED
	e.EventStatusID = PublishedStatusID
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

	e.EventStatusID = CancelledStatusID
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

	e.EventStatusID = CompletedStatusID
	e.UpdatedAt = time.Now()
	return nil
}

// CanPublish - Check if event can be published
func (e *Event) CanPublish() bool {
	return e.EventTypeID != "" &&
		e.EventTypeID != DefaultEventTypeID &&
		e.EventStatusID != "" &&
		e.Date.After(time.Now()) &&
		!e.IsDeleted()
}

// IsFull - Check if event is at capacity
func (e *Event) IsFull() bool {
	return e.MaxAttendees > 0 && e.CurrentAttendees >= e.MaxAttendees
}

// ============================================================
// SOFT DELETE METHODS
// ============================================================

// SoftDelete - Soft delete the event (sets deleted_at and deleted_by)
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

// IsDeleted - Check if event is soft-deleted
func (e *Event) IsDeleted() bool {
	return e.DeletedAt != nil && !e.DeletedAt.IsZero()
}

// IsDraft - Check if event is a draft
func (e *Event) IsDraft() bool {
	return e.EventStatusID == "" || e.EventStatusID == DraftStatusID
}

// IsPublished - Check if event is published
func (e *Event) IsPublished() bool {
	return e.EventStatusID == PublishedStatusID
}

// IsCancelled - Check if event is cancelled
func (e *Event) IsCancelled() bool {
	return e.EventStatusID == CancelledStatusID
}

// IsCompleted - Check if event is completed
func (e *Event) IsCompleted() bool {
	return e.EventStatusID == CompletedStatusID
}

// ============================================================
// VALIDATION METHODS
// ============================================================

// ValidateForPublish - Validate event before publishing
func (e *Event) ValidateForPublish() error {
	if e.IsDeleted() {
		return errors.New("cannot publish a deleted event")
	}
	if e.Name == "" {
		return errors.New("event name is required")
	}
	if e.EventTypeID == "" || e.EventTypeID == DefaultEventTypeID {
		return errors.New("valid event type is required")
	}
	if e.Date.IsZero() {
		return errors.New("event date is required")
	}
	if e.Date.Before(time.Now()) {
		return errors.New("cannot publish past event")
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

// ValidateForUpdate - Validate event before update
func (e *Event) ValidateForUpdate() error {
	if e.IsDeleted() {
		return errors.New("cannot update a deleted event")
	}
	if e.Name == "" {
		return errors.New("event name is required")
	}
	if e.EventTypeID == "" || e.EventTypeID == DefaultEventTypeID {
		return errors.New("valid event type is required")
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

// UpdateTimestamps - Update the updated_at timestamp
func (e *Event) UpdateTimestamps() {
	e.UpdatedAt = time.Now()
}

// ============================================================
// ADDITIONAL HELPER METHODS
// ============================================================

// HasCapacity - Check if there's capacity for more attendees
func (e *Event) HasCapacity() bool {
	if e.MaxAttendees == 0 {
		return true
	}
	return e.CurrentAttendees < e.MaxAttendees
}

// AvailableSpots - Get number of available spots
func (e *Event) AvailableSpots() int {
	if e.MaxAttendees == 0 {
		return -1 // Unlimited
	}
	available := e.MaxAttendees - e.CurrentAttendees
	if available < 0 {
		return 0
	}
	return available
}

// IsUpcoming - Check if event is upcoming
func (e *Event) IsUpcoming() bool {
	return e.Date.After(time.Now())
}

// IsPast - Check if event is past
func (e *Event) IsPast() bool {
	return e.Date.Before(time.Now())
}

// IsPublic - Check if event is public (not private)
func (e *Event) IsPublic() bool {
	return !e.IsPrivate
}

// IsVisibleToUser - Check if event is visible to a specific user
// For now, private events are only visible to the creator
// This can be extended for team/role-based visibility
func (e *Event) IsVisibleToUser(userID string) bool {
	if !e.IsPrivate {
		return true
	}
	return e.CreatedBy == userID || e.AccountID == userID
}

// SetFeatured - Set the event as featured
func (e *Event) SetFeatured(featured bool) {
	e.IsFeatured = featured
	e.UpdatedAt = time.Now()
}

// SetPrivate - Set the event as private
func (e *Event) SetPrivate(private bool) {
	e.IsPrivate = private
	e.UpdatedAt = time.Now()
}