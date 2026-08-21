// internal/modules/events/service/service_impl.go

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
// SERVICE IMPLEMENTATION
// ============================================================

type eventService struct {
	repo        domain.Repository
	permChecker domain.PermissionChecker
	mediaSvc    domain.MediaService
}

func NewService(
	repo domain.Repository,
	permChecker domain.PermissionChecker,
	mediaSvc domain.MediaService,
) Service {
	return &eventService{
		repo:        repo,
		permChecker: permChecker,
		mediaSvc:    mediaSvc,
	}
}

// ============================================================
// CREATE DRAFT (with optional image)
// ============================================================

func (s *eventService) CreateDraft(ctx context.Context, cmd CreateDraftCommand) (*domain.Event, error) {
	log.Printf("🔄 CreateDraft called: Name='%s', TypeID='%s', AccountID='%s'",
		cmd.Name, cmd.EventTypeID, cmd.AccountID)

	// Check permission
	if !s.permChecker.CanManageEvent(ctx, cmd.CreatedBy, cmd.AccountID) {
		log.Printf("❌ Permission denied for user %s on account %s", cmd.CreatedBy, cmd.AccountID)
		return nil, errors.New("insufficient permissions to create events for this account")
	}

	// Name: Default if empty
	if cmd.Name == "" {
		cmd.Name = "Untitled Event"
		log.Printf("📝 Generated default name: 'Untitled Event'")
	}

	// Event Type: Optional - only validate if provided
	if cmd.EventTypeID != "" {
		log.Printf("🔍 Validating event type: %s", cmd.EventTypeID)
		eventType, err := s.repo.GetEventTypeByID(ctx, cmd.EventTypeID)
		if err != nil {
			log.Printf("❌ Failed to get event type: %v", err)
			return nil, fmt.Errorf("failed to get event type: %w", err)
		}
		if eventType == nil {
			log.Printf("❌ Event type not found: %s", cmd.EventTypeID)
			return nil, domain.ErrEventTypeNotFound
		}
		log.Printf("✅ Event type validated: %s", eventType.Name)
	} else {
		log.Printf("ℹ️ No event type provided - saving as empty")
	}

	// Create domain entity
	log.Printf("🏗️ Creating domain entity...")

	if s.repo == nil {
		log.Printf("❌ Repository is nil!")
		return nil, errors.New("repository is not initialized")
	}

	event, err := domain.NewEvent(
		cmd.Name,
		cmd.Description,
		cmd.EventTypeID,
		cmd.AccountID,
		cmd.CreatedBy,
	)
	if err != nil {
		log.Printf("❌ Failed to create domain entity: %v", err)
		return nil, err
	}
	log.Printf("✅ Domain entity created with ID: %s", event.ID)

	// Set additional fields (all optional for drafts)
	if cmd.Date != "" {
		event.Date = parseDate(cmd.Date)
		log.Printf("📅 Date set: %s", cmd.Date)
	}
	event.Time = cmd.Time
	event.Duration = cmd.Duration
	event.Price = cmd.Price
	event.CertificatePrice = cmd.CertificatePrice
	event.Location = cmd.Location
	event.IsVirtual = cmd.IsVirtual
	event.ZoomLink = cmd.ZoomLink
	event.MeetLink = cmd.MeetLink
	event.MaxAttendees = cmd.MaxAttendees

	// ✅ Set feature flags
	event.IsFeatured = cmd.IsFeatured
	event.IsPrivate = cmd.IsPrivate

	// Set status to DRAFT
	log.Printf("🔍 Getting status by slug: draft")
	status, err := s.repo.GetEventStatusBySlug(ctx, "draft")
	if err != nil {
		log.Printf("❌ Failed to get event status: %v", err)
		return nil, err
	}
	if status == nil {
		log.Printf("❌ Event status not found: draft")
		return nil, domain.ErrEventStatusNotFound
	}
	event.EventStatusID = status.ID
	log.Printf("✅ Status set: %s (%s)", status.Name, status.ID)

	// Save draft
	log.Printf("💾 Saving draft to database...")
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to create draft: %v", err)
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}
	log.Printf("✅ Draft created successfully: %s", event.ID)

	// ✅ If image is provided, upload it using pure domain types
	if len(cmd.ImageData) > 0 {
		log.Printf("🔄 Uploading image for draft %s", event.ID)

		if s.mediaSvc == nil {
			log.Printf("⚠️ Media service is nil, skipping image upload")
			return event, nil
		}

		media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
			File:          cmd.ImageData,
			FileName:      cmd.ImageName,
			ContentType:   cmd.ContentType,
			MediaTypeName: "Event",
			EntityID:      event.ID,
			UploadedBy:    cmd.CreatedBy,
		})
		if err != nil {
			log.Printf("⚠️ Failed to upload image for draft %s: %v", event.ID, err)
			return event, nil
		}
		log.Printf("✅ Image uploaded: %s", media.URL)

		// Update event with image URL
		event.ImageURL = media.URL
		if err := s.repo.UpdateEvent(ctx, event); err != nil {
			log.Printf("⚠️ Failed to update draft %s with image: %v", event.ID, err)
			return event, nil
		}
		log.Printf("✅ Draft updated with image URL")
	} else {
		log.Printf("ℹ️ No image provided for draft %s", event.ID)
	}

	return event, nil
}

