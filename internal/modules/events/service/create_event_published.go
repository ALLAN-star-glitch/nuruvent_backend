// internal/modules/events/service/create_event_published.go

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

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
//
// NOTE: This function does NOT provision video meetings. Provisioning
// happens asynchronously via the worker after the event is saved.
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
	event.AccountID = cmd.AccountID

	log.Printf("✅ Domain entity created: ID=%s, Name=%s, Slug=%s, DisplayName=%s, TeamID=%s, AccountID=%s",
		event.ID, event.Name, event.Slug, event.DisplayName, event.TeamID, event.AccountID)

	s.populatePublishedEventFields(ctx, event, cmd)

	if err := s.populateEventSchedules(ctx, event, cmd.Schedules); err != nil {
		return nil, err
	}

	if err := s.expandRecurrenceIfNeeded(event); err != nil {
		log.Printf("⚠️ recurrence repair on publish: %v", err)
	}

	normalizeSchedulePlatforms(event.Schedules)

	return event, nil
}

func (s *eventService) populatePublishedEventFields(ctx context.Context, event *domain.Event, cmd CreateEventCommand) {
	event.ShortDescription = cmd.ShortDescription
	event.Tags = cmd.Tags
	event.Language = cmd.Language
	event.CategoryID = cmd.CategoryID

	if cmd.Recurrence != nil {
		event.RecurrenceRule = buildRecurrenceRule(cmd.Recurrence)
	}
	event.IsRecurring = (event.RecurrenceRule != nil)

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

	if cmd.Visibility != "" {
		event.Visibility = cmd.Visibility
	}
	event.Password = cmd.Password
	event.InviteOnly = cmd.InviteOnly
	event.InvitedEmails = cmd.InvitedEmails

	// Registration is enabled by default. Only override when the
	// caller explicitly provided a value.
	if cmd.RegistrationEnabled != nil {
		event.RegistrationEnabled = *cmd.RegistrationEnabled
	}

	event.IsFeatured = cmd.IsFeatured
	event.CertificateEnabled = cmd.CertificateEnabled
	event.CertificatePrice = cmd.CertificatePrice
	event.CertificateTemplateID = cmd.CertificateTemplateID

	if cmd.SEO != nil {
		s.applySEO(event, cmd.SEO)
	}
}

func (s *eventService) populateEventSchedules(ctx context.Context, event *domain.Event, schedules []ScheduleInput) error {
	if len(schedules) == 0 && event.RecurrenceRule == nil {
		return nil
	}

	if event.RecurrenceRule != nil && len(schedules) > 0 {
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
			clone.ID = ""
			clone.StartDate = date
			if baseSchedule.EndDate != nil {
				duration := baseSchedule.EndDate.Sub(baseSchedule.StartDate)
				newEnd := date.Add(duration)
				clone.EndDate = &newEnd
			}
			clone.SessionNumber = i + 1
			clone.SessionName = expandSessionName(baseSchedule.SessionName, i+1)
			generatedSchedules = append(generatedSchedules, clone)
		}
		event.Schedules = generatedSchedules
	} else if len(schedules) > 0 {
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
	if err := s.createEventWithSlugRetry(ctx, event); err != nil {
		log.Printf("❌ Failed to create event: %v", err)
		return err
	}

	if reloaded, reloadErr := s.repo.GetEventByID(ctx, event.ID); reloadErr == nil && reloaded != nil {
		event.Schedules = reloaded.Schedules
		event.Tickets = reloaded.Tickets
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event for attendance sync: %v", reloadErr)
	}

	s.syncEventSchedulesToAttendance(ctx, event)
	s.enqueueVideoProvisioning(ctx, event.ID)

	return nil
}

// ============================================================
// SLUG COLLISION RETRY
// ============================================================

// createEventWithSlugRetry attempts to persist the event. On unique-slug
// collisions (concurrent create, or leftover from a failed prior attempt),
// it bumps the slug (appends -1, -2, ...) and retries up to 5 times.
//
// This is the backend's safety net for races the frontend can't prevent.
func (s *eventService) createEventWithSlugRetry(ctx context.Context, event *domain.Event) error {
	const maxAttempts = 5
	originalSlug := event.Slug

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := s.repo.CreateEvent(ctx, event)
		if err == nil {
			if attempt > 0 {
				log.Printf(
					"[events] slug collision recovered: original=%s final=%s attempts=%d",
					originalSlug, event.Slug, attempt+1,
				)
			}
			return nil
		}

		if !isSlugUniqueViolation(err) {
			return fmt.Errorf("failed to create event: %w", err)
		}

		log.Printf(
			"[events] slug collision on %s (attempt %d/%d); bumping and retrying",
			event.Slug, attempt+1, maxAttempts,
		)

		bumped, bumpErr := s.nextAvailableSlug(ctx, originalSlug, attempt+1)
		if bumpErr != nil {
			return fmt.Errorf("failed to resolve slug collision: %w", bumpErr)
		}
		event.Slug = bumped
	}

	return fmt.Errorf(
		"failed to create event after %d attempts (slug collision on %s)",
		maxAttempts, originalSlug,
	)
}

// nextAvailableSlug returns "<base>-<suffix>" for the first suffix >= start
// that isn't taken by a live (non-soft-deleted) event.
func (s *eventService) nextAvailableSlug(
	ctx context.Context,
	baseSlug string,
	start int,
) (string, error) {
	clean := stripNumericSuffix(baseSlug)

	for i := start; i < start+100; i++ {
		candidate := fmt.Sprintf("%s-%d", clean, i)
		existing, err := s.repo.GetEventBySlug(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("check slug %s: %w", candidate, err)
		}
		if existing == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no available slug in range for base %s", clean)
}

// isSlugUniqueViolation reports whether err is a Postgres 23505 on the
// events_slug_unique constraint.
func isSlugUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" &&
			(pgErr.ConstraintName == "events_slug_unique" ||
				strings.Contains(pgErr.ConstraintName, "slug"))
	}
	return strings.Contains(err.Error(), "events_slug_unique") ||
		(strings.Contains(err.Error(), "duplicate key") &&
			strings.Contains(err.Error(), "slug"))
}

// stripNumericSuffix removes a trailing "-123" from a slug.
func stripNumericSuffix(slug string) string {
	idx := strings.LastIndex(slug, "-")
	if idx < 0 || idx == len(slug)-1 {
		return slug
	}
	suffix := slug[idx+1:]
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return slug
		}
	}
	return slug[:idx]
}

// ============================================================
// ENQUEUE
// ============================================================

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