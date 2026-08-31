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
// PUBLIC METHOD - Entry Point (30 lines)
// ============================================================

// CreateEvent creates a published event
func (s *eventService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	// 1. Extract and validate context
	institutionID := s.extractInstitutionIDFromEvent(cmd)
	s.logCreateEvent(cmd, institutionID)

	// 2. Check permissions
	if err := s.validateEventPermissions(ctx, cmd, institutionID); err != nil {
		return nil, err
	}

	// 3. Validate required fields for published events
	if err := s.validatePublishedEventFields(cmd); err != nil {
		return nil, err
	}

	// 4. Generate identifiers
	displayName, name, slug := s.generateEventIdentifiers(ctx, cmd.Name)

	// 5. Validate event type
	eventTypeID, err := s.validateEventType(ctx, cmd.EventTypeID)
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
// HELPER FUNCTIONS (< 20 lines each)
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

// validateEventPermissions checks if user has permission to create event
func (s *eventService) validateEventPermissions(ctx context.Context, cmd CreateEventCommand, institutionID string) error {
	if cmd.OwnerType == "personal" {
		return s.validatePersonalEventPermission(ctx, cmd.CreatedBy)
	}
	return s.validateInstitutionEventPermission(ctx, cmd.CreatedBy, institutionID)
}

// validatePersonalEventPermission checks personal event permissions
func (s *eventService) validatePersonalEventPermission(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required for personal events")
	}
	if !s.permChecker.CanManagePersonalEvent(ctx, userID) {
		log.Printf("❌ Permission denied: user %s cannot create personal events", userID)
		return errors.New("insufficient permissions to create personal event")
	}
	return nil
}

// validateInstitutionEventPermission checks institution event permissions
func (s *eventService) validateInstitutionEventPermission(ctx context.Context, userID, institutionID string) error {
	if institutionID == "" {
		return errors.New("institution ID is required for institution events")
	}
	if !s.permChecker.CanManageEvent(ctx, userID, institutionID) {
		log.Printf("❌ Permission denied for user %s on institution %s", userID, institutionID)
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

	// ✅ FIX: Convert and set tickets
	if len(cmd.Tickets) > 0 {
		tickets, err := s.convertTickets(cmd.Tickets)
		if err != nil {
			// Log error but don't fail - validation will catch empty tickets
			log.Printf("⚠️ Failed to convert tickets: %v", err)
		} else {
			event.Tickets = tickets
			log.Printf("🔍 DEBUG: Set %d tickets on event", len(tickets))
		}
	} else {
		log.Printf("🔍 DEBUG: No tickets to set on event")
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

	// Set start date from first schedule
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

	// Validate for publish
	if err := event.ValidateForPublish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return err
	}

	// Publish event
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