// ============================================================
// CREATE PUBLISHED EVENT (with optional image)
// ============================================================

func (s *eventService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	log.Printf("🔄 CreateEvent (published) called: Name='%s', TypeID='%s', AccountID='%s'",
		cmd.Name, cmd.EventTypeID, cmd.AccountID)

	// Check permission
	if !s.permChecker.CanManageEvent(ctx, cmd.CreatedBy, cmd.AccountID) {
		log.Printf("❌ Permission denied for user %s on account %s", cmd.CreatedBy, cmd.AccountID)
		return nil, errors.New("insufficient permissions to create events for this account")
	}

	// STRICT validation for published events
	if cmd.Name == "" {
		return nil, errors.New("event name is required for published events")
	}
	if cmd.EventTypeID == "" {
		return nil, errors.New("event type is required for published events")
	}
	if cmd.Date == "" {
		return nil, errors.New("event date is required for published events")
	}
	if cmd.Time == "" {
		return nil, errors.New("event time is required for published events")
	}
	if cmd.Duration <= 0 {
		return nil, errors.New("event duration must be greater than 0 for published events")
	}
	if cmd.Duration < 15 {
		return nil, errors.New("duration must be at least 15 minutes for published events")
	}
	if !cmd.IsVirtual && cmd.Location == "" {
		return nil, errors.New("location is required for in-person events")
	}
	if cmd.IsVirtual && cmd.ZoomLink == "" && cmd.MeetLink == "" {
		return nil, errors.New("at least one meeting link is required for virtual events")
	}

	// Get event type
	log.Printf("🔍 Validating event type: %s", cmd.EventTypeID)
	eventType, err := s.repo.GetEventTypeByID(ctx, cmd.EventTypeID)
	if err != nil {
		log.Printf("❌ Failed to get event type: %v", err)
		return nil, fmt.Errorf("failed to get event type: %w", err)
	}
	if eventType == nil {
		log.Printf("❌ Event type not found: %s", cmd.EventTypeID)
		return nil, domain.ErrEventTypeNotFound
	}
	log.Printf("✅ Event type validated: %s", eventType.Name)

	// Create domain entity
	log.Printf("🏗️ Creating domain entity...")

	if s.repo == nil {
		log.Printf("❌ Repository is nil!")
		return nil, errors.New("repository is not initialized")
	}

	event, err := domain.NewEvent(
		cmd.Name,
		cmd.Description,
		cmd.EventTypeID,
		cmd.AccountID,
		cmd.CreatedBy,
	)
	if err != nil {
		log.Printf("❌ Failed to create domain entity: %v", err)
		return nil, err
	}
	log.Printf("✅ Domain entity created with ID: %s", event.ID)

	// Set all fields
	event.Date = parseDate(cmd.Date)
	event.Time = cmd.Time
	event.Duration = cmd.Duration
	event.Price = cmd.Price
	event.CertificatePrice = cmd.CertificatePrice
	event.Location = cmd.Location
	event.IsVirtual = cmd.IsVirtual
	event.ZoomLink = cmd.ZoomLink
	event.MeetLink = cmd.MeetLink
	event.MaxAttendees = cmd.MaxAttendees

	// ✅ Set feature flags
	event.IsFeatured = cmd.IsFeatured
	event.IsPrivate = cmd.IsPrivate

	// Set status to PUBLISHED
	log.Printf("🔍 Getting status by slug: published")
	status, err := s.repo.GetEventStatusBySlug(ctx, "published")
	if err != nil {
		log.Printf("❌ Failed to get event status: %v", err)
		return nil, err
	}
	if status == nil {
		log.Printf("❌ Event status not found: published")
		return nil, domain.ErrEventStatusNotFound
	}
	event.EventStatusID = status.ID
	log.Printf("✅ Status set: %s (%s)", status.Name, status.ID)

	// Save event
	log.Printf("💾 Saving published event to database...")
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to create event: %v", err)
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	log.Printf("✅ Published event created successfully: %s", event.ID)

	// ✅ If image is provided, upload it using pure domain types
	if len(cmd.ImageData) > 0 {
		log.Printf("🔄 Uploading image for event %s", event.ID)

		if s.mediaSvc == nil {
			log.Printf("⚠️ Media service is nil, skipping image upload")
			return event, nil
		}

		media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
			File:          cmd.ImageData,
			FileName:      cmd.ImageName,
			ContentType:   cmd.ContentType,
			MediaTypeName: "Event",
			EntityID:      event.ID,
			UploadedBy:    cmd.CreatedBy,
		})
		if err != nil {
			log.Printf("⚠️ Failed to upload image for event %s: %v", event.ID, err)
			return event, nil
		}
		log.Printf("✅ Image uploaded: %s", media.URL)

		// Update event with image URL
		event.ImageURL = media.URL
		if err := s.repo.UpdateEvent(ctx, event); err != nil {
			log.Printf("⚠️ Failed to update event %s with image URL: %v", event.ID, err)
			return event, nil
		}
		log.Printf("✅ Event updated with image URL")
	} else {
		log.Printf("ℹ️ No image provided for event %s", event.ID)
	}

	return event, nil
}

