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

// UpdateEvent updates an existing event
func (s *eventService) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error) {
	if cmd.ID == "" {
		return nil, errors.New("event ID is required")
	}

	// 1. Get event and check permissions
	event, err := s.getEventAndCheckUpdatePermission(ctx, cmd.ID, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	// 2. Process name update if needed
	nameData, err := s.processNameUpdate(ctx, event, cmd)
	if err != nil {
		return nil, err
	}

	// 3. Validate based on status
	if err := s.validateUpdateFields(ctx, event, cmd, nameData); err != nil {
		return nil, err
	}

	// 4. Apply all updates
	s.applyAllUpdates(ctx, event, cmd, nameData)

	// 5. Save to database
	if err := s.saveUpdatedEvent(ctx, event); err != nil {
		return nil, err
	}

	log.Printf("✅ Event updated: %s by %s", event.ID, cmd.UpdatedBy)
	return event, nil
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// NameUpdateData holds name update information
type NameUpdateData struct {
	NewName        string
	NewDisplayName string
	NewSlug        string
	NameChanged    bool
}

// processNameUpdate processes name updates and generates new slug if needed
func (s *eventService) processNameUpdate(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) (*NameUpdateData, error) {
	data := &NameUpdateData{
		NameChanged: false,
	}

	// Handle Name update
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

	// Handle DisplayName update
	if cmd.DisplayName != nil && *cmd.DisplayName != "" {
		if !data.NameChanged {
			data.NewDisplayName = *cmd.DisplayName
		} else {
			data.NewDisplayName = *cmd.DisplayName
		}
		data.NameChanged = true
		log.Printf("📝 Display name explicitly updated: '%s'", data.NewDisplayName)
	}

	return data, nil
}

// validateUpdateFields validates fields based on event status
func (s *eventService) validateUpdateFields(ctx context.Context, event *domain.Event, cmd UpdateEventCommand, nameData *NameUpdateData) error {
	// Build updated values for validation
	name, displayName, slug, eventTypeID := s.buildUpdatedFields(event, cmd, nameData)

	// Check if event is a draft
	isDraft := s.isDraftStatus(ctx, event.EventStatusID)

	// Strict validation for published events
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

// buildUpdatedFields builds the updated field values for validation
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

// applyAllUpdates applies all updates to the event
func (s *eventService) applyAllUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand, nameData *NameUpdateData) {
	// Apply name updates
	if nameData.NameChanged {
		event.Name = nameData.NewName
		event.DisplayName = nameData.NewDisplayName
		event.Slug = nameData.NewSlug
	}
	if cmd.DisplayName != nil && *cmd.DisplayName != "" && !nameData.NameChanged {
		event.DisplayName = *cmd.DisplayName
	}

	// Apply basic field updates
	s.applyBasicUpdates(event, cmd)

	// Apply schedule updates
	s.applyScheduleUpdates(ctx, event, cmd)

	// Apply venue updates
	s.applyVenueUpdates(event, cmd)

	// Apply ticket updates
	s.applyTicketUpdates(ctx, event, cmd)

	// Apply access & privacy updates
	s.applyAccessUpdates(event, cmd)

	// Apply monetization updates
	s.applyMonetizationUpdates(event, cmd)

	// Apply speakers, materials, SEO
	s.applySpeakersMaterialsSEO(ctx, event, cmd)

	// Update timestamp
	event.UpdatedAt = time.Now()
}

// applyBasicUpdates applies basic field updates
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

// applyScheduleUpdates applies schedule-related updates
func (s *eventService) applyScheduleUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) {
	if cmd.IsMultiDay != nil {
		event.IsMultiDay = *cmd.IsMultiDay
	}
	if cmd.IsRecurring != nil {
		event.IsRecurring = *cmd.IsRecurring
	}
	if cmd.Schedules != nil {
		if schedules, err := s.convertSchedules(cmd.Schedules); err == nil {
			event.Schedules = schedules
		}
	}
	if cmd.Recurrence != nil {
		s.applyRecurrence(event, cmd.Recurrence)
	}
}

// applyVenueUpdates applies venue-related updates
func (s *eventService) applyVenueUpdates(event *domain.Event, cmd UpdateEventCommand) {
	if cmd.IsVirtual != nil {
		event.IsVirtual = *cmd.IsVirtual
	}
	if cmd.IsHybrid != nil {
		event.IsHybrid = *cmd.IsHybrid
	}
	if cmd.InPersonLocation != nil {
		event.InPersonLocation = *cmd.InPersonLocation
	}
	if cmd.VirtualPlatform != nil {
		event.VirtualPlatform = *cmd.VirtualPlatform
	}
	if cmd.VirtualPlatformURL != nil {
		event.VirtualPlatformURL = *cmd.VirtualPlatformURL
	}
	if cmd.ZoomLink != nil {
		event.ZoomLink = *cmd.ZoomLink
	}
	if cmd.MeetLink != nil {
		event.MeetLink = *cmd.MeetLink
	}
	if cmd.VenueName != nil {
		event.VenueName = *cmd.VenueName
	}
	if cmd.VenueAddress != nil {
		event.VenueAddress = *cmd.VenueAddress
	}
	if cmd.VenueCity != nil {
		event.VenueCity = *cmd.VenueCity
	}
	if cmd.VenueCountry != nil {
		event.VenueCountry = *cmd.VenueCountry
	}
}

// applyTicketUpdates applies ticket-related updates
func (s *eventService) applyTicketUpdates(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) {
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
		if tickets, err := s.convertTickets(cmd.Tickets); err == nil {
			event.Tickets = tickets
		}
	}
}

// applyAccessUpdates applies access & privacy updates
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

// applyMonetizationUpdates applies monetization updates
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

// applySpeakersMaterialsSEO applies speakers, materials, and SEO updates
func (s *eventService) applySpeakersMaterialsSEO(ctx context.Context, event *domain.Event, cmd UpdateEventCommand) {
	if cmd.Speakers != nil {
		if speakers, err := s.convertSpeakers(cmd.Speakers); err == nil {
			event.Speakers = speakers
		}
	}
	if cmd.Materials != nil {
		if materials, err := s.convertMaterials(cmd.Materials); err == nil {
			event.Materials = materials
		}
	}
	if cmd.SEO != nil {
		s.applySEO(event, cmd.SEO)
	}
}

// saveUpdatedEvent saves the updated event to the database
func (s *eventService) saveUpdatedEvent(ctx context.Context, event *domain.Event) error {
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			log.Printf("❌ Duplicate slug detected: %s", event.Slug)
			return fmt.Errorf("an event with the name '%s' already exists. Please use a different name", event.DisplayName)
		}
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

// isDraftStatus checks if the event status is DRAFT
func (s *eventService) isDraftStatus(ctx context.Context, statusID string) bool {
	status, err := s.repo.GetEventStatusByID(ctx, statusID)
	if err != nil || status == nil {
		return false
	}
	return status.Slug == domain.EventStatusDraft.GetSlug()
}