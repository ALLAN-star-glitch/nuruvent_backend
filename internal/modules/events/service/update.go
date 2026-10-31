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
	"github.com/jackc/pgx/v5/pgconn"
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
		if err := s.attachVideoMeetings(ctx, event, cmd.UpdatedBy, ""); err != nil {
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
//   - true  → apply the incoming Recurrence
//   - false → clear the recurrence
func (s *eventService) applyScheduleUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) error {
	if cmd.Schedules != nil {
		incoming, err := s.convertSchedules(cmd.Schedules)
		if err != nil {
			return fmt.Errorf("invalid schedules: %w", err)
		}

		// Merge incoming schedules with the existing ones so we don't
		// lose fields the client never sends (platform, links, venue,
		// internal video meeting ids, etc.).
		event.Schedules = mergeSchedules(event.Schedules, incoming)

		// Reconcile platform ↔ link consistency.
		normalizeSchedulePlatforms(event.Schedules)
	}

	// Recurrence: IsRecurring is the master switch.
	switch {
	case cmd.IsRecurring != nil && !*cmd.IsRecurring:
		// Explicitly off — clear the rule entirely.
		event.RecurrenceRule = nil
		event.IsRecurring = false

	case cmd.IsRecurring != nil && *cmd.IsRecurring:
		// Explicitly on — the recurrence block must be present.
		if cmd.Recurrence == nil {
			return errors.New("is_recurring is true but no recurrence block was provided")
		}
		event.RecurrenceRule = buildRecurrenceRule(cmd.Recurrence)
		event.IsRecurring = true

	case cmd.Recurrence != nil:
		// Recurrence block present but no explicit flag — treat as "on"
		// so older clients that don't send is_recurring still work.
		event.RecurrenceRule = buildRecurrenceRule(cmd.Recurrence)
		event.IsRecurring = true
	}

	return nil
}

// mergeSchedules merges incoming schedules onto existing ones.
func mergeSchedules(
	existing []domain.EventSchedule,
	incoming []domain.EventSchedule,
) []domain.EventSchedule {
	log.Printf("🔎 mergeSchedules: existing=%d incoming=%d", len(existing), len(incoming))
	for i, s := range incoming {
		log.Printf("🔎   incoming[%d] id=%q platform=%q video_meeting_id=%v",
			i, s.ID, s.Platform, s.VideoMeetingID)
	}
	for i, s := range existing {
		log.Printf("🔎   existing[%d] id=%q platform=%q video_meeting_id=%v",
			i, s.ID, s.Platform, s.VideoMeetingID)
	}

	if len(existing) == 0 {
		out := make([]domain.EventSchedule, len(incoming))
		copy(out, incoming)
		for i := range out {
			if out[i].ID == "" {
				out[i].ID = ""
			}
		}
		return out
	}

	// Index existing schedules by ID for fast lookup.
	byID := make(map[string]domain.EventSchedule, len(existing))
	for _, s := range existing {
		if s.ID != "" {
			byID[s.ID] = s
		}
	}

	// Track which existing schedules have been claimed, so positional
	// fallback doesn't reuse the same one twice.
	claimed := make(map[string]bool, len(existing))

	merged := make([]domain.EventSchedule, 0, len(incoming))
	for i, inc := range incoming {
		var base domain.EventSchedule
		matched := false

		// 1. Match by ID.
		if inc.ID != "" {
			if prev, ok := byID[inc.ID]; ok {
				base = prev
				claimed[prev.ID] = true
				matched = true
			}
		}

		// 2. Fall back to positional match (client dropped the id).
		if !matched && i < len(existing) {
			candidate := existing[i]
			if candidate.ID != "" && !claimed[candidate.ID] {
				base = candidate
				claimed[candidate.ID] = true
				matched = true
				log.Printf("🔎   incoming[%d] had no matching id; "+
					"falling back to positional match with existing[%d] id=%q",
					i, i, candidate.ID)
			}
		}

		// 3. No match — treat as brand new.
		if !matched {
			base = domain.EventSchedule{}
			log.Printf("🔎   incoming[%d] is a new schedule", i)
		}

		merged = append(merged, mergeScheduleFields(base, inc))
	}

	return merged
}

