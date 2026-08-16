// internal/modules/events/service/service.go

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
// CREATE EVENT
// ============================================================

func (s *eventService) CreateEvent(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	// Check permission
	if !s.permChecker.CanManageEvent(ctx, cmd.CreatedBy, cmd.AccountID) {
		return nil, errors.New("insufficient permissions to create events for this account")
	}

	// Get event type to validate and get the type name
	eventType, err := s.repo.GetEventTypeByID(ctx, cmd.EventTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type: %w", err)
	}
	if eventType == nil {
		return nil, domain.ErrEventTypeNotFound
	}

	// Create domain entity
	event, err := domain.NewEvent(
		cmd.Name,
		cmd.Description,
		cmd.EventTypeID,
		cmd.AccountID,
		cmd.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Set additional fields
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

	// Set default status to draft
	draftStatus, err := s.repo.GetEventStatusBySlug(ctx, "draft")
	if err != nil {
		return nil, err
	}
	if draftStatus == nil {
		return nil, domain.ErrEventStatusNotFound
	}
	event.EventStatusID = draftStatus.ID

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// ============================================================
// CREATE EVENT WITH IMAGE
// ============================================================

// internal/modules/events/service/service.go

func (s *eventService) CreateEventWithImage(ctx context.Context, cmd CreateEventWithImageCommand) (*domain.Event, error) {
	// First create the event
	event, err := s.CreateEvent(ctx, cmd.CreateEventCommand)
	if err != nil {
		return nil, err
	}

	// If image is provided, upload it
	if cmd.ImageFile != nil && cmd.ImageHeader != nil {
		media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
			File:          cmd.ImageFile,
			FileHeader:    cmd.ImageHeader,
			MediaTypeName: "Event",
			EntityID:      event.ID,
			UploadedBy:    cmd.CreatedBy,
		})
		if err != nil {
			// Log error but return event without image
			log.Printf("Failed to upload image for event %s: %v", event.ID, err)
			return event, nil
		}

		// Update event with image URL
		event.ImageURL = media.URL
		if err := s.repo.UpdateEvent(ctx, event); err != nil {
			log.Printf("Failed to update event %s with image URL: %v", event.ID, err)
			return event, nil
		}
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
// UPDATE EVENT
// ============================================================

func (s *eventService) UpdateEvent(ctx context.Context, cmd UpdateEventCommand) (*domain.Event, error) {
	if cmd.ID == "" {
		return nil, errors.New("event ID is required")
	}

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

	// Apply updates
	if cmd.Name != "" {
		event.Name = cmd.Name
	}
	if cmd.Description != "" {
		event.Description = cmd.Description
	}
	if cmd.EventTypeID != "" {
		// Validate event type exists
		eventType, err := s.repo.GetEventTypeByID(ctx, cmd.EventTypeID)
		if err != nil {
			return nil, err
		}
		if eventType == nil {
			return nil, domain.ErrEventTypeNotFound
		}
		event.EventTypeID = cmd.EventTypeID
	}
	if cmd.EventStatusID != "" {
		status, err := s.repo.GetEventStatusByID(ctx, cmd.EventStatusID)
		if err != nil {
			return nil, err
		}
		if status == nil {
			return nil, domain.ErrEventStatusNotFound
		}
		event.EventStatusID = cmd.EventStatusID
	}
	if cmd.Date != "" {
		event.Date = parseDate(cmd.Date)
	}
	if cmd.Time != "" {
		event.Time = cmd.Time
	}
	if cmd.Duration > 0 {
		event.Duration = cmd.Duration
	}
	if cmd.Price >= 0 {
		event.Price = cmd.Price
	}
	if cmd.CertificatePrice >= 0 {
		event.CertificatePrice = cmd.CertificatePrice
	}
	if cmd.Location != "" {
		event.Location = cmd.Location
	}
	if cmd.IsVirtual {
		event.IsVirtual = cmd.IsVirtual
	}
	if cmd.ZoomLink != "" {
		event.ZoomLink = cmd.ZoomLink
	}
	if cmd.MeetLink != "" {
		event.MeetLink = cmd.MeetLink
	}
	if cmd.MaxAttendees > 0 {
		event.MaxAttendees = cmd.MaxAttendees
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

// ============================================================
// DELETE EVENT
// ============================================================

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

	// Delete media files
	if err := s.mediaSvc.DeleteMediaByEntity(ctx, id); err != nil {
		// Log but continue
	}

	return s.repo.DeleteEvent(ctx, id)
}

// ============================================================
// PUBLISH EVENT
// ============================================================

func (s *eventService) PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
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
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return event, nil
}

// ============================================================
// CANCEL EVENT
// ============================================================

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

// ============================================================
// COMPLETE EVENT
// ============================================================

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
// LIST EVENTS
// ============================================================

func (s *eventService) ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error) {
	domainFilters := domain.ListEventsFilters{
		AccountID:     filters.AccountID,
		EventTypeID:   filters.EventTypeID,
		EventStatusID: filters.EventStatusID,
		Limit:         filters.Limit,
		Offset:        filters.Offset,
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
		AccountID:   filters.AccountID,
		EventTypeID: filters.EventTypeID,
		Limit:       filters.Limit,
		Offset:      filters.Offset,
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
// MEDIA OPERATIONS
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

	media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
		File:          cmd.ImageFile,
		FileHeader:    cmd.ImageHeader,
		MediaTypeName: "event",
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	event.ImageURL = media.URL
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return media, nil
	}

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

	media, err := s.mediaSvc.UploadFile(ctx, domain.UploadMediaCommand{
		File:          cmd.CertificateFile,
		FileHeader:    cmd.CertificateHeader,
		MediaTypeName: "certificate",
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate template: %w", err)
	}

	return media, nil
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