// ============================================================
// GET EVENT
// ============================================================

func (s *eventService) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	if id == "" {
		return nil, errors.New("event ID is required")
	}
	return s.repo.GetEventByID(ctx, id)
}

func (s *eventService) GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error) {
	if slug == "" {
		return nil, errors.New("event slug is required")
	}
	return s.repo.GetEventBySlug(ctx, slug)
}

// ============================================================
// UPDATE EVENT (with validation)
// ============================================================


func (s *eventService) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error) {
	
    if cmd.ID == "" {
        return nil, errors.New("event ID is required")
    }

    // ✅ Use GetEventByID to find event
    event, err := s.repo.GetEventByID(ctx, cmd.ID)
    if err != nil {
        return nil, err
    }
    if event == nil {
        return nil, domain.ErrEventNotFound
    }

    if !s.permChecker.CanUpdateEvent(ctx, cmd.UpdatedBy, event.AccountID) {
        return nil, errors.New("insufficient permissions to update this event")
    }

    // ✅ Build updated values for validation - using pointers
    var name string
    if cmd.Name != nil {
        name = *cmd.Name
    }
    if name == "" {
        name = event.Name
    }

    var eventTypeID string
    if cmd.EventTypeID != nil {
        eventTypeID = *cmd.EventTypeID
    }
    if eventTypeID == "" {
        eventTypeID = event.EventTypeID
    }

    var date string
    if cmd.Date != nil {
        date = *cmd.Date
    }
    if date == "" {
        date = event.Date.Format("2006-01-02")
    }

    var eventTime string
    if cmd.Time != nil {
        eventTime = *cmd.Time
    }
    if eventTime == "" {
        eventTime = event.Time
    }

    var duration int
    if cmd.Duration != nil {
        duration = *cmd.Duration
    }
    if duration == 0 {
        duration = event.Duration
    }

    var isVirtual bool
    if cmd.IsVirtual != nil {
        isVirtual = *cmd.IsVirtual
    } else {
        isVirtual = event.IsVirtual
    }

    var location string
    if cmd.Location != nil {
        location = *cmd.Location
    }
    if location == "" {
        location = event.Location
    }

    var zoomLink string
    if cmd.ZoomLink != nil {
        zoomLink = *cmd.ZoomLink
    }
    if zoomLink == "" {
        zoomLink = event.ZoomLink
    }

    var meetLink string
    if cmd.MeetLink != nil {
        meetLink = *cmd.MeetLink
    }
    if meetLink == "" {
        meetLink = event.MeetLink
    }

    // ✅ Check if the event is a draft
    isDraft := event.EventStatusID != "" && s.isDraftStatus(ctx, event.EventStatusID)

    // ✅ Relax validation for drafts
    if !isDraft {
        // Strict validation for published events
        if name == "" {
            return nil, errors.New("event name is required")
        }
        if eventTypeID == "" {
            return nil, errors.New("event type is required")
        }
        if date == "" {
            return nil, errors.New("event date is required")
        }
        if eventTime == "" {
            return nil, errors.New("event time is required")
        }
        if duration <= 0 {
            return nil, errors.New("event duration must be greater than 0")
        }
        if duration < 15 {
            return nil, errors.New("duration must be at least 15 minutes")
        }
        if duration > 1440 {
            return nil, errors.New("duration cannot exceed 1440 minutes (24 hours)")
        }
        if !isVirtual && location == "" {
            return nil, errors.New("location is required for in-person events")
        }
        if isVirtual && zoomLink == "" && meetLink == "" {
            return nil, errors.New("at least one meeting link is required for virtual events")
        }
    } else {
        // ✅ Relaxed validation for drafts
        if name == "" {
            name = "Untitled Event"
        }
        if eventTypeID != "" {
            eventType, err := s.repo.GetEventTypeByID(ctx, eventTypeID)
            if err != nil {
                return nil, err
            }
            if eventType == nil {
                return nil, domain.ErrEventTypeNotFound
            }
        }
        if duration > 1440 {
            return nil, errors.New("duration cannot exceed 1440 minutes (24 hours)")
        }
    }

    // ✅ If event type is being updated, validate it exists
    if cmd.EventTypeID != nil && *cmd.EventTypeID != "" && *cmd.EventTypeID != event.EventTypeID {
        eventType, err := s.repo.GetEventTypeByID(ctx, *cmd.EventTypeID)
        if err != nil {
            return nil, err
        }
        if eventType == nil {
            return nil, domain.ErrEventTypeNotFound
        }
        event.EventTypeID = *cmd.EventTypeID
    }

    // ✅ Apply updates - only if provided
    if cmd.Name != nil && *cmd.Name != "" {
        event.Name = *cmd.Name
    }
    if cmd.Description != nil && *cmd.Description != "" {
        event.Description = *cmd.Description
    }
    if cmd.Date != nil && *cmd.Date != "" {
        event.Date = parseDate(*cmd.Date)
    }
    if cmd.Time != nil && *cmd.Time != "" {
        event.Time = *cmd.Time
    }
    if cmd.Duration != nil && *cmd.Duration > 0 {
        event.Duration = *cmd.Duration
    }
    if cmd.Price != nil && *cmd.Price >= 0 {
        event.Price = *cmd.Price
    }
    if cmd.CertificatePrice != nil && *cmd.CertificatePrice >= 0 {
        event.CertificatePrice = *cmd.CertificatePrice
    }
    if cmd.Location != nil && *cmd.Location != "" {
        event.Location = *cmd.Location
    }
    if cmd.IsVirtual != nil {
        event.IsVirtual = *cmd.IsVirtual
    }
    if cmd.ZoomLink != nil && *cmd.ZoomLink != "" {
        event.ZoomLink = *cmd.ZoomLink
    }
    if cmd.MeetLink != nil && *cmd.MeetLink != "" {
        event.MeetLink = *cmd.MeetLink
    }
    if cmd.MaxAttendees != nil && *cmd.MaxAttendees > 0 {
        event.MaxAttendees = *cmd.MaxAttendees
    }

    // ✅ FIX: Only update feature flags if they were provided
    if cmd.IsFeatured != nil {
        event.IsFeatured = *cmd.IsFeatured
    }
    if cmd.IsPrivate != nil {
        event.IsPrivate = *cmd.IsPrivate
    }

    // ✅ If status is being updated, validate it
    if cmd.EventStatusID != nil && *cmd.EventStatusID != "" {
        status, err := s.repo.GetEventStatusByID(ctx, *cmd.EventStatusID)
        if err != nil {
            return nil, err
        }
        if status == nil {
            return nil, domain.ErrEventStatusNotFound
        }
        event.EventStatusID = *cmd.EventStatusID
    }

    // ✅ Update timestamp
    currentTime := time.Now()
    event.UpdatedAt = currentTime

    if err := s.repo.UpdateEvent(ctx, event); err != nil {
        return nil, fmt.Errorf("failed to update event: %w", err)
    }

    log.Printf("✅ Event updated: %s by %s", event.ID, cmd.UpdatedBy)
    return event, nil
}

