// internal/modules/events/service/update.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// UPDATE EVENT
// ============================================================

// UpdateEvent updates an existing event.
func (s *eventService) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error) {
	if cmd.ID == "" {
		return nil, errors.New("event ID is required")
	}

	// 1. Get event and check permissions.
	event, err := s.getEventAndCheckUpdatePermission(ctx, cmd.ID, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	// 2. Process name update if needed.
	nameData, err := s.processNameUpdate(ctx, event, cmd)
	if err != nil {
		return nil, err
	}

	// 3. Validate based on status.
	if err := s.validateUpdateFields(ctx, event, cmd, nameData); err != nil {
		return nil, err
	}

	// 4. Apply all updates. Any conversion failure (bad date, bad ticket
	//    shape, bad speaker record) is surfaced here rather than being
	//    silently dropped.
	if err := s.applyAllUpdates(ctx, event, cmd, nameData); err != nil {
		return nil, err
	}

	// 5. Re-derive event-level fields from schedules only when the
	//    caller supplied a schedules array. If cmd.Schedules is nil,
	//    the existing derived values stay.
		if cmd.Schedules != nil {
		deriveEventFromSchedules(event)

		// Create meetings for any newly-added virtual sessions.
		// Existing sessions keep their VideoMeetingID and are skipped
		// by attachVideoMeetings.
		if err := s.attachVideoMeetings(ctx, event, cmd.UpdatedBy); err != nil {
			log.Printf("⚠️ video integration: %v", err)
		}
	}

	// 6. Save to database.
	if err := s.saveUpdatedEvent(ctx, event); err != nil {
		return nil, err
	}

	log.Printf("✅ Event updated: %s by %s", event.ID, cmd.UpdatedBy)
	return event, nil
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// NameUpdateData holds name update information.
type NameUpdateData struct {
	NewName        string
	NewDisplayName string
	NewSlug        string
	NameChanged    bool
}

// processNameUpdate processes name updates and generates new slug if needed.
func (s *eventService) processNameUpdate(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) (*NameUpdateData, error) {
	data := &NameUpdateData{
		NameChanged: false,
	}

	// Handle Name update.
	if cmd.Name != nil {
		rawName := *cmd.Name
		if rawName == "" {
			return nil, errors.New("event name cannot be empty")
		}

		data.NewDisplayName = rawName
		data.NewName = s.validator.Sanitize.Name(rawName)
		if data.NewName == "" {
			return nil, errors.New("event name must contain valid characters after sanitization")
		}

		baseSlug := s.validator.Sanitize.GenerateSlugFromName(data.NewName)
		if baseSlug == "" {
			baseSlug = "untitled"
		}

		data.NewSlug = s.validator.Sanitize.GenerateUniqueSlug(
			baseSlug,
			event.ID,
			func(slug string, excludeID string) bool {
				existing, err := s.repo.GetEventBySlug(ctx, slug)
				if err != nil {
					log.Printf("⚠️ Error checking slug existence: %v", err)
					return false
				}
				return existing != nil && existing.ID != excludeID
			},
		)

		data.NameChanged = true
		log.Printf("📝 Name update: display='%s', internal='%s', slug='%s'",
			data.NewDisplayName, data.NewName, data.NewSlug)
	}

	// Handle DisplayName update.
	if cmd.DisplayName != nil && *cmd.DisplayName != "" {
		data.NewDisplayName = *cmd.DisplayName
		data.NameChanged = true
		log.Printf("📝 Display name explicitly updated: '%s'", data.NewDisplayName)
	}

	return data, nil
}

// validateUpdateFields validates fields based on event status.
func (s *eventService) validateUpdateFields(ctx context.Context, event *domain.Event, cmd UpdateEventCommand, nameData *NameUpdateData) error {
	name, displayName, slug, eventTypeID := s.buildUpdatedFields(event, cmd, nameData)

	isDraft := s.isDraftStatus(ctx, event.EventStatusID)

	// Strict validation for published events.
	if !isDraft {
		if name == "" {
			return errors.New("event name is required")
		}
		if displayName == "" {
			return errors.New("display name is required")
		}
		if slug == "" {
			return errors.New("slug is required")
		}
		if eventTypeID == "" {
			return errors.New("event type is required")
		}
	}

	return nil
}

// buildUpdatedFields builds the updated field values for validation.
func (s *eventService) buildUpdatedFields(event *domain.Event, cmd UpdateEventCommand, nameData *NameUpdateData) (string, string, string, string) {
	var name, displayName, slug, eventTypeID string

	if nameData.NameChanged && nameData.NewName != "" {
		name = nameData.NewName
	} else {
		name = event.Name
	}

	if nameData.NameChanged && nameData.NewDisplayName != "" {
		displayName = nameData.NewDisplayName
	} else {
		displayName = event.DisplayName
	}

	if nameData.NameChanged && nameData.NewSlug != "" {
		slug = nameData.NewSlug
	} else {
		slug = event.Slug
	}

	if cmd.EventTypeID != nil {
		eventTypeID = *cmd.EventTypeID
	}
	if eventTypeID == "" {
		eventTypeID = event.EventTypeID
	}

	return name, displayName, slug, eventTypeID
}

// applyAllUpdates applies all updates to the event.
//
// Any conversion failure is returned so the caller can surface a real
// error instead of silently keeping the old value.
//
// Venue, virtual/hybrid flags, and meeting links are NOT applied from
// the command — they are derived from schedules by the caller. There is
// no applyVenueUpdates step.
func (s *eventService) applyAllUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand, nameData *NameUpdateData) error {
	// Apply name updates.
	if nameData.NameChanged {
		event.Name = nameData.NewName
		event.DisplayName = nameData.NewDisplayName
		event.Slug = nameData.NewSlug
	}
	if cmd.DisplayName != nil && *cmd.DisplayName != "" && !nameData.NameChanged {
		event.DisplayName = *cmd.DisplayName
	}

	// Apply basic field updates.
	s.applyBasicUpdates(event, cmd)

	// Apply schedule updates.
	if err := s.applyScheduleUpdates(ctx, event, cmd); err != nil {
		return fmt.Errorf("schedule update failed: %w", err)
	}

	// Apply ticket updates.
	if err := s.applyTicketUpdates(ctx, event, cmd); err != nil {
		return fmt.Errorf("ticket update failed: %w", err)
	}

	// Apply access & privacy updates.
	s.applyAccessUpdates(event, cmd)

	// Apply monetization updates.
	s.applyMonetizationUpdates(event, cmd)

	// Apply speakers, materials, SEO.
	if err := s.applySpeakersMaterialsSEO(ctx, event, cmd); err != nil {
		return fmt.Errorf("speakers/materials/SEO update failed: %w", err)
	}

	// Update timestamp.
	event.UpdatedAt = time.Now()
	return nil
}

// applyBasicUpdates applies basic field updates.
func (s *eventService) applyBasicUpdates(event *domain.Event, cmd UpdateEventCommand) {
	if cmd.Description != nil {
		event.Description = *cmd.Description
	}
	if cmd.ShortDescription != nil {
		event.ShortDescription = *cmd.ShortDescription
	}
	if cmd.EventTypeID != nil && *cmd.EventTypeID != "" {
		event.EventTypeID = *cmd.EventTypeID
	}
	if cmd.CategoryID != nil {
		event.CategoryID = cmd.CategoryID
	}
	if cmd.Tags != nil {
		event.Tags = cmd.Tags
	}
	if cmd.Language != nil {
		event.Language = *cmd.Language
	}
}

// applyScheduleUpdates applies schedule-related updates.
//
// IsRecurring is treated as the master switch for the recurrence block:
//
//   - nil   → leave the existing recurrence untouched
//   - true  → apply the incoming RecurrenceRequest
//   - false → clear every recurrence column
//
// IsMultiDay, IsVirtual, IsHybrid, VirtualPlatform, VirtualPlatformURL,
// ZoomLink, MeetLink, and venue fields are NOT set here — they are
// derived from schedules by the caller.
func (s *eventService) applyScheduleUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) error {
	if cmd.Schedules != nil {
		schedules, err := s.convertSchedules(cmd.Schedules)
		if err != nil {
			return fmt.Errorf("invalid schedules: %w", err)
		}
		event.Schedules = schedules
	}

	// Recurrence: IsRecurring is the master switch.
	switch {
	case cmd.IsRecurring != nil && !*cmd.IsRecurring:
		// Explicitly off — clear every recurrence field so nothing
		// stale survives into the publish-readiness check.
		event.IsRecurring = false
		event.RecurrencePatternID = nil
		event.RecurrencePatternSlug = ""
		event.RecurrenceInterval = 0
		event.RecurrenceDaysOfWeek = nil
		event.RecurrenceDayOfMonth = nil
		event.RecurrenceWeekOfMonth = nil
		event.RecurrenceEndsOn = nil
		event.RecurrenceOccurrences = nil

	case cmd.IsRecurring != nil && *cmd.IsRecurring:
		// Explicitly on — the recurrence block must be present.
		if cmd.Recurrence == nil {
			return errors.New("is_recurring is true but no recurrence block was provided")
		}
		event.IsRecurring = true
		if err := s.applyRecurrence(ctx, event, cmd.Recurrence); err != nil {
			return fmt.Errorf("invalid recurrence: %w", err)
		}

	case cmd.Recurrence != nil:
		// Recurrence block present but no explicit flag — treat as "on"
		// so older clients that don't send is_recurring still work.
		event.IsRecurring = true
		if err := s.applyRecurrence(ctx, event, cmd.Recurrence); err != nil {
			return fmt.Errorf("invalid recurrence: %w", err)
		}
	}

	return nil
}

