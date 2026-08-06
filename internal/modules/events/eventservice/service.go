// internal/modules/events/eventservice/service.go

package eventservice

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media/mediaservice"
	"github.com/google/uuid"
)

type EventService struct {
	eventRepo   *eventrepo.EventRepository
	enforcer    *authorization.Enforcer
	permService *authorization.Service
	mediaService *mediaservice.MediaService // Added media service
}

func NewEventService(
	eventRepo *eventrepo.EventRepository,
	enforcer *authorization.Enforcer,
	permService *authorization.Service,
	mediaService *mediaservice.MediaService, // Added parameter
) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		enforcer:     enforcer,
		permService:  permService,
		mediaService: mediaService,
	}
}

// ================================================
// EVENT CRUD WITH MEDIA
// ================================================

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, userID, businessID uuid.UUID, event *models.Event) (*models.Event, error) {
	// Check if user has permission to create events for this business
	canCreate, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(businessID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionCreate.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canCreate {
		return nil, fmt.Errorf("insufficient permissions to create events for this business")
	}

	// Generate slug if not provided
	if event.Slug == "" {
		event.Slug = s.generateSlug(event.Name)
	}

	// Set default status to draft if not set
	if event.EventStatusID == uuid.Nil {
		draftStatus, err := s.eventRepo.GetEventStatusByName("draft")
		if err != nil {
			return nil, fmt.Errorf("failed to get draft status: %w", err)
		}
		if draftStatus == nil {
			return nil, fmt.Errorf("draft status not found")
		}
		event.EventStatusID = draftStatus.ID
	}

	event.CreatedBy = userID
	event.CurrentAttendees = 0
	event.IsVirtual = true

	if err := s.eventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// CreateEventWithImage creates a new event with an image upload
func (s *EventService) CreateEventWithImage(
	ctx context.Context,
	userID, businessID uuid.UUID,
	event *models.Event,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (*models.Event, *models.Media, error) {
	// Create the event first
	createdEvent, err := s.CreateEvent(ctx, userID, businessID, event)
	if err != nil {
		return nil, nil, err
	}

	// Upload image if provided
	var media *models.Media
	if file != nil && fileHeader != nil {
		media, err = s.mediaService.UploadFile(
			ctx,
			file,
			fileHeader,
			"event",
			createdEvent.ID,
			userID,
		)
		if err != nil {
			// Log error but don't fail the event creation
			// Event is already created, just return it without the image
			return createdEvent, nil, nil
		}

		// Update event with image URL
		createdEvent.ImageURL = media.URL
		if err := s.eventRepo.Update(createdEvent); err != nil {
			return createdEvent, media, nil // Event created, image uploaded, but update failed
		}
	}

	return createdEvent, media, nil
}

// UploadEventImage uploads an image for an existing event
func (s *EventService) UploadEventImage(
	ctx context.Context,
	userID, eventID uuid.UUID,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (*models.Media, error) {
	// Get event
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Check authorization
	canUpdate, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(event.InstitutionID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionUpdate.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to update this event")
	}

	// Upload image
	media, err := s.mediaService.UploadFile(
		ctx,
		file,
		fileHeader,
		"event",
		eventID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Update event with image URL
	event.ImageURL = media.URL
	if err := s.eventRepo.Update(event); err != nil {
		return media, fmt.Errorf("image uploaded but failed to update event: %w", err)
	}

	return media, nil
}

// UploadCertificateTemplate uploads a certificate template for an event
func (s *EventService) UploadCertificateTemplate(
	ctx context.Context,
	userID, eventID uuid.UUID,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) (*models.Media, error) {
	// Get event
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Check authorization - only event managers and business admins can upload certificates
	canUpdate, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(event.InstitutionID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionUpdate.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to upload certificate template")
	}

	// Upload certificate template
	media, err := s.mediaService.UploadFile(
		ctx,
		file,
		fileHeader,
		"certificate",
		eventID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate template: %w", err)
	}

	return media, nil
}

// ================================================
// EXISTING METHODS (same as before)
// ================================================

// GetEventByID gets an event by ID with permission check
func (s *EventService) GetEventByID(ctx context.Context, userID, eventID uuid.UUID) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, nil
	}

	canRead, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(event.InstitutionID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionRead.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canRead {
		return nil, fmt.Errorf("insufficient permissions to view this event")
	}

	return event, nil
}

// GetEventByIDPublic gets an event by ID (public access)
func (s *EventService) GetEventByIDPublic(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, nil
	}

	// Only return if event is published
	if event.EventStatus.Name != "published" {
		return nil, fmt.Errorf("event not found")
	}

	return event, nil
}

// GetEventBySlug gets an event by slug (public)
func (s *EventService) GetEventBySlug(ctx context.Context, slug string) (*models.Event, error) {
	event, err := s.eventRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, nil
	}

	if event.EventStatus.Name != "published" {
		return nil, fmt.Errorf("event not found")
	}

	return event, nil
}

// UpdateEvent updates an event
func (s *EventService) UpdateEvent(ctx context.Context, userID, eventID uuid.UUID, updates map[string]interface{}) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	canUpdate, err := s.enforcer.Enforce(
		userID.String(),
	    authorization.AccountDomain(event.InstitutionID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionUpdate.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to update this event")
	}



	// Apply updates (same as before)
	if name, ok := updates["name"].(string); ok && name != "" {
		event.Name = name
		if event.Slug == "" || updates["slug"] == nil {
			event.Slug = s.generateSlug(name)
		}
	}
	if description, ok := updates["description"].(string); ok {
		event.Description = description
	}
	if imageURL, ok := updates["image_url"].(string); ok {
		event.ImageURL = imageURL
	}
	if thumbnailURL, ok := updates["thumbnail_url"].(string); ok {
		event.ThumbnailURL = thumbnailURL
	}
	if eventTypeID, ok := updates["event_type_id"].(string); ok {
		id, err := uuid.Parse(eventTypeID)
		if err == nil {
			event.EventTypeID = id
		}
	}
	if eventStatusID, ok := updates["event_status_id"].(string); ok {
		id, err := uuid.Parse(eventStatusID)
		if err == nil {
			status, err := s.eventRepo.GetEventStatusByID(id)
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %w", err)
			}
			if status == nil {
				return nil, fmt.Errorf("invalid status ID")
			}
			event.EventStatusID = id
		}
	}
	if date, ok := updates["date"].(time.Time); ok {
		event.Date = date
	}
	if timeStr, ok := updates["time"].(string); ok {
		event.Time = timeStr
	}
	if duration, ok := updates["duration"].(int); ok {
		event.Duration = duration
	}
	if price, ok := updates["price"].(float64); ok {
		event.Price = price
	}
	if certificatePrice, ok := updates["certificate_price"].(float64); ok {
		event.CertificatePrice = certificatePrice
	}
	if location, ok := updates["location"].(string); ok {
		event.Location = location
	}
	if zoomLink, ok := updates["zoom_link"].(string); ok {
		event.ZoomLink = zoomLink
	}
	if meetLink, ok := updates["meet_link"].(string); ok {
		event.MeetLink = meetLink
	}
	if maxAttendees, ok := updates["max_attendees"].(int); ok {
		event.MaxAttendees = maxAttendees
	}
	if isVirtual, ok := updates["is_virtual"].(bool); ok {
		event.IsVirtual = isVirtual
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, userID, eventID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	canDelete, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(event.InstitutionID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionDelete.String(),
	)
	if err != nil {
		return fmt.Errorf("authorization error: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("insufficient permissions to delete this event")
	}

	// Delete event image from storage if exists
	if event.ImageURL != "" {
		// Get media record by URL or path
		// Note: You might want to add a method to find media by URL
		// For now, we'll just delete the event
	}

	if err := s.eventRepo.Delete(eventID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetBusinessEvents gets all events for a business
func (s *EventService) GetBusinessEvents(ctx context.Context, userID, businessID uuid.UUID, eventTypeID *uuid.UUID, eventStatusID *uuid.UUID, page, pageSize int) ([]models.Event, int64, error) {
	canRead, err := s.enforcer.Enforce(
		userID.String(),
		authorization.AccountDomain(businessID.String()),
		authorization.ResourceEvent.String(),
		authorization.ActionRead.String(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("authorization error: %w", err)
	}
	if !canRead {
		return nil, 0, fmt.Errorf("insufficient permissions to view events for this business")
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	return s.eventRepo.GetEventsByBusiness(businessID, eventTypeID, eventStatusID, limit, offset)
}

// GetEventsByType gets events by type (public)
func (s *EventService) GetEventsByType(ctx context.Context, eventTypeName string, page, pageSize int) ([]models.Event, int64, error) {
	eventType, err := s.eventRepo.GetEventTypeByName(eventTypeName)
	if err != nil {
		return nil, 0, err
	}
	if eventType == nil {
		return nil, 0, fmt.Errorf("event type not found")
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	return s.eventRepo.GetEventsByType(eventType.ID, limit, offset)
}

// GetUpcomingEvents gets upcoming events (public)
func (s *EventService) GetUpcomingEvents(ctx context.Context, limit int) ([]models.Event, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.eventRepo.GetUpcomingEvents(limit)
}

// GetUpcomingEventsByBusiness gets upcoming events for a business
func (s *EventService) GetUpcomingEventsByBusiness(ctx context.Context, businessID uuid.UUID, limit int) ([]models.Event, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.eventRepo.GetUpcomingEventsByBusiness(businessID, limit)
}

// GetPastEvents gets past events (public)
func (s *EventService) GetPastEvents(ctx context.Context, limit int) ([]models.Event, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.eventRepo.GetPastEvents(limit)
}

// SearchEvents searches events
func (s *EventService) SearchEvents(ctx context.Context, query string, businessID *uuid.UUID, eventTypeID *uuid.UUID, page, pageSize int) ([]models.Event, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize

	return s.eventRepo.SearchEvents(query, businessID, eventTypeID, limit, offset)
}

// ================================================
// EVENT STATUS OPERATIONS
// ================================================

// GetEventStatusByID gets an event status by ID
func (s *EventService) GetEventStatusByID(ctx context.Context, id uuid.UUID) (*models.EventStatus, error) {
	return s.eventRepo.GetEventStatusByID(id)
}

// GetEventStatusByName gets an event status by name
func (s *EventService) GetEventStatusByName(ctx context.Context, name string) (*models.EventStatus, error) {
	return s.eventRepo.GetEventStatusByName(name)
}

// GetAllEventStatuses gets all event statuses
func (s *EventService) GetAllEventStatuses(ctx context.Context) ([]models.EventStatus, error) {
	return s.eventRepo.GetAllEventStatuses()
}

// ================================================
// EVENT TYPE OPERATIONS
// ================================================

// GetEventTypes gets all event types
func (s *EventService) GetEventTypes(ctx context.Context) ([]models.EventType, error) {
	return s.eventRepo.GetAllEventTypes()
}

// GetEventTypeByID gets an event type by ID
func (s *EventService) GetEventTypeByID(ctx context.Context, id uuid.UUID) (*models.EventType, error) {
	return s.eventRepo.GetEventTypeByID(id)
}

// GetEventTypeByName gets an event type by name
func (s *EventService) GetEventTypeByName(ctx context.Context, name string) (*models.EventType, error) {
	return s.eventRepo.GetEventTypeByName(name)
}

// ================================================
// HELPER METHODS
// ================================================

// generateSlug generates a unique slug from a title
func (s *EventService) generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "&", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")

	slug = slug + "-" + uuid.New().String()[:8]

	return slug
}