// Helper function to check if a status ID is a draft
func (s *eventService) isDraftStatus(ctx context.Context, statusID string) bool {
    status, err := s.repo.GetEventStatusByID(ctx, statusID)
    if err != nil || status == nil {
        return false
    }
    return status.Slug == "draft"
}

// ============================================================
// DELETE - Single Event
// ============================================================

// Soft delete - sets deleted_at timestamp
func (s *eventService) DeleteEvent(ctx context.Context, id, deletedBy string) error {
	if id == "" {
		return errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	if !s.permChecker.CanDeleteEvent(ctx, deletedBy, event.AccountID) {
		return errors.New("insufficient permissions to delete this event")
	}

	// ✅ Soft delete - set deleted_at
	if err := event.SoftDelete(deletedBy); err != nil {
		return err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	log.Printf("✅ Event soft deleted: %s by %s", id, deletedBy)
	return nil
}

// internal/modules/events/service/service_impl.go

// Hard delete - permanently removes from database
func (s *eventService) PermanentlyDeleteEvent(ctx context.Context, id, deletedBy string) error {
    if id == "" {
        return errors.New("event ID is required")
    }

    // ✅ Use GetEventByIDIncludingDeleted to find soft-deleted events
    event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
    if err != nil {
        return err
    }
    if event == nil {
        return domain.ErrEventNotFound
    }

    if !s.permChecker.CanDeleteEvent(ctx, deletedBy, event.AccountID) {
        return errors.New("insufficient permissions to delete this event")
    }

    // Delete media files
    if err := s.mediaSvc.DeleteMediaByEntity(ctx, id); err != nil {
        log.Printf("⚠️ Failed to delete media for event %s: %v", id, err)
        // Continue with deletion
    }

    // Hard delete - remove from database
    if err := s.repo.PermanentlyDeleteEvent(ctx, id); err != nil {
        return fmt.Errorf("failed to permanently delete event: %w", err)
    }

    log.Printf("✅ Event permanently deleted: %s by %s", id, deletedBy)
    return nil
}

// Restore soft-deleted event
func (s *eventService) RestoreEvent(ctx context.Context, id, restoredBy string) (*domain.Event, error) {
	if id == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, restoredBy, event.AccountID) {
		return nil, errors.New("insufficient permissions to restore this event")
	}

	if err := event.Restore(restoredBy); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to restore event: %w", err)
	}

	log.Printf("✅ Event restored: %s by %s", id, restoredBy)
	return event, nil
}