// mergeScheduleFields overlays the incoming schedule onto the existing
// schedule, only overwriting fields the incoming schedule has a
// non-zero value for.
func mergeScheduleFields(existing, incoming domain.EventSchedule) domain.EventSchedule {
	out := existing

	// Preserve ID when the client dropped it.
	if incoming.ID != "" {
		out.ID = incoming.ID
	}

	// --- client-visible fields ---
	if incoming.SessionName != "" {
		out.SessionName = incoming.SessionName
	}
	if incoming.SessionNumber != 0 {
		out.SessionNumber = incoming.SessionNumber
	}
	if !incoming.StartDate.IsZero() {
		out.StartDate = incoming.StartDate
	}
	if incoming.EndDate != nil {
		out.EndDate = incoming.EndDate
	}
	if incoming.StartTime != "" {
		out.StartTime = incoming.StartTime
	}
	if incoming.EndTime != "" {
		out.EndTime = incoming.EndTime
	}
	if incoming.Timezone != "" {
		out.Timezone = incoming.Timezone
	}
	if incoming.Location != "" {
		out.Location = incoming.Location
	}

	// Booleans: incoming true always wins.
	if incoming.IsVirtual {
		out.IsVirtual = true
	}

	if incoming.Platform != "" {
		out.Platform = incoming.Platform
	}
	if incoming.ZoomLink != "" {
		out.ZoomLink = incoming.ZoomLink
	}
	if incoming.MeetLink != "" {
		out.MeetLink = incoming.MeetLink
	}
	if incoming.MaxAttendees != nil {
		out.MaxAttendees = incoming.MaxAttendees
	}

	// --- internal fields the client never sends ---
	if incoming.VideoMeetingID != nil {
		out.VideoMeetingID = incoming.VideoMeetingID
	}
	if incoming.VideoMeetingExternalID != "" {
		out.VideoMeetingExternalID = incoming.VideoMeetingExternalID
	}

	return out
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
func (s *eventService) saveUpdatedEvent(ctx context.Context, event *domain.Event) error {
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		if msg, ok := classifyUniqueViolation(err, event); ok {
			return errors.New(msg)
		}
		return fmt.Errorf("failed to update event: %w", err)
	}

	// Reload the whole event so the returned struct reflects what's
	// actually in the DB.
	if reloaded, reloadErr := s.repo.GetEventByID(ctx, event.ID); reloadErr == nil && reloaded != nil {
		*event = *reloaded
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event for attendance sync: %v", reloadErr)
	}

	s.syncEventSchedulesToAttendance(ctx, event)

	return nil
}

// classifyUniqueViolation inspects a Postgres 23505 error and returns
// a user-facing message that names the actual constraint that fired.
func classifyUniqueViolation(err error, event *domain.Event) (string, bool) {
	if err == nil {
		return "", false
	}

	errStr := err.Error()

	if !strings.Contains(errStr, "23505") && !strings.Contains(errStr, "duplicate key") {
		return "", false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch {
		case strings.Contains(pgErr.ConstraintName, "event_schedules") &&
			strings.Contains(pgErr.ConstraintName, "session_number"):
			return "sessions must have unique numbers within the event; reload the page and try again", true
		case strings.Contains(pgErr.ConstraintName, "slug"):
			log.Printf("❌ Duplicate slug detected: %s", event.Slug)
			return fmt.Sprintf(
				"an event with the name '%s' already exists. Please use a different name",
				event.DisplayName,
			), true
		default:
			log.Printf("❌ Unhandled unique violation on %s: %s", pgErr.ConstraintName, errStr)
			return fmt.Sprintf("update rejected by constraint %s", pgErr.ConstraintName), true
		}
	}

	switch {
	case strings.Contains(errStr, "uniq_event_schedules_event_session_number"):
		return "sessions must have unique numbers within the event; reload the page and try again", true
	case strings.Contains(errStr, "slug"):
		log.Printf("❌ Duplicate slug detected: %s", event.Slug)
		return fmt.Sprintf(
			"an event with the name '%s' already exists. Please use a different name",
			event.DisplayName,
		), true
	}

	return "", false
}

// isDraftStatus checks if the event status is DRAFT.
func (s *eventService) isDraftStatus(ctx context.Context, statusID string) bool {
	status, err := s.repo.GetEventStatusByID(ctx, statusID)
	if err != nil || status == nil {
		return false
	}
	return status.Slug == domain.EventStatusDraft.GetSlug()
}

// normalizeSchedulePlatforms reconciles the relationship between
// schedule.platform and schedule.{zoom_link,meet_link}.
func normalizeSchedulePlatforms(schedules []domain.EventSchedule) {
	for i := range schedules {
		s := &schedules[i]

		switch s.Platform {
		case "zoom":
			if s.MeetLink != "" {
				log.Printf("🔧 normalize: schedule %s platform=zoom, clearing meet_link", s.ID)
				s.MeetLink = ""
			}
		case "google_meet":
			if s.ZoomLink != "" {
				log.Printf("🔧 normalize: schedule %s platform=google_meet, clearing zoom_link", s.ID)
				s.ZoomLink = ""
			}
		case "":
			switch {
			case s.ZoomLink != "" && s.MeetLink != "":
				log.Printf("🔧 normalize: schedule %s has both links, defaulting to zoom", s.ID)
				s.Platform = "zoom"
				s.MeetLink = ""
			case s.ZoomLink != "":
				s.Platform = "zoom"
			case s.MeetLink != "":
				s.Platform = "google_meet"
			}
		}
	}
}