// internal/modules/events/service/create_event_draft.go

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
// PUBLIC METHOD - Entry Point (25 lines)
// ============================================================

// CreateDraft creates a draft event
func (s *eventService) CreateDraft(ctx context.Context, cmd CreateDraftCommand) (*domain.Event, error) {
	// 1. Extract and validate context
	institutionID := s.extractInstitutionID(cmd)
	s.logCreateDraft(cmd, institutionID)

	// 2. Check permissions
	if err := s.validateDraftPermissions(ctx, cmd, institutionID); err != nil {
		return nil, err
	}

	// 3. Generate identifiers
	displayName, name, slug := s.generateEventIdentifiers(ctx, cmd.Name)

	// 4. Get or validate event type
	eventTypeID, err := s.resolveEventType(ctx, cmd.EventTypeID)
	if err != nil {
		return nil, err
	}

	// 5. Create and populate domain entity
	event, err := s.buildDraftEvent(ctx, cmd, name, displayName, slug, eventTypeID, institutionID)
	if err != nil {
		return nil, err
	}

	// 6. Set status and save
	if err := s.setDraftStatusAndSave(ctx, event); err != nil {
		return nil, err
	}

	log.Printf("✅ Draft created successfully: %s", event.ID)
	return event, nil
}

// ============================================================
// HELPER FUNCTIONS (< 20 lines each)
// ============================================================

// extractInstitutionID extracts institution ID from command
func (s *eventService) extractInstitutionID(cmd CreateDraftCommand) string {
	if cmd.InstitutionID != nil {
		return *cmd.InstitutionID
	}
	return ""
}

// logCreateDraft logs the draft creation request
func (s *eventService) logCreateDraft(cmd CreateDraftCommand, institutionID string) {
	log.Printf("🔄 CreateDraft called: Name='%s', TypeID='%s', InstitutionID='%s', OwnerType='%s'",
		cmd.Name, cmd.EventTypeID, institutionID, cmd.OwnerType)
}

// validateDraftPermissions checks if user has permission to create draft
func (s *eventService) validateDraftPermissions(ctx context.Context, cmd CreateDraftCommand, institutionID string) error {
	if cmd.OwnerType == "personal" {
		return s.validatePersonalDraftPermission(ctx, cmd.CreatedBy)
	}
	return s.validateInstitutionDraftPermission(ctx, cmd.CreatedBy, institutionID)
}

// validatePersonalDraftPermission checks personal event permissions
func (s *eventService) validatePersonalDraftPermission(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required for personal events")
	}
	if !s.permChecker.CanManagePersonalEvent(ctx, userID) {
		log.Printf("❌ Permission denied: user %s cannot create personal draft events", userID)
		return errors.New("insufficient permissions to create personal draft event")
	}
	return nil
}

// validateInstitutionDraftPermission checks institution event permissions
func (s *eventService) validateInstitutionDraftPermission(ctx context.Context, userID, institutionID string) error {
	if institutionID == "" {
		return errors.New("institution ID is required for institution events")
	}
	if !s.permChecker.CanManageEvent(ctx, userID, institutionID) {
		log.Printf("❌ Permission denied for user %s on institution %s", userID, institutionID)
		return errors.New("insufficient permissions to create events for this institution")
	}
	return nil
}

// generateEventIdentifiers creates name, display name, and slug from user input
func (s *eventService) generateEventIdentifiers(ctx context.Context, rawInput string) (string, string, string) {
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

// resolveEventType gets or validates event type
func (s *eventService) resolveEventType(ctx context.Context, eventTypeID string) (string, error) {
	if eventTypeID == "" {
		return s.getDefaultEventType(ctx)
	}
	return s.validateEventType(ctx, eventTypeID)
}

// getDefaultEventType returns the uncategorized event type ID
func (s *eventService) getDefaultEventType(ctx context.Context) (string, error) {
	log.Printf("ℹ️ No event type provided - using default: %s", domain.EventTypeUncategorized.GetDisplayName())

	uncategorized, err := s.repo.GetEventTypeBySlug(ctx, domain.EventTypeUncategorized.GetSlug())
	if err != nil {
		log.Printf("❌ Failed to get uncategorized event type: %v", err)
		return "", fmt.Errorf("failed to get default event type: %w", err)
	}
	if uncategorized == nil {
		log.Printf("❌ Uncategorized event type not found - please run seeders")
		return "", fmt.Errorf("default event type '%s' not found. Please run seeders", domain.EventTypeUncategorized.GetSlug())
	}
	log.Printf("✅ Using uncategorized event type: %s (%s)", uncategorized.Name, uncategorized.ID)
	return uncategorized.ID, nil
}

// validateEventType validates that the event type exists
func (s *eventService) validateEventType(ctx context.Context, eventTypeID string) (string, error) {
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

// buildDraftEvent creates and populates the domain entity
func (s *eventService) buildDraftEvent(
	ctx context.Context,
	cmd CreateDraftCommand,
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
	s.populateEventFields(event, cmd)

	// Handle schedules separately (needs parsing)
	if err := s.populateSchedules(ctx, event, cmd.Schedules); err != nil {
		return nil, err
	}

	return event, nil
}

// populateEventFields fills all non-schedule fields
func (s *eventService) populateEventFields(event *domain.Event, cmd CreateDraftCommand) {
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

	// ✅ FIX: Convert and set tickets for draft
	if len(cmd.Tickets) > 0 {
		tickets, err := s.convertTickets(cmd.Tickets)
		if err != nil {
			log.Printf("⚠️ Failed to convert tickets for draft: %v", err)
		} else {
			event.Tickets = tickets
			log.Printf("🔍 DEBUG: Set %d tickets on draft event", len(tickets))
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

// populateSchedules converts and sets schedules
func (s *eventService) populateSchedules(ctx context.Context, event *domain.Event, schedules []ScheduleInput) error {
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

// setDraftStatusAndSave sets status to DRAFT, validates, and saves
func (s *eventService) setDraftStatusAndSave(ctx context.Context, event *domain.Event) error {
	log.Printf("🔍 Getting status by slug: %s", domain.EventStatusDraft.GetSlug())
	status, err := s.repo.GetEventStatusBySlug(ctx, domain.EventStatusDraft.GetSlug())
	if err != nil {
		log.Printf("❌ Failed to get event status: %v", err)
		return err
	}
	if status == nil {
		log.Printf("❌ Event status not found: %s", domain.EventStatusDraft.GetSlug())
		return domain.ErrEventStatusNotFound
	}
	event.EventStatusID = status.ID
	log.Printf("✅ Status set: %s (%s)", status.Name, status.ID)

	if err := event.ValidateForDraft(); err != nil {
		log.Printf("❌ Draft validation failed: %v", err)
		return err
	}

	log.Printf("💾 Saving draft to database...")
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to create draft: %v", err)
		return fmt.Errorf("failed to create draft: %w", err)
	}
	return nil
}