// ============================================================
// DELETE - Bulk Events
// ============================================================

// Soft delete multiple events
func (s *eventService) DeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.DeleteEvent(ctx, id, deletedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// Hard delete multiple events
func (s *eventService) PermanentlyDeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.PermanentlyDeleteEvent(ctx, id, deletedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// Restore multiple soft-deleted events
func (s *eventService) RestoreEvents(ctx context.Context, ids []string, restoredBy string) (*BulkRestoreResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkRestoreResult{
		RestoredCount: 0,
		FailedIDs:     []string{},
		Errors:        []string{},
	}

	for _, id := range ids {
		if _, err := s.RestoreEvent(ctx, id, restoredBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.RestoredCount++
	}

	return result, nil
}

// Delete all events for an account (soft delete)
func (s *eventService) DeleteEventsByAccount(ctx context.Context, accountID string, deletedBy string) (*BulkDeleteResult, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	// Get all events for this account
	events, _, err := s.repo.GetEventsByAccount(ctx, accountID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for account: %w", err)
	}

	if len(events) == 0 {
		return &BulkDeleteResult{DeletedCount: 0}, nil
	}

	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.ID
	}

	return s.DeleteEvents(ctx, ids, deletedBy)
}

// Permanently delete all events for an account
func (s *eventService) PermanentlyDeleteEventsByAccount(ctx context.Context, accountID string, deletedBy string) (*BulkDeleteResult, error) {
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}

	// Get all events for this account (including soft-deleted)
	events, _, err := s.repo.GetEventsByAccount(ctx, accountID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for account: %w", err)
	}

	if len(events) == 0 {
		return &BulkDeleteResult{DeletedCount: 0}, nil
	}

	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.ID
	}

	return s.PermanentlyDeleteEvents(ctx, ids, deletedBy)
}

