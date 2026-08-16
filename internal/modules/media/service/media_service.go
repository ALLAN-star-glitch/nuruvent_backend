package service

import (
	"slices"
	"bytes"
	"context"
	"fmt"
	"io"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type mediaService struct {
	repo      domain.Repository
	storage   *storage.Client
}

func NewService(
	repo domain.Repository,
	storageClient *storage.Client,
) Service {
	return &mediaService{
		repo:    repo,
		storage: storageClient,
	}
}

// ============================================================
// UPLOAD
// ============================================================


func (s *mediaService) UploadFile(ctx context.Context, cmd UploadCommand) (*domain.Media, error) {
	// 1. Get media type
	mediaType, err := s.repo.GetMediaTypeByName(ctx, cmd.MediaTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get media type: %w", err)
	}
	if mediaType == nil {
		return nil, domain.ErrMediaTypeNotFound
	}

	// 2. Validate file size
	if cmd.FileHeader.Size > mediaType.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mediaType.MaxFileSize)
	}

	// 3. Validate mime type
	contentType := cmd.FileHeader.Header.Get("Content-Type")
	if !s.isAllowedMimeType(contentType, mediaType) {
		return nil, fmt.Errorf("file type %s is not allowed for %s", contentType, cmd.MediaTypeName)
	}

	// 4. Parse entity ID
	entityID, err := uuid.Parse(cmd.EntityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	// 5. Generate file path
	filePath := s.generateFilePath(cmd.MediaTypeName, entityID, cmd.FileHeader.Filename)

	// 6. Read file
	fileContent, err := io.ReadAll(cmd.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	fileReader := bytes.NewReader(fileContent)

	// 7. Upload to storage
	publicURL, err := s.storage.UploadFile(ctx, mediaType.Bucket, filePath, fileReader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 8. Create domain entity - match table columns
	media, err := domain.NewMedia(
		cmd.FileHeader.Filename, // FileName
		mediaType.Bucket,        // Bucket
		filePath,                // Path
		publicURL,               // URL
		mediaType.ID,            // MediaTypeID
		cmd.EntityID,            // EntityID
		cmd.UploadedBy,          // UploadedBy
		contentType,             // MimeType
		cmd.FileHeader.Size,     // FileSize
	)
	if err != nil {
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, err
	}

	// 9. Save to database
	if err := s.repo.CreateMedia(ctx, media); err != nil {
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	return media, nil
}

// ============================================================
// DELETE
// ============================================================

func (s *mediaService) DeleteFile(ctx context.Context, id string) error {
	media, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return err
	}
	if media == nil {
		return domain.ErrMediaNotFound
	}

	_ = s.storage.DeleteFile(ctx, media.Bucket, media.Path)
	return s.repo.DeleteMedia(ctx, id)
}

func (s *mediaService) DeleteFilesByEntity(ctx context.Context, entityID string) error {
	mediaList, _, err := s.repo.GetMediaByEntity(ctx, entityID, 0, 0)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		_ = s.DeleteFile(ctx, media.ID)
	}

	return nil
}

func (s *mediaService) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
	mediaList, err := s.repo.GetMediaByEntityAndType(ctx, entityID, mediaTypeID)
	if err != nil {
		return err
	}

	for _, media := range mediaList {
		_ = s.DeleteFile(ctx, media.ID)
	}

	return nil
}

// ============================================================
// GET
// ============================================================

func (s *mediaService) GetMediaByID(ctx context.Context, id string) (*domain.Media, error) {
	return s.repo.GetMediaByID(ctx, id)
}

func (s *mediaService) GetMediaByEntity(ctx context.Context, entityID string, page, pageSize int) ([]*domain.Media, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.repo.GetMediaByEntity(ctx, entityID, limit, offset)
}

func (s *mediaService) GetMediaByEntityAndType(ctx context.Context, entityID, mediaTypeID string) ([]*domain.Media, error) {
	return s.repo.GetMediaByEntityAndType(ctx, entityID, mediaTypeID)
}

func (s *mediaService) GetImagesByEntity(ctx context.Context, entityID string) ([]*domain.Media, error) {
	// Get image media type ID
	imageType, err := s.repo.GetMediaTypeByName(ctx, "event")
	if err != nil {
		return nil, err
	}
	if imageType == nil {
		return nil, domain.ErrMediaTypeNotFound
	}

	return s.repo.GetMediaByEntityAndType(ctx, entityID, imageType.ID)
}

func (s *mediaService) GetPrimaryMediaForEntity(ctx context.Context, entityID string) (*domain.Media, error) {
	return s.repo.GetPrimaryMediaForEntity(ctx, entityID)
}

func (s *mediaService) GetMediaByUploader(ctx context.Context, userID string, page, pageSize int) ([]*domain.Media, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.repo.GetMediaByUploader(ctx, userID, limit, offset)
}

func (s *mediaService) GetMediaByType(ctx context.Context, mediaTypeName string, page, pageSize int) ([]*domain.Media, int64, error) {
	mediaType, err := s.repo.GetMediaTypeByName(ctx, mediaTypeName)
	if err != nil {
		return nil, 0, err
	}
	if mediaType == nil {
		return nil, 0, domain.ErrMediaTypeNotFound
	}

	limit := pageSize
	offset := (page - 1) * pageSize
	return s.repo.GetMediaByType(ctx, mediaType.ID, limit, offset)
}

// ============================================================
// MEDIA TYPES
// ============================================================

func (s *mediaService) GetMediaTypes(ctx context.Context) ([]*domain.MediaType, error) {
	return s.repo.GetAllMediaTypes(ctx)
}

func (s *mediaService) GetMediaTypeByName(ctx context.Context, name string) (*domain.MediaType, error) {
	return s.repo.GetMediaTypeByName(ctx, name)
}

// ============================================================
// HELPERS
// ============================================================

func (s *mediaService) generateFilePath(mediaTypeName string, entityID uuid.UUID, fileName string) string {
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

func (s *mediaService) isAllowedMimeType(mimeType string, mediaType *domain.MediaType) bool {
	allowedTypes := map[string][]string{
		"event":       {"image/jpeg", "image/png", "image/gif", "image/webp"},
		"business":    {"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"},
		"profile":     {"image/jpeg", "image/png", "image/gif", "image/webp"},
		"certificate": {"image/jpeg", "image/png", "image/webp", "application/pdf"},
		"recording":   {"video/mp4", "video/webm", "video/ogg"},
	}

	// Convert to lowercase to match map keys
	key := strings.ToLower(mediaType.Name)
	allowed, ok := allowedTypes[key]
	if !ok {
		return false
	}

	return slices.Contains(allowed, mimeType)
}