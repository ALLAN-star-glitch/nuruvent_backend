// internal/modules/events/service/media.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// MEDIA - Upload
// ============================================================

// UploadEventImage uploads an image for an event
func (s *eventService) UploadEventImage(ctx context.Context, cmd UploadEventImageCommand) (*MediaInfo, error) {
	// 1. Validate and get event
	if cmd.EventID == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.getEventAndCheckUpdatePermission(ctx, cmd.EventID, cmd.UploadedBy)
	if err != nil {
		return nil, err
	}

	// 2. Upload image
	media, err := s.uploadMedia(ctx, domain.UploadMediaCommand{
		File:          cmd.ImageData,
		FileName:      cmd.ImageName,
		ContentType:   cmd.ContentType,
		MediaTypeName: types.MediaTypeEvent.GetName(),
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// 3. Update event with image URL
	if err := s.updateEventImageURL(ctx, event, media.URL); err != nil {
		return nil, err
	}

	return &MediaInfo{
		ID:          media.ID,
		URL:         media.URL,
		Filename:    cmd.ImageName,
		Size:        int64(len(cmd.ImageData)),
		ContentType: cmd.ContentType,
		UploadedAt:  media.CreatedAt,
	}, nil
}

// UploadCertificateTemplate uploads a certificate template for an event
func (s *eventService) UploadCertificateTemplate(ctx context.Context, cmd UploadCertificateCommand) (*MediaInfo, error) {
	// 1. Validate and get event
	if cmd.EventID == "" {
		return nil, errors.New("event ID is required")
	}

	if _, err := s.getEventAndCheckUpdatePermission(ctx, cmd.EventID, cmd.UploadedBy); err != nil {
		return nil, err
	}

	// 2. Upload certificate
	media, err := s.uploadMedia(ctx, domain.UploadMediaCommand{
		File:          cmd.CertificateData,
		FileName:      cmd.CertificateName,
		ContentType:   cmd.ContentType,
		MediaTypeName: types.MediaTypeCertificate.GetName(),
		EntityID:      cmd.EventID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate template: %w", err)
	}

	return &MediaInfo{
		ID:          media.ID,
		URL:         media.URL,
		Filename:    cmd.CertificateName,
		Size:        int64(len(cmd.CertificateData)),
		ContentType: cmd.ContentType,
		UploadedAt:  media.CreatedAt,
	}, nil
}

// ============================================================
// MEDIA - Delete Single
// ============================================================

// DeleteEventImage deletes the image for an event
func (s *eventService) DeleteEventImage(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	// 1. Validate and get event
	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, deletedBy)
	if err != nil {
		return err
	}

	// 2. Delete image from storage
	mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, types.MediaTypeEvent.GetName())
	if err != nil {
		return fmt.Errorf("failed to get media type: %w", err)
	}

	if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, eventID, mediaType.ID); err != nil {
		return fmt.Errorf("failed to delete event image: %w", err)
	}

	// 3. Clear image URL from event
	if err := s.clearEventImageURL(ctx, event); err != nil {
		return err
	}

	log.Printf("✅ Event image deleted for event: %s", eventID)
	return nil
}

// DeleteEventCertificate deletes the certificate template for an event
func (s *eventService) DeleteEventCertificate(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	// 1. Validate and get event
	if _, err := s.getEventAndCheckUpdatePermission(ctx, eventID, deletedBy); err != nil {
		return err
	}

	// 2. Delete certificate from storage
	mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, types.MediaTypeCertificate.GetName())
	if err != nil {
		return fmt.Errorf("failed to get media type: %w", err)
	}

	if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, eventID, mediaType.ID); err != nil {
		return fmt.Errorf("failed to delete certificate template: %w", err)
	}

	log.Printf("✅ Certificate template deleted for event: %s", eventID)
	return nil
}

// DeleteAllEventMedia deletes all media for an event
func (s *eventService) DeleteAllEventMedia(ctx context.Context, eventID string, deletedBy string) error {
	if eventID == "" {
		return errors.New("event ID is required")
	}

	// 1. Validate and get event
	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, deletedBy)
	if err != nil {
		return err
	}

	// 2. Delete all media from storage
	if err := s.mediaSvc.DeleteFilesByEntity(ctx, eventID); err != nil {
		return fmt.Errorf("failed to delete all media: %w", err)
	}

	// 3. Clear image URL from event
	if err := s.clearEventImageURL(ctx, event); err != nil {
		return err
	}

	log.Printf("✅ All media deleted for event: %s", eventID)
	return nil
}

// ============================================================
// MEDIA - Delete Bulk
// ============================================================

// BulkDeleteEventMedia deletes all media for multiple events
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
// PRIVATE HELPER FUNCTIONS
// ============================================================

// getEventAndCheckUpdatePermission gets event and checks update permission
func (s *eventService) getEventAndCheckUpdatePermission(ctx context.Context, eventID, userID string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	institutionID := s.getInstitutionIDFromEvent(event)
	if !s.permChecker.CanUpdateEvent(ctx, userID, institutionID) {
		return nil, errors.New("insufficient permissions to update this event")
	}

	return event, nil
}

// uploadMedia uploads a file to the media service
func (s *eventService) uploadMedia(ctx context.Context, cmd domain.UploadMediaCommand) (*domain.MediaInfo, error) {
	media, err := s.mediaSvc.UploadFile(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// updateEventImageURL updates the event with the image URL
func (s *eventService) updateEventImageURL(ctx context.Context, event *domain.Event, imageURL string) error {
	event.ImageURL = imageURL
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to update event %s with image URL: %v", event.ID, err)
		return fmt.Errorf("failed to update event with image URL: %w", err)
	}
	return nil
}

// clearEventImageURL clears the image URL from the event
func (s *eventService) clearEventImageURL(ctx context.Context, event *domain.Event) error {
	event.ImageURL = ""
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("⚠️ Failed to update event %s after image deletion: %v", event.ID, err)
		return fmt.Errorf("failed to update event after image deletion: %w", err)
	}
	return nil
}