// ============================================================
// STATUS - Single
// ============================================================

func (s *eventService) PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
	log.Printf("📤 Publishing event: %s", id)

	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, publishedBy, event.AccountID) {
		return nil, errors.New("insufficient permissions to publish this event")
	}

	if err := event.Publish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("✅ Event published successfully: %s", id)
	return event, nil
}

func (s *eventService) CancelEvent(ctx context.Context, id, cancelledBy string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, cancelledBy, event.AccountID) {
		return nil, errors.New("insufficient permissions to cancel this event")
	}

	if err := event.Cancel(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to cancel event: %w", err)
	}

	return event, nil
}

func (s *eventService) CompleteEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := event.Complete(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to complete event: %w", err)
	}

	return event, nil
}

// ============================================================
// STATUS - Bulk
// ============================================================

func (s *eventService) BulkPublishEvents(ctx context.Context, ids []string, publishedBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.PublishEvent(ctx, id, publishedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.ProcessedCount++
	}

	return result, nil
}

func (s *eventService) BulkCancelEvents(ctx context.Context, ids []string, cancelledBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.CancelEvent(ctx, id, cancelledBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.ProcessedCount++
	}

	return result, nil
}

func (s *eventService) BulkCompleteEvents(ctx context.Context, ids []string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.CompleteEvent(ctx, id); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.ProcessedCount++
	}

	return result, nil
}

// ============================================================
// DUPLICATE - Single Event
// ============================================================

func (s *eventService) DuplicateEvent(ctx context.Context, id string, cmd DuplicateEventCommand) (*domain.Event, error) {
	log.Printf("📋 Duplicating event: %s", id)

	// Get original event
	original, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, domain.ErrEventNotFound
	}

	// Check permission
	if !s.permChecker.CanManageEvent(ctx, "", original.AccountID) {
		return nil, errors.New("insufficient permissions to duplicate this event")
	}

	// Prepare new event name
	newName := cmd.Name
	if newName == "" {
		newName = fmt.Sprintf("Copy of %s", original.Name)
	}

	// Prepare new date (use original if not provided)
	newDate := cmd.Date
	if newDate == "" {
		newDate = original.Date.Format("2006-01-02")
	}

	// Use original time
	newTime := original.Time

	// Create new event as draft
	createCmd := CreateDraftCommand{
		Name:             newName,
		Description:      original.Description,
		EventTypeID:      original.EventTypeID,
		AccountID:        original.AccountID,
		CreatedBy:        "", // Will be set by auth
		Date:             newDate,
		Time:             newTime,
		Duration:         original.Duration,
		Price:            original.Price,
		CertificatePrice: original.CertificatePrice,
		Location:         original.Location,
		IsVirtual:        original.IsVirtual,
		ZoomLink:         original.ZoomLink,
		MeetLink:         original.MeetLink,
		MaxAttendees:     original.MaxAttendees,
		// ✅ Preserve feature flags
		IsFeatured: original.IsFeatured,
		IsPrivate:  original.IsPrivate,
	}

	// Create the draft
	newEvent, err := s.CreateDraft(ctx, createCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate event: %w", err)
	}

	log.Printf("✅ Event duplicated successfully: %s -> %s", id, newEvent.ID)
	return newEvent, nil
}

