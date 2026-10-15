// internal/app/adapters/certificate/storage_adapter.go

package certificate

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	sharedstorage "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
)

// StorageAdapter adapts internal/shared/storage.Client to the
// certificate module's domain.ObjectStorage port.
//
// The domain passes an opaque key; the adapter resolves the bucket
// ("certificates") and forwards the call to the shared client.
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

// compile-time assertion
var _ domain.ObjectStorage = (*StorageAdapter)(nil)