// internal/modules/media/service/service_impl.go

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

	"github.com/google/uuid"
)

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type mediaService struct {
	repo    domain.Repository
	storage *storage.Client
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
// UPLOAD - Updated to use pure domain types
// ============================================================

func (s *mediaService) UploadFile(ctx context.Context, cmd UploadCommand) (*domain.Media, error) {
	// 1. Validate input
	if len(cmd.File) == 0 {
		return nil, errors.New("file data is required")
	}
	if cmd.FileName == "" {
		return nil, errors.New("file name is required")
	}
	if cmd.MediaTypeName == "" {
		return nil, errors.New("media type is required")
	}
	if cmd.EntityID == "" {
		return nil, errors.New("entity ID is required")
	}
	if cmd.UploadedBy == "" {
		return nil, errors.New("uploaded by is required")
	}

	// 2. Get media type
	mediaType, err := s.repo.GetMediaTypeByName(ctx, cmd.MediaTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get media type: %w", err)
	}
	if mediaType == nil {
		return nil, domain.ErrMediaTypeNotFound
	}

	// 3. Validate file size
	if int64(len(cmd.File)) > mediaType.MaxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", len(cmd.File), mediaType.MaxFileSize)
	}

	// 4. Validate mime type
	if !s.isAllowedMimeType(cmd.ContentType, mediaType) {
		return nil, fmt.Errorf("file type %s is not allowed for %s", cmd.ContentType, cmd.MediaTypeName)
	}

	// 5. Parse entity ID
	entityID, err := uuid.Parse(cmd.EntityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	// 6. Generate file path
	filePath := s.generateFilePath(cmd.MediaTypeName, entityID, cmd.FileName)

	// 7. Create reader from bytes
	fileReader := bytes.NewReader(cmd.File)

	// 8. ✅ Convert domain types to infrastructure types for storage
	// Create a multipart.File from the reader
	file := &bytesReaderWrapper{Reader: fileReader}

	// Create a file header with metadata
	fileHeader := &multipart.FileHeader{
		Filename: cmd.FileName,
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len(cmd.File)),
	}
	fileHeader.Header.Set("Content-Type", cmd.ContentType)

	// 9. Upload to storage
	publicURL, err := s.storage.UploadFile(ctx, mediaType.Bucket, filePath, file, cmd.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 10. Create domain entity
	media, err := domain.NewMedia(
		cmd.FileName,    // FileName
		mediaType.Bucket, // Bucket
		filePath,        // Path
		publicURL,       // URL
		mediaType.ID,    // MediaTypeID
		cmd.EntityID,    // EntityID
		cmd.UploadedBy,  // UploadedBy
		cmd.ContentType, // MimeType
		int64(len(cmd.File)), // FileSize
	)
	if err != nil {
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, err
	}

	// 11. Save to database
	if err := s.repo.CreateMedia(ctx, media); err != nil {
		_ = s.storage.DeleteFile(ctx, mediaType.Bucket, filePath)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	return media, nil
}


// ============================================================
// DELETE - FIXED with proper error handling
// ============================================================

func (s *mediaService) DeleteFile(ctx context.Context, id string) error {
    // 1. Get media metadata
    media, err := s.repo.GetMediaByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get media record: %w", err)
    }
    if media == nil {
        return domain.ErrMediaNotFound
    }

    // 2. Delete from Supabase Storage first
    log.Printf("🗑️ Deleting file from storage: bucket=%s, path=%s", media.Bucket, media.Path)
    if err := s.storage.DeleteFile(ctx, media.Bucket, media.Path); err != nil {
        log.Printf("⚠️ Failed to delete file from storage: %v", err)
        // Continue to delete metadata even if storage deletion fails
        // This prevents orphaned metadata records
    } else {
        log.Printf("✅ File deleted from storage: %s", media.Path)
    }

    // 3. Delete metadata from PostgreSQL
    log.Printf("🗑️ Deleting media record from database: %s", id)
    if err := s.repo.DeleteMedia(ctx, id); err != nil {
        return fmt.Errorf("failed to delete media record: %w", err)
    }
    log.Printf("✅ Media record deleted: %s", id)

    return nil
}

func (s *mediaService) DeleteFilesByEntity(ctx context.Context, entityID string) error {
    log.Printf("🗑️ Deleting all media for entity: %s", entityID)
    
    // 1. Get all media for this entity
    mediaList, _, err := s.repo.GetMediaByEntity(ctx, entityID, 0, 0)
    if err != nil {
        return fmt.Errorf("failed to get media list: %w", err)
    }
    
    if len(mediaList) == 0 {
        log.Printf("ℹ️ No media found for entity: %s", entityID)
        return nil
    }
    
    log.Printf("📦 Found %d media items to delete", len(mediaList))
    
    var deleteErrors []string
    var successCount int
    
    // 2. Delete each file
    for _, media := range mediaList {
        log.Printf("🗑️ Deleting media: %s (path: %s)", media.ID, media.Path)
        
        if err := s.DeleteFile(ctx, media.ID); err != nil {
            errMsg := fmt.Sprintf("failed to delete media %s: %v", media.ID, err)
            log.Printf("⚠️ %s", errMsg)
            deleteErrors = append(deleteErrors, errMsg)
        } else {
            successCount++
        }
    }
    
    // 3. Report results
    if len(deleteErrors) > 0 {
        log.Printf("⚠️ Deleted %d/%d media files with errors", successCount, len(mediaList))
        return fmt.Errorf("errors during deletion: %s", strings.Join(deleteErrors, "; "))
    }
    
    log.Printf("✅ All %d media files deleted successfully for entity: %s", len(mediaList), entityID)
    return nil
}

func (s *mediaService) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
    log.Printf("🗑️ Deleting media for entity %s with media type: %s", entityID, mediaTypeID)
    
    // 1. Get media for this entity and media type
    mediaList, err := s.repo.GetMediaByEntityAndType(ctx, entityID, mediaTypeID)
    if err != nil {
        return fmt.Errorf("failed to get media list: %w", err)
    }
    
    if len(mediaList) == 0 {
        log.Printf("ℹ️ No media found for entity %s with type %s", entityID, mediaTypeID)
        return nil
    }
    
    log.Printf("📦 Found %d media items to delete", len(mediaList))
    
    var deleteErrors []string
    var successCount int
    
    // 2. Delete each file
    for _, media := range mediaList {
        log.Printf("🗑️ Deleting media: %s (path: %s)", media.ID, media.Path)
        
        if err := s.DeleteFile(ctx, media.ID); err != nil {
            errMsg := fmt.Sprintf("failed to delete media %s: %v", media.ID, err)
            log.Printf("⚠️ %s", errMsg)
            deleteErrors = append(deleteErrors, errMsg)
        } else {
            successCount++
        }
    }
    
    // 3. Report results
    if len(deleteErrors) > 0 {
        log.Printf("⚠️ Deleted %d/%d media files with errors", successCount, len(mediaList))
        return fmt.Errorf("errors during deletion: %s", strings.Join(deleteErrors, "; "))
    }
    
    log.Printf("✅ All %d media files deleted successfully for entity %s with type %s", len(mediaList), entityID, mediaTypeID)
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

func (s *mediaService) GetMediaTypeByID(ctx context.Context, id string) (*domain.MediaType, error) {
	if id == "" {
		return nil, errors.New("media type ID is required")
	}
	return s.repo.GetMediaTypeByID(ctx, id)
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
    // ✅ Use shared constants for keys
    allowedTypes := map[string][]string{
        types.MediaTypeEventName: {
            "image/jpeg",
            "image/png",
            "image/gif",
            "image/webp",
        },
        types.MediaTypeBusinessName: {
            "image/jpeg",
            "image/png",
            "image/gif",
            "image/webp",
            "image/svg+xml",
        },
        types.MediaTypeProfileName: {
            "image/jpeg",
            "image/png",
            "image/gif",
            "image/webp",
        },
        types.MediaTypeCertificateName: {
            "image/jpeg",
            "image/png",
            "image/webp",
            "application/pdf",
        },
        types.MediaTypeRecordingName: {
            "video/mp4",
            "video/webm",
            "video/ogg",
            "audio/mpeg",
            "audio/wav",
        },
    }

    // Use Name (internal) for lookup
    key := strings.ToLower(mediaType.Name)
    allowed, ok := allowedTypes[key] // allowed is storing a slice
    if !ok {
        log.Printf("⚠️ No allowed types defined for media type: %s", mediaType.Name)
        return false
    }

    return slices.Contains(allowed, mimeType)
}

// ============================================================
// BYTES READER WRAPPER - Implements multipart.File
// ============================================================

// bytesReaderWrapper wraps a bytes.Reader to implement multipart.File interface
type bytesReaderWrapper struct {
	*bytes.Reader
}

func (b *bytesReaderWrapper) Close() error {
	return nil
}

func (b *bytesReaderWrapper) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (b *bytesReaderWrapper) Stat() (fs.FileInfo, error) {
	return nil, nil
}