// ============================================================
// DUPLICATE - Bulk Events
// ============================================================

func (s *eventService) BulkDuplicateEvents(ctx context.Context, ids []string, cmd BulkDuplicateCommand) (*BulkDuplicateResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDuplicateResult{
		DuplicatedCount: 0,
		CreatedEvents:   []DuplicatedEvent{},
		FailedIDs:       []string{},
		Errors:          []string{},
	}

	for _, id := range ids {
		// Create duplicate command for each event
		dupCmd := DuplicateEventCommand{
			Name:    cmd.NamePrefix,
			IsDraft: cmd.IsDraft,
			Date:    "", // Will be set below if offset provided
		}

		// If date offset is provided, calculate new date
		if cmd.DateOffsetDays != 0 {
			event, err := s.repo.GetEventByID(ctx, id)
			if err != nil {
				result.FailedIDs = append(result.FailedIDs, id)
				result.Errors = append(result.Errors, fmt.Sprintf("event %s: failed to get event: %v", id, err))
				continue
			}
			if event != nil {
				newDate := event.Date.AddDate(0, 0, cmd.DateOffsetDays)
				dupCmd.Date = newDate.Format("2006-01-02")
			}
		}

		event, err := s.DuplicateEvent(ctx, id, dupCmd)
		if err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}

		result.CreatedEvents = append(result.CreatedEvents, DuplicatedEvent{
			ID:   event.ID,
			Name: event.Name,
			Slug: event.Slug,
		})
		result.DuplicatedCount++
	}

	return result, nil
}

// ============================================================
// MEDIA - Upload
// ============================================================

func (s *eventService) UploadEventImage(ctx context.Context, cmd UploadEventImageCommand) (*domain.MediaInfo, error) {
	if cmd.EventID == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, cmd.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, cmd.UploadedBy, event.AccountID) {
		return nil, errors.New("insufficient permissions to upload image for this event")
	}

	// Use pure domain types
	media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
		File:          cmd.ImageData,
		FileName:      cmd.ImageName,
		ContentType:   cmd.ContentType,
		MediaTypeName: "Event",
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Update event with image URL
	event.ImageURL = media.URL
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to update event %s with image URL: %v", cmd.EventID, err)
		return nil, fmt.Errorf("failed to update event with image URL: %w", err)
	}
	log.Printf("✅ Event %s updated with image URL", cmd.EventID)

	return media, nil
}

func (s *eventService) UploadCertificateTemplate(ctx context.Context, cmd UploadCertificateCommand) (*domain.MediaInfo, error) {
	if cmd.EventID == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, cmd.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, cmd.UploadedBy, event.AccountID) {
		return nil, errors.New("insufficient permissions to upload certificate template")
	}

	// Use pure domain types
	media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
		File:          cmd.CertificateData,
		FileName:      cmd.CertificateName,
		ContentType:   cmd.ContentType,
		MediaTypeName: "Certificate",
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate template: %w", err)
	}

	return media, nil
}

// ============================================================
// MEDIA - Delete Single
// ============================================================

func (s *eventService) DeleteEventCertificate(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, deletedBy, event.AccountID) {
		return errors.New("insufficient permissions to delete certificate for this event")
	}

	// Get media type for Certificate
	mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "Certificate")
	if err != nil {
		return fmt.Errorf("failed to get media type: %w", err)
	}

	// Delete all Certificate media for this event
	if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, eventID, mediaType.ID); err != nil {
		return fmt.Errorf("failed to delete certificate template: %w", err)
	}

	log.Printf("✅ Certificate template deleted for event: %s", eventID)
	return nil
}

// DeleteAllEventMedia - Delete all media for an event
func (s *eventService) DeleteAllEventMedia(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, deletedBy, event.AccountID) {
		return errors.New("insufficient permissions to delete media for this event")
	}

	// Use DeleteMediaByEntity
	if err := s.mediaSvc.DeleteMediaByEntity(ctx, eventID); err != nil {
		return fmt.Errorf("failed to delete all media: %w", err)
	}

	// Clear image URL from event
	event.ImageURL = ""
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("⚠️ Failed to update event %s after media deletion: %v", eventID, err)
		// Continue - media is already deleted
	}

	log.Printf("✅ All media deleted for event: %s", eventID)
	return nil
}

