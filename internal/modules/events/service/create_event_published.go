// internal/modules/events/service/create_event_published.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// PUBLIC METHOD - Entry Point
// ============================================================

// CreateEvent creates a published event
func (s *eventService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	// 1. Extract and validate context
	institutionID := s.extractInstitutionIDFromEvent(cmd)
	s.logCreateEvent(cmd, institutionID)

	// 2. Check permissions using the new Scope-based approach
	if err := s.validateEventPermissions(ctx, cmd, institutionID); err != nil {
		return nil, err
	}

	// 3. Validate required fields for published events
	if err := s.validatePublishedEventFields(cmd); err != nil {
		return nil, err
	}

	// 4. Generate identifiers
	displayName, name, slug := s.generateEventIdentifiersForPublished(ctx, cmd.Name)

	// 5. Validate event type
	eventTypeID, err := s.validateEventTypeForPublished(ctx, cmd.EventTypeID)
	if err != nil {
		return nil, err
	}

	// 6. Create and populate domain entity
	event, err := s.buildPublishedEvent(ctx, cmd, name, displayName, slug, eventTypeID, institutionID)
	if err != nil {
		return nil, err
	}

	// 7. Set status to PUBLISHED, validate, and save
	if err := s.setPublishedStatusAndSave(ctx, event); err != nil {
		return nil, err
	}

	log.Printf("✅ Published event created successfully: %s", event.ID)
	return event, nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// extractInstitutionIDFromEvent extracts institution ID from command
func (s *eventService) extractInstitutionIDFromEvent(cmd CreateEventCommand) string {
	if cmd.InstitutionID != nil {
		return *cmd.InstitutionID
	}
	return ""
}

// logCreateEvent logs the event creation request
func (s *eventService) logCreateEvent(cmd CreateEventCommand, institutionID string) {
	log.Printf("🔄 CreateEvent called: Name='%s', TypeID='%s', InstitutionID='%s', OwnerType='%s'",
		cmd.Name, cmd.EventTypeID, institutionID, cmd.OwnerType)
}

// validateEventPermissions checks if user has permission to create event using Scope
func (s *eventService) validateEventPermissions(ctx context.Context, cmd CreateEventCommand, institutionID string) error {
	var scope domain.Scope

	if cmd.OwnerType == "personal" {
		if cmd.CreatedBy == "" {
			return errors.New("user ID is required for personal events")
		}
		scope = domain.NewPersonalTeamScope(cmd.CreatedBy) // Personal events are scoped to the user

		allowed, err := s.permChecker.CanCreateEvent(ctx, cmd.CreatedBy, scope)
		if err != nil {
			return fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			log.Printf("❌ Permission denied: user %s cannot create personal published events", cmd.CreatedBy)
			return errors.New("insufficient permissions to create personal event")
		}
		return nil
	}

	// Institution event
	if institutionID == "" {
		return errors.New("institution ID is required for institution events")
	}
	if cmd.CreatedBy == "" {
		return errors.New("user ID is required")
	}

	scope = domain.NewInstitutionTeamScope(institutionID)

	allowed, err := s.permChecker.CanCreateEvent(ctx, cmd.CreatedBy, scope)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot create events for institution %s", cmd.CreatedBy, institutionID)
		return errors.New("insufficient permissions to create events for this institution")
	}

	return nil
}

// validatePublishedEventFields validates required fields for published events
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
	if len(cmd.Schedules) == 0 {
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

//  generateEventIdentifiersForPublished creates name, display name, and slug from user input
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

// ✅ validateEventTypeForPublished validates that the event type exists
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

// buildPublishedEvent creates and populates the domain entity
func (s *eventService) buildPublishedEvent(
	ctx context.Context,
	cmd CreateEventCommand,
	name, displayName, slug, eventTypeID, institutionID string,
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
		cmd.CreatedBy,
	)
	if err != nil {
		log.Printf("❌ Failed to create domain entity: %v", err)
		return nil, err
	}

	event.Slug = slug
	if institutionID != "" {
		event.InstitutionID = &institutionID
	}

	log.Printf("✅ Domain entity created: ID=%s, Name=%s, Slug=%s, DisplayName=%s",
		event.ID, event.Name, event.Slug, event.DisplayName)

	// Populate all fields
	s.populatePublishedEventFields(event, cmd)

	// Handle schedules separately (needs parsing)
	if err := s.populateEventSchedules(ctx, event, cmd.Schedules); err != nil {
		return nil, err
	}

	return event, nil
}

// populatePublishedEventFields fills all non-schedule fields for published events
func (s *eventService) populatePublishedEventFields(event *domain.Event, cmd CreateEventCommand) {
	event.ShortDescription = cmd.ShortDescription
	event.Tags = cmd.Tags
	event.Language = cmd.Language
	event.CategoryID = cmd.CategoryID

	event.IsMultiDay = cmd.IsMultiDay
	event.IsRecurring = cmd.IsRecurring

	if cmd.Recurrence != nil {
		s.applyRecurrence(event, cmd.Recurrence)
	}

	// Venue fields
	event.IsVirtual = cmd.IsVirtual
	event.IsHybrid = cmd.IsHybrid
	event.InPersonLocation = cmd.InPersonLocation
	event.VirtualPlatform = cmd.VirtualPlatform
	event.VirtualPlatformURL = cmd.VirtualPlatformURL
	event.ZoomLink = cmd.ZoomLink
	event.MeetLink = cmd.MeetLink
	event.VenueName = cmd.VenueName
	event.VenueAddress = cmd.VenueAddress
	event.VenueCity = cmd.VenueCity
	event.VenueCountry = cmd.VenueCountry

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

// populateEventSchedules converts and sets schedules
func (s *eventService) populateEventSchedules(ctx context.Context, event *domain.Event, schedules []ScheduleInput) error {
	if len(schedules) == 0 {
		return nil
	}

	converted, err := s.convertSchedules(schedules)
	if err != nil {
		return err
	}
	event.Schedules = converted

	startDate, err := time.Parse("2006-01-02", schedules[0].StartDate)
	if err == nil {
		event.StartDate = startDate
	}
	return nil
}

// setPublishedStatusAndSave sets status to PUBLISHED, validates, and saves
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
	return nil
}