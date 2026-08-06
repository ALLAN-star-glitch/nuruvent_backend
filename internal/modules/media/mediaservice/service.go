// internal/modules/media/mediaservice/service.go

package mediaservice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media/mediarepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/storage"
	"github.com/google/uuid"
)

type MediaService struct {
	mediaRepo     *mediarepo.MediaRepository
	mediaTypeRepo *mediarepo.MediaTypeRepository
	storage       *storage.Client
}

func NewMediaService(
	mediaRepo *mediarepo.MediaRepository,
	mediaTypeRepo *mediarepo.MediaTypeRepository,
	storage *storage.Client,
) *MediaService {
	return &MediaService{
		mediaRepo:     mediaRepo,
		mediaTypeRepo: mediaTypeRepo,
		storage:       storage,
	}
}

// ================================================
// UPLOAD OPERATIONS
// ================================================

// UploadFile uploads a file to Supabase Storage and creates a media record
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) UploadFile(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	mediaTypeName string,
	entityID uuid.UUID,
	uploadedBy uuid.UUID,
) (*models.Media, error) {
	// Get media type from database
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get media type: %w", err)
	}
	if mediaType == nil {
		return nil, fmt.Errorf("invalid media type: %s", mediaTypeName)
	}

	// Validate file size
	if fileHeader.Size > mediaType.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mediaType.MaxFileSize)
	}

	// Validate mime type
	contentType := fileHeader.Header.Get("Content-Type")
	if !s.isAllowedMimeType(contentType, mediaType) {
		return nil, fmt.Errorf("file type %s is not allowed for %s", contentType, mediaTypeName)
	}

	// Generate unique file path
	filePath := s.generateFilePath(mediaTypeName, entityID, fileHeader.Filename)

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert []byte to io.Reader
	fileReader := bytes.NewReader(fileContent)

	// Upload to Supabase Storage
	publicURL, err := s.storage.UploadFile(ctx, mediaType.Bucket, filePath, fileReader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create media record
	media := &models.Media{
		FileName:    fileHeader.Filename,
		FileSize:    fileHeader.Size,
		MimeType:    contentType,
		Bucket:      mediaType.Bucket,
		Path:        filePath,
		URL:         publicURL,
		MediaTypeID: mediaType.ID,
		EntityID:    entityID,
		UploadedBy:  uploadedBy,
		IsActive:    true,
	}

	if err := s.mediaRepo.Create(media); err != nil {
		// Attempt to delete the uploaded file if DB record creation fails
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	// Reload with MediaType preloaded
	return s.mediaRepo.GetByID(media.ID)
}

// ================================================
// DELETE OPERATIONS
// ================================================

// DeleteFile deletes a file from storage and the media record
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) DeleteFile(ctx context.Context, mediaID uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return fmt.Errorf("media not found")
	}

	// Delete from Supabase Storage
	if err := s.storage.DeleteFile(ctx, media.Bucket, media.Path); err != nil {
		// Log error but continue - file might already be deleted
	}

	// Soft delete from database
	if err := s.mediaRepo.Delete(mediaID); err != nil {
		return fmt.Errorf("failed to delete media record: %w", err)
	}

	return nil
}

// DeleteFilesByEntity deletes all media for an entity
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) DeleteFilesByEntity(ctx context.Context, entityID uuid.UUID) error {
	mediaList, _, err := s.mediaRepo.GetByEntity(entityID, 0, 0)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		if err := s.DeleteFile(ctx, media.ID); err != nil {
			// Log error but continue with other files
			continue
		}
	}

	return nil
}

// DeleteFilesByEntityAndMediaType deletes all media for an entity and media type
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID uuid.UUID) error {
	mediaList, err := s.mediaRepo.GetByEntityAndMediaType(entityID, mediaTypeID)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		if err := s.DeleteFile(ctx, media.ID); err != nil {
			// Log error but continue with other files
			continue
		}
	}

	return nil
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetMediaByID gets a media record by ID
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*models.Media, error) {
	return s.mediaRepo.GetByID(id)
}

// GetMediaByEntity gets all media for an entity with pagination
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetMediaByEntity(ctx context.Context, entityID uuid.UUID, page, pageSize int) ([]models.Media, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.mediaRepo.GetByEntity(entityID, limit, offset)
}

// GetMediaByEntityAndType gets media by entity and media type name
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetMediaByEntityAndType(ctx context.Context, entityID uuid.UUID, mediaTypeName string) ([]models.Media, error) {
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, fmt.Errorf("invalid media type: %s", mediaTypeName)
	}

	return s.mediaRepo.GetByEntityAndMediaType(entityID, mediaType.ID)
}