// DeleteEventImage - Delete only Event media type for an event
func (s *eventService) DeleteEventImage(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	if !s.permChecker.CanUpdateEvent(ctx, deletedBy, event.AccountID) {
		return errors.New("insufficient permissions to delete image for this event")
	}

	// Use DeleteMediaByEntity - this deletes ALL media for the event
	if err := s.mediaSvc.DeleteMediaByEntity(ctx, eventID); err != nil {
		return fmt.Errorf("failed to delete event image: %w", err)
	}

	// Clear image URL from event
	event.ImageURL = ""
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("⚠️ Failed to update event %s after image deletion: %v", eventID, err)
	}

	log.Printf("✅ Event image deleted for event: %s", eventID)
	return nil
}

// ============================================================
// MEDIA - Delete Bulk
// ============================================================

func (s *eventService) BulkDeleteEventMedia(ctx context.Context, eventIDs []string, deletedBy string) (*BulkDeleteResult, error) {
	if len(eventIDs) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, eventID := range eventIDs {
		if err := s.DeleteAllEventMedia(ctx, eventID, deletedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, eventID)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", eventID, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// ============================================================
// LIST EVENTS
// ============================================================

func (s *eventService) ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error) {
	fmt.Printf("DEBUG Service: OnlyDeleted=%v, AccountID=%s\n", filters.OnlyDeleted, filters.AccountID)
	domainFilters := domain.ListEventsFilters{
		AccountID:      filters.AccountID,
		EventTypeID:    filters.EventTypeID,
		EventStatusID:  filters.EventStatusID,
		IncludeDeleted: filters.IncludeDeleted,
		OnlyDeleted:    filters.OnlyDeleted,  // ✅ NEW
		Limit:          filters.Limit,
		Offset:         filters.Offset,
	}
	return s.repo.ListEvents(ctx, domainFilters)
}

func (s *eventService) GetEventsByType(ctx context.Context, eventTypeSlug string, page, pageSize int) ([]*domain.Event, int64, error) {
	if eventTypeSlug == "" {
		return nil, 0, errors.New("event type slug is required")
	}

	eventType, err := s.repo.GetEventTypeBySlug(ctx, eventTypeSlug)
	if err != nil {
		return nil, 0, err
	}
	if eventType == nil {
		return nil, 0, domain.ErrEventTypeNotFound
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	return s.repo.GetEventsByType(ctx, eventType.ID, limit, offset)
}

func (s *eventService) GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.GetUpcomingEvents(ctx, limit)
}

func (s *eventService) GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.GetPastEvents(ctx, limit)
}

func (s *eventService) SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error) {
	if query == "" && filters.EventTypeID == "" && filters.AccountID == "" {
		return nil, 0, errors.New("search query or filter is required")
	}

	domainFilters := domain.SearchFilters{
		AccountID:      filters.AccountID,
		EventTypeID:    filters.EventTypeID,
		IncludeDeleted: filters.IncludeDeleted,
		OnlyDeleted:    filters.OnlyDeleted, 
		Limit:          filters.Limit,
		Offset:         filters.Offset,
	}
	return s.repo.SearchEvents(ctx, query, domainFilters)
}

func (s *eventService) GetEventsByAccount(ctx context.Context, accountID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if accountID == "" {
		return nil, 0, errors.New("account ID is required")
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.repo.GetEventsByAccount(ctx, accountID, limit, offset)
}

// ============================================================
// EVENT TYPES & STATUSES - Return entities (with IDs)
// ============================================================

// GetEventTypes returns all event types (entities with IDs)
func (s *eventService) GetEventTypes(ctx context.Context) ([]*domain.EventType, error) {
	return s.repo.GetAllEventTypes(ctx)
}

// GetEventStatuses returns all event statuses (entities with IDs)
func (s *eventService) GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error) {
	return s.repo.GetAllEventStatuses(ctx)
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}