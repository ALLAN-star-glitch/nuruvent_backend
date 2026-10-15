// internal/modules/certificate/domain/object_storage.go

package domain

import "context"

// ObjectStorage stores rendered artefacts.
type ObjectStorage interface {
	// Put stores data under key and returns a publicly reachable URL.
	Put(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// Delete removes an object. Missing objects are not an error.
	Delete(ctx context.Context, key string) error
}