// GetImagesByEntity gets all images for an entity
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetImagesByEntity(ctx context.Context, entityID uuid.UUID) ([]models.Media, error) {
	return s.mediaRepo.GetByEntityAndMimeType(entityID, "image")
}

// GetPrimaryMediaForEntity gets the primary media for an entity
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetPrimaryMediaForEntity(ctx context.Context, entityID uuid.UUID) (*models.Media, error) {
	return s.mediaRepo.GetPrimaryMediaForEntity(entityID)
}

// GetMediaByUploader gets all media uploaded by a user
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetMediaByUploader(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Media, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.mediaRepo.GetByUploader(userID, limit, offset)
}

// GetMediaByType gets all media of a specific type
// Note: Authorization should be handled by the caller (business layer)
func (s *MediaService) GetMediaByType(ctx context.Context, mediaTypeName string, page, pageSize int) ([]models.Media, int64, error) {
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return nil, 0, err
	}
	if mediaType == nil {
		return nil, 0, fmt.Errorf("invalid media type: %s", mediaTypeName)
	}

	limit := pageSize
	offset := (page - 1) * pageSize
	return s.mediaRepo.GetByMediaType(mediaType.ID, limit, offset)
}

// ================================================
// MEDIA TYPE OPERATIONS
// ================================================

// GetMediaTypes gets all media types
func (s *MediaService) GetMediaTypes(ctx context.Context) ([]models.MediaType, error) {
	return s.mediaTypeRepo.GetAll()
}

// GetMediaTypeByName gets a media type by name
func (s *MediaService) GetMediaTypeByName(ctx context.Context, name string) (*models.MediaType, error) {
	return s.mediaTypeRepo.GetByName(name)
}

// ================================================
// HELPER METHODS
// ================================================

// generateFilePath generates a unique file path for storage
func (s *MediaService) generateFilePath(mediaTypeName string, entityID uuid.UUID, fileName string) string {
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Sanitize filename
	safeName := strings.ToLower(nameWithoutExt)
	safeName = strings.ReplaceAll(safeName, " ", "-")
	safeName = strings.ReplaceAll(safeName, "_", "-")
	safeName = strings.ReplaceAll(safeName, "[", "")
	safeName = strings.ReplaceAll(safeName, "]", "")
	safeName = strings.ReplaceAll(safeName, "(", "")
	safeName = strings.ReplaceAll(safeName, ")", "")
	safeName = strings.ReplaceAll(safeName, "'", "")
	safeName = strings.ReplaceAll(safeName, `"`, "")
	safeName = strings.ReplaceAll(safeName, "&", "and")

	// Limit length
	if len(safeName) > 50 {
		safeName = safeName[:50]
	}

	return fmt.Sprintf("%s/%s/%d-%s-%s%s",
		mediaTypeName,
		entityID.String(),
		timestamp,
		safeName,
		uuid.New().String()[:8],
		ext,
	)
}

// isAllowedMimeType checks if the mime type is allowed for the media type
func (s *MediaService) isAllowedMimeType(mimeType string, mediaType *models.MediaType) bool {
	allowedTypes := map[string][]string{
		"event":       {"image/jpeg", "image/png", "image/gif", "image/webp"},
		"business":    {"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"},
		"profile":     {"image/jpeg", "image/png", "image/gif", "image/webp"},
		"certificate": {"image/jpeg", "image/png", "image/webp", "application/pdf"},
		"recording":   {"video/mp4", "video/webm", "video/ogg"},
	}

	allowed, ok := allowedTypes[mediaType.Name]
	if !ok {
		return false
	}

	for _, t := range allowed {
		if mimeType == t {
			return true
		}
	}
	return false
}

// GetFileSizeLimit gets the maximum file size for a media type
func (s *MediaService) GetFileSizeLimit(ctx context.Context, mediaTypeName string) (int64, error) {
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return 0, err
	}
	if mediaType == nil {
		return 0, fmt.Errorf("invalid media type: %s", mediaTypeName)
	}
	return mediaType.MaxFileSize, nil
}

// GetBucketForMediaType gets the bucket name for a media type
func (s *MediaService) GetBucketForMediaType(ctx context.Context, mediaTypeName string) (string, error) {
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return "", err
	}
	if mediaType == nil {
		return "", fmt.Errorf("invalid media type: %s", mediaTypeName)
	}
	return mediaType.Bucket, nil
}