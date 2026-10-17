// internal/modules/certificate/domain/object_storage.go

package domain

import "context"

// ObjectStorage stores rendered artefacts.
type ObjectStorage interface {
	// Put stores data under key and returns a publicly reachable URL.
	Put(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// Delete removes an object. Missing objects are not an error.
	Delete(ctx context.Context, key string) error


	UploadPDF(ctx context.Context, path string, pdfBytes []byte) (string, error)
	GetPublicURL(path string) string
	
	// ADD THESE: For handling template images (logo, seal, signature, background)
	UploadTemplateImage(ctx context.Context, templateID string, filename string, contentType string, fileData []byte) (string, error)
	DeleteFileByURL(ctx context.Context, url string) error
}