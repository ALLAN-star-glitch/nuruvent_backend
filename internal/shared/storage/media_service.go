package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// MEDIA SERVICE
// ============================================================

type MediaService struct {
	mediaRepo     *MediaRepository
	mediaTypeRepo *MediaTypeRepository
	storage       *Client
}

func NewMediaService(
	mediaRepo *MediaRepository,
	mediaTypeRepo *MediaTypeRepository,
	storage *Client,
) *MediaService {
	return &MediaService{
		mediaRepo:     mediaRepo,
		mediaTypeRepo: mediaTypeRepo,
		storage:       storage,
	}
}

// ============================================================
// UPLOAD OPERATIONS
// ============================================================

// UploadFile uploads a file to storage and creates a media record
func (s *MediaService) UploadFile(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	mediaTypeName string,
	entityID uuid.UUID,
	uploadedBy uuid.UUID,
) (*MediaModel, error) {
	// Get media type
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

	// Generate file path
	filePath := s.generateFilePath(mediaTypeName, entityID, fileHeader.Filename)

	// Read file
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	fileReader := bytes.NewReader(fileContent)

	// Upload to storage
	publicURL, err := s.storage.UploadFile(ctx, mediaType.Bucket, filePath, fileReader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create media record
	media := &MediaModel{
		Slug:        generateSlug(fileHeader.Filename),
		Name:        fileHeader.Filename,
		DisplayName: fileHeader.Filename,
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
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	return s.mediaRepo.GetByID(media.ID)
}

// ============================================================
// DELETE OPERATIONS
// ============================================================

func (s *MediaService) DeleteFile(ctx context.Context, mediaID uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return fmt.Errorf("media not found")
	}

	if err := s.storage.DeleteFile(ctx, media.Bucket, media.Path); err != nil {
		// Log but continue
	}

	if err := s.mediaRepo.Delete(mediaID); err != nil {
		return fmt.Errorf("failed to delete media record: %w", err)
	}

	return nil
}

func (s *MediaService) DeleteFilesByEntity(ctx context.Context, entityID uuid.UUID) error {
	mediaList, _, err := s.mediaRepo.GetByEntity(entityID, 0, 0)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		_ = s.DeleteFile(ctx, media.ID)
	}

	return nil
}

func (s *MediaService) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID uuid.UUID) error {
	mediaList, err := s.mediaRepo.GetByEntityAndMediaType(entityID, mediaTypeID)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		_ = s.DeleteFile(ctx, media.ID)
	}

	return nil
}

// ============================================================
// QUERY OPERATIONS
// ============================================================

func (s *MediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*MediaModel, error) {
	return s.mediaRepo.GetByID(id)
}

func (s *MediaService) GetMediaByEntity(ctx context.Context, entityID uuid.UUID, page, pageSize int) ([]MediaModel, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.mediaRepo.GetByEntity(entityID, limit, offset)
}

func (s *MediaService) GetMediaByEntityAndType(ctx context.Context, entityID uuid.UUID, mediaTypeName string) ([]MediaModel, error) {
	mediaType, err := s.mediaTypeRepo.GetByName(mediaTypeName)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, fmt.Errorf("invalid media type: %s", mediaTypeName)
	}

	return s.mediaRepo.GetByEntityAndMediaType(entityID, mediaType.ID)
}

func (s *MediaService) GetImagesByEntity(ctx context.Context, entityID uuid.UUID) ([]MediaModel, error) {
	return s.mediaRepo.GetByEntityAndMimeType(entityID, "image")
}

func (s *MediaService) GetPrimaryMediaForEntity(ctx context.Context, entityID uuid.UUID) (*MediaModel, error) {
	return s.mediaRepo.GetPrimaryMediaForEntity(entityID)
}

func (s *MediaService) GetMediaByUploader(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]MediaModel, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.mediaRepo.GetByUploader(userID, limit, offset)
}

func (s *MediaService) GetMediaByType(ctx context.Context, mediaTypeName string, page, pageSize int) ([]MediaModel, int64, error) {
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

// ============================================================
// MEDIA TYPE OPERATIONS
// ============================================================

func (s *MediaService) GetMediaTypes(ctx context.Context) ([]MediaTypeModel, error) {
	return s.mediaTypeRepo.GetAll()
}

func (s *MediaService) GetMediaTypeByName(ctx context.Context, name string) (*MediaTypeModel, error) {
	return s.mediaTypeRepo.GetByName(name)
}

// ============================================================
// HELPER METHODS
// ============================================================

func (s *MediaService) generateFilePath(mediaTypeName string, entityID uuid.UUID, fileName string) string {
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

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

func (s *MediaService) isAllowedMimeType(mimeType string, mediaType *MediaTypeModel) bool {
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

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, `"`, "")

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	if len(slug) > 50 {
		slug = slug[:50]
	}
	return slug + "-" + uuid.New().String()[:8]
}