// applyTicketUpdates applies ticket-related updates.
func (s *eventService) applyTicketUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) error {
	if cmd.IsFree != nil {
		event.IsFreeEvent = *cmd.IsFree
	}
	if cmd.Capacity != nil {
		event.Capacity = cmd.Capacity
	}
	if cmd.Waitlist != nil {
		event.WaitlistEnabled = *cmd.Waitlist
	}

	if cmd.Tickets != nil {
		tickets, err := s.convertTickets(cmd.Tickets)
		if err != nil {
			return fmt.Errorf("invalid tickets: %w", err)
		}
		event.Tickets = tickets
	}

	return nil
}

// applyAccessUpdates applies access & privacy updates.
func (s *eventService) applyAccessUpdates(event *domain.Event, cmd UpdateEventCommand) {
	if cmd.Visibility != nil {
		event.Visibility = *cmd.Visibility
	}
	if cmd.Password != nil {
		event.Password = cmd.Password
	}
	if cmd.InviteOnly != nil {
		event.InviteOnly = *cmd.InviteOnly
	}
	if cmd.InvitedEmails != nil {
		event.InvitedEmails = cmd.InvitedEmails
	}
}

// applyMonetizationUpdates applies monetization updates.
func (s *eventService) applyMonetizationUpdates(event *domain.Event, cmd UpdateEventCommand) {
	if cmd.IsFeatured != nil {
		event.IsFeatured = *cmd.IsFeatured
	}
	if cmd.CertificateEnabled != nil {
		event.CertificateEnabled = *cmd.CertificateEnabled
	}
	if cmd.CertificatePrice != nil {
		event.CertificatePrice = *cmd.CertificatePrice
	}
	if cmd.CertificateTemplateID != nil {
		event.CertificateTemplateID = cmd.CertificateTemplateID
	}
}

