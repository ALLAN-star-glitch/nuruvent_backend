// internal/app/adapters/certificate/storage_adapter.go

package certificate

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	sharedstorage "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
	"github.com/google/uuid"
)

// StorageAdapter adapts internal/shared/storage.Client to the
// certificate module's domain.ObjectStorage port.
type StorageAdapter struct {
	client *sharedstorage.Client
}

// NewStorageAdapter constructs a StorageAdapter.
func NewStorageAdapter(client *sharedstorage.Client) *StorageAdapter {
	return &StorageAdapter{client: client}
}

// Put implements domain.ObjectStorage.
func (a *StorageAdapter) Put(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	bucket, err := a.client.GetBucket("certificates")
	if err != nil {
		return "", fmt.Errorf("resolve certificates bucket: %w", err)
	}
	return a.client.UploadFile(ctx, bucket, key, bytes.NewReader(data), contentType)
}

// Delete implements domain.ObjectStorage.
func (a *StorageAdapter) Delete(ctx context.Context, key string) error {
	bucket, err := a.client.GetBucket("certificates")
	if err != nil {
		return fmt.Errorf("resolve certificates bucket: %w", err)
	}
	return a.client.DeleteFile(ctx, bucket, key)
}

// UploadPDF implements domain.ObjectStorage.
func (a *StorageAdapter) UploadPDF(ctx context.Context, path string, pdfBytes []byte) (string, error) {
	return a.Put(ctx, path, pdfBytes, "application/pdf")
}

// GetPublicURL implements domain.ObjectStorage.
func (a *StorageAdapter) GetPublicURL(path string) string {
	bucket, err := a.client.GetBucket("certificates")
	if err != nil {
		return ""
	}
	return a.client.GetPublicURL(bucket, path)
}

// UploadTemplateImage stores a template asset (logo, seal, signature,
// background) and returns its public URL.
func (a *StorageAdapter) UploadTemplateImage(
	ctx context.Context,
	templateID string,
	filename string,
	contentType string,
	fileData []byte,
) (string, error) {
	if templateID == "" {
		return "", fmt.Errorf("template ID is required")
	}
	if len(fileData) == 0 {
		return "", fmt.Errorf("file data is empty")
	}

	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/webp":      true,
		"image/gif":       true,
		"image/svg+xml":   true,
	}
	if !allowedTypes[contentType] {
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".png"
	}

	key := fmt.Sprintf(
		"templates/%s/%d-%s%s",
		templateID,
		time.Now().UnixNano(),
		uuid.New().String()[:8],
		ext,
	)

	return a.Put(ctx, key, fileData, contentType)
}

// DeleteFileByURL extracts the object key from a Supabase public URL
// and deletes it.
//
// Supabase URLs look like:
// https://<project>.supabase.co/storage/v1/object/public/certificates/templates/<id>/<file>
func (a *StorageAdapter) DeleteFileByURL(ctx context.Context, url string) error {
	if url == "" {
		return nil
	}

	const marker = "/public/"
	idx := strings.Index(url, marker)
	if idx == -1 {
		return fmt.Errorf("invalid storage URL: %s", url)
	}

	remainder := url[idx+len(marker):]

	bucketName := "certificates"
	prefix := bucketName + "/"
	if !strings.HasPrefix(remainder, prefix) {
		return fmt.Errorf("URL does not belong to certificates bucket: %s", url)
	}

	key := strings.TrimPrefix(remainder, prefix)
	if key == "" {
		return fmt.Errorf("empty key in URL: %s", url)
	}

	return a.Delete(ctx, key)
}

var _ domain.ObjectStorage = (*StorageAdapter)(nil)