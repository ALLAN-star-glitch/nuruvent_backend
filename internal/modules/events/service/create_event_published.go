// internal/modules/events/service/create_event_published.go

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// PUBLIC METHOD - Entry Point
// ============================================================

// CreateEvent creates a published event.
func (s *eventService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	// 1. Validate required fields
	if cmd.TeamID == "" {
		return nil, errors.New("team ID is required")
	}
	if cmd.CreatedBy == "" {
		return nil, errors.New("created by is required")
	}

	s.logCreateEvent(cmd)

	// 2. Resolve the parent account (required for authz and event ownership)
	accountID, err := s.resolveAccountForTeam(ctx, cmd.TeamID, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	cmd.AccountID = accountID

	// 3. Check permissions using the account domain
	if err := s.validatePublishedEventPermissions(ctx, cmd); err != nil {
		return nil, err
	}

	// 4. Validate required fields for published events
	if err := s.validatePublishedEventFields(cmd); err != nil {
		return nil, err
	}

	// 5. Generate identifiers
	displayName, name, slug := s.generateEventIdentifiersForPublished(ctx, cmd.Name)

	// 6. Validate event type
	eventTypeID, err := s.validateEventTypeForPublished(ctx, cmd.EventTypeID)
	if err != nil {
		return nil, err
	}

	// 7. Create and populate domain entity
	event, err := s.buildPublishedEvent(ctx, cmd, name, displayName, slug, eventTypeID)
	if err != nil {
		return nil, err
	}

	// 8. Set status to PUBLISHED, validate, and save
	if err := s.setPublishedStatusAndSave(ctx, event); err != nil {
		return nil, err
	}

	log.Printf("✅ Published event created successfully: %s", event.ID)
	return event, nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func (s *eventService) logCreateEvent(cmd CreateEventCommand) {
	log.Printf("🔄 CreateEvent called: Name='%s', TypeID='%s', TeamID='%s', CreatedBy='%s'",
		cmd.Name, cmd.EventTypeID, cmd.TeamID, cmd.CreatedBy)
}

// validatePublishedEventPermissions checks whether the caller may create
// published events in this account.
func (s *eventService) validatePublishedEventPermissions(ctx context.Context, cmd CreateEventCommand) error {
	accountDomain := domain.AccountDomain(cmd.AccountID)
	if accountDomain == "" {
		return errors.New("cannot resolve account domain: AccountID is empty")
	}

	log.Printf("🔍 AUTHZ: checking event:create for user=%s on domain=%s",
		cmd.CreatedBy, accountDomain)

	allowed, err := s.permChecker.CanCreateEvent(ctx, cmd.CreatedBy, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot create events in account %s",
			cmd.CreatedBy, cmd.AccountID)
		return domain.ErrForbidden
	}

	return nil
}

// validatePublishedEventFields validates required fields for published events.
func (s *eventService) validatePublishedEventFields(cmd CreateEventCommand) error {
	if cmd.Name == "" {
		return domain.ErrInvalidEventName
	}
	if cmd.Description == "" {
		return domain.ErrInvalidEventDescription
	}
	if cmd.EventTypeID == "" {
		return domain.ErrInvalidEventType
	}
	
	// Recurrence may replace schedules
	if len(cmd.Schedules) == 0 && cmd.Recurrence == nil {
		return domain.ErrEventScheduleRequired
	}
	
	if len(cmd.Tickets) == 0 {
		return domain.ErrEventTicketRequired
	}
	if cmd.Visibility == "" {
		return errors.New("visibility is required (public, private, unlisted)")
	}
	return nil
}

// generateEventIdentifiersForPublished creates name, display name, and slug from user input.
func (s *eventService) generateEventIdentifiersForPublished(ctx context.Context, rawInput string) (string, string, string) {
	if rawInput == "" {
		rawInput = "Untitled Event"
	}
	log.Printf("📝 User input: '%s'", rawInput)

	displayName := s.validator.Sanitize.DisplayName(rawInput)
	if displayName == "" {
		displayName = "Untitled Event"
	}
	log.Printf("📝 Display name (preserved): '%s'", displayName)

	name := s.validator.Sanitize.Name(rawInput)
	if name == "" {
		name = "untitled_event"
	}
	log.Printf("📝 Internal name (sanitized): '%s'", name)

	baseSlug := s.validator.Sanitize.GenerateSlugFromName(rawInput)
	if baseSlug == "" {
		baseSlug = "untitled"
	}
	log.Printf("📝 Base slug: '%s'", baseSlug)

	uniqueSlug := s.validator.Sanitize.GenerateUniqueSlug(
		baseSlug,
		"",
		func(slug string, excludeID string) bool {
			existing, err := s.repo.GetEventBySlug(ctx, slug)
			if err != nil {
				log.Printf("⚠️ Error checking slug existence: %v", err)
				return false
			}
			return existing != nil
		},
	)
	log.Printf("📝 Unique slug: '%s'", uniqueSlug)

	return displayName, name, uniqueSlug
}

// validateEventTypeForPublished validates that the event type exists.
func (s *eventService) validateEventTypeForPublished(ctx context.Context, eventTypeID string) (string, error) {
	log.Printf("🔍 Validating event type: %s", eventTypeID)
	eventType, err := s.repo.GetEventTypeByID(ctx, eventTypeID)
	if err != nil {
		log.Printf("❌ Failed to get event type: %v", err)
		return "", fmt.Errorf("failed to get event type: %w", err)
	}
	if eventType == nil {
		log.Printf("❌ Event type not found: %s", eventTypeID)
		return "", domain.ErrEventTypeNotFound
	}
	log.Printf("✅ Event type validated: %s", eventType.Name)
	return eventType.ID, nil
}

// buildPublishedEvent creates and populates the domain entity.
func (s *eventService) buildPublishedEvent(
	ctx context.Context,
	cmd CreateEventCommand,
	name, displayName, slug, eventTypeID string,
) (*domain.Event, error) {
	log.Printf("🏗️ Creating domain entity...")

	if s.repo == nil {
		log.Printf("❌ Repository is nil!")
		return nil, errors.New("repository is not initialized")
	}

	event, err := domain.NewEvent(
		name,
		displayName,
		cmd.Description,
		eventTypeID,
		cmd.TeamID,
		cmd.CreatedBy,
	)
	if err != nil {
		log.Printf("❌ Failed to create domain entity: %v", err)
		return nil, err
	}

	event.Slug = slug
	event.AccountID = cmd.AccountID // set explicitly so ResolveAccountDomain works downstream

	log.Printf("✅ Domain entity created: ID=%s, Name=%s, Slug=%s, DisplayName=%s, TeamID=%s, AccountID=%s",
		event.ID, event.Name, event.Slug, event.DisplayName, event.TeamID, event.AccountID)

	s.populatePublishedEventFields(ctx, event, cmd)

	if err := s.populateEventSchedules(ctx, event, cmd.Schedules); err != nil {
		return nil, err
	}

	// Create meetings on the host's video platform for every virtual
	// session that doesn't already have a link.
	if err := s.attachVideoMeetings(ctx, event, cmd.CreatedBy, ""); err != nil {
		log.Printf("⚠️ video integration: %v", err)
	}

	normalizeSchedulePlatforms(event.Schedules)

	return event, nil
}

// populatePublishedEventFields fills all non-schedule, non-derived fields
// for published events.
func (s *eventService) populatePublishedEventFields(ctx context.Context, event *domain.Event, cmd CreateEventCommand) {
	event.ShortDescription = cmd.ShortDescription
	event.Tags = cmd.Tags
	event.Language = cmd.Language
	event.CategoryID = cmd.CategoryID

	// Recurrence logic
	if cmd.Recurrence != nil {
		event.RecurrenceRule = buildRecurrenceRule(cmd.Recurrence)
	}
	event.IsRecurring = (event.RecurrenceRule != nil)

	// Ticket fields
	event.IsFreeEvent = cmd.IsFree
	event.Capacity = cmd.Capacity
	event.WaitlistEnabled = cmd.Waitlist

	if len(cmd.Tickets) > 0 {
		tickets, err := s.convertTickets(cmd.Tickets)
		if err != nil {
			log.Printf("⚠️ Failed to convert tickets: %v", err)
		} else {
			event.Tickets = tickets
			log.Printf("🔍 Set %d tickets on event", len(tickets))
		}
	}

	// Access & Privacy
	if cmd.Visibility != "" {
		event.Visibility = cmd.Visibility
	}
	event.Password = cmd.Password
	event.InviteOnly = cmd.InviteOnly
	event.InvitedEmails = cmd.InvitedEmails

	// Monetization
	event.IsFeatured = cmd.IsFeatured
	event.CertificateEnabled = cmd.CertificateEnabled
	event.CertificatePrice = cmd.CertificatePrice
	event.CertificateTemplateID = cmd.CertificateTemplateID

	// SEO
	if cmd.SEO != nil {
		s.applySEO(event, cmd.SEO)
	}
}

// populateEventSchedules converts and sets schedules.
func (s *eventService) populateEventSchedules(ctx context.Context, event *domain.Event, schedules []ScheduleInput) error {
	if len(schedules) == 0 && event.RecurrenceRule == nil {
		return nil
	}

	// Handle recurrence expansion
	if event.RecurrenceRule != nil && len(schedules) > 0 {
		// Use the first schedule as the template
		template, err := s.convertSchedules(schedules[:1])
		if err != nil {
			return err
		}

		baseSchedule := template[0]
		dates, err := event.ExpandRecurrence(baseSchedule.StartDate)
		if err != nil {
			return fmt.Errorf("failed to expand recurrence: %w", err)
		}

		var generatedSchedules []domain.EventSchedule
		for i, date := range dates {
			clone := baseSchedule
			clone.ID = "" // let DB assign
			clone.StartDate = date
			// We can't auto-calculate EndDate reliably without knowing duration,
			// but if the template had an EndDate, we can offset it.
			if baseSchedule.EndDate != nil {
				duration := baseSchedule.EndDate.Sub(baseSchedule.StartDate)
				newEnd := date.Add(duration)
				clone.EndDate = &newEnd
			}
			clone.SessionNumber = i + 1 // auto-increment session number
			generatedSchedules = append(generatedSchedules, clone)
		}
		event.Schedules = generatedSchedules
	} else if len(schedules) > 0 {
		// Manual schedules path
		converted, err := s.convertSchedules(schedules)
		if err != nil {
			return err
		}
		event.Schedules = converted
	}

	deriveEventFromSchedules(event)
	return nil
}

// setPublishedStatusAndSave sets status to PUBLISHED, validates, and saves.
func (s *eventService) setPublishedStatusAndSave(ctx context.Context, event *domain.Event) error {
	log.Printf("🔍 Getting status by slug: %s", domain.EventStatusPublished.GetSlug())
	status, err := s.repo.GetEventStatusBySlug(ctx, domain.EventStatusPublished.GetSlug())
	if err != nil {
		log.Printf("❌ Failed to get event status: %v", err)
		return err
	}
	if status == nil {
		log.Printf("❌ Event status not found: %s", domain.EventStatusPublished.GetSlug())
		return domain.ErrEventStatusNotFound
	}
	event.EventStatusID = status.ID
	log.Printf("✅ Status set: %s (%s)", status.Name, status.ID)

	if err := event.ValidateForPublish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return err
	}

	if err := event.Publish(); err != nil {
		log.Printf("❌ Failed to publish event: %v", err)
		return err
	}

	log.Printf("💾 Saving published event to database...")
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to create event: %v", err)
		return fmt.Errorf("failed to create event: %w", err)
	}

	// Reload so DB-assigned schedule IDs land on event.Schedules
	if reloaded, reloadErr := s.repo.GetEventByID(ctx, event.ID); reloadErr == nil && reloaded != nil {
		event.Schedules = reloaded.Schedules
		event.Tickets = reloaded.Tickets
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event for attendance sync: %v", reloadErr)
	}

	// Mirror schedules into the attendance module.
	s.syncEventSchedulesToAttendance(ctx, event)

	// Fire-and-forget: provision video meetings for schedules within
	// the 30-day window. Failures are logged and swallowed — the
	// daily sweep will catch any missed schedules.
	s.enqueueVideoProvisioning(ctx, event.ID)

	return nil
}


// enqueueVideoProvisioning enqueues a fire-and-forget task that
// provisions video meetings for the event's schedules within the
// rolling 30-day window.
//
// This does NOT block the HTTP response. If the queue is unreachable
// or the enqueue fails, we log and move on. The daily sweep will
// catch any schedule that the worker missed.
func (s *eventService) enqueueVideoProvisioning(ctx context.Context, eventID string) {
	if s.queue == nil {
		log.Printf("⚠️ queue not configured; skipping video provisioning for event %s", eventID)
		return
	}

	payload, err := json.Marshal(map[string]string{"event_id": eventID})
	if err != nil {
		log.Printf("⚠️ failed to marshal video provisioning payload for event %s: %v", eventID, err)
		return
	}

	if err := s.queue.Enqueue(ctx, domain.TaskProvisionVideoMeetings, payload); err != nil {
		log.Printf("⚠️ failed to enqueue video provisioning for event %s: %v", eventID, err)
		return
	}

	log.Printf("[events] enqueued video provisioning event=%s", eventID)
}