// applySpeakersMaterialsSEO applies speakers, materials, and SEO updates.
func (s *eventService) applySpeakersMaterialsSEO(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) error {
	if cmd.Speakers != nil {
		speakers, err := s.convertSpeakers(cmd.Speakers)
		if err != nil {
			return fmt.Errorf("invalid speakers: %w", err)
		}
		event.Speakers = speakers
	}

	if cmd.Materials != nil {
		materials, err := s.convertMaterials(cmd.Materials)
		if err != nil {
			return fmt.Errorf("invalid materials: %w", err)
		}
		event.Materials = materials
	}

	if cmd.SEO != nil {
		s.applySEO(event, cmd.SEO)
	}

	return nil
}

// saveUpdatedEvent saves the updated event to the database.
//
// After the write, the event is reloaded so DB-assigned schedule IDs
// are populated on the domain struct before we mirror the schedules
// into the attendance module. Without this, a schedule created by
// this update would sync with an empty provider_session_id.
func (s *eventService) saveUpdatedEvent(ctx context.Context, event *domain.Event) error {
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			log.Printf("❌ Duplicate slug detected: %s", event.Slug)
			return fmt.Errorf("an event with the name '%s' already exists. Please use a different name", event.DisplayName)
		}
		return fmt.Errorf("failed to update event: %w", err)
	}

	// Reload so schedule IDs are hydrated before syncing.
	if reloaded, reloadErr := s.repo.GetEventByID(ctx, event.ID); reloadErr == nil && reloaded != nil {
		event.Schedules = reloaded.Schedules
		event.Tickets = reloaded.Tickets  
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event for attendance sync: %v", reloadErr)
	}

	// Mirror schedules into attendance if this is a published event.
	// Drafts don't sync.
	if event.IsPublished() {
		s.syncEventSchedulesToAttendance(ctx, event)
	}

	return nil
}

// isDraftStatus checks if the event status is DRAFT.
func (s *eventService) isDraftStatus(ctx context.Context, statusID string) bool {
	status, err := s.repo.GetEventStatusByID(ctx, statusID)
	if err != nil || status == nil {
		return false
	}
	return status.Slug == domain.EventStatusDraft.GetSlug()
}