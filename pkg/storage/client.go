// pkg/storage/client.go

package storage

import (
	"context"
	"fmt"
	"io"

	storage_go "github.com/supabase-community/storage-go"
)

// Client wraps the Supabase storage client
type Client struct {
	client     *storage_go.Client
	buckets    BucketConfig
}

// BucketConfig holds all bucket names
type BucketConfig struct {
	Events      string
	Businesses  string
	Profiles    string
	Certificates string
	Recordings  string
}

// NewClient creates a new Supabase storage client
func NewClient(supabaseURL, serviceKey string, bucketConfig BucketConfig) *Client {
	client := storage_go.NewClient(
		supabaseURL+"/storage/v1",
		serviceKey,
		nil,
	)
	
	return &Client{
		client:  client,
		buckets: bucketConfig,
	}
}

// GetClient returns the underlying storage client
func (c *Client) GetClient() *storage_go.Client {
	return c.client
}

// GetBucket returns the bucket name for a given type
func (c *Client) GetBucket(bucketType string) (string, error) {
	switch bucketType {
	case "events":
		return c.buckets.Events, nil
	case "businesses":
		return c.buckets.Businesses, nil
	case "profiles":
		return c.buckets.Profiles, nil
	case "certificates":
		return c.buckets.Certificates, nil
	case "recordings":
		return c.buckets.Recordings, nil
	default:
		return "", fmt.Errorf("unknown bucket type: %s", bucketType)
	}
}

// ================================================
// STORAGE OPERATIONS (Wrapped)
// ================================================

// UploadFile uploads a file to Supabase Storage
// Based on: UploadFile(bucketId string, relativePath string, data io.Reader, fileOptions ...FileOptions) (FileUploadResponse, error)
func (c *Client) UploadFile(ctx context.Context, bucket, path string, file io.Reader, contentType string) (string, error) {
	opts := storage_go.FileOptions{
		ContentType: &contentType,
	}
	
	_, err := c.client.UploadFile(bucket, path, file, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Get public URL
	publicURLResp := c.client.GetPublicUrl(bucket, path)
	
	return publicURLResp.SignedURL, nil
}

// DeleteFile deletes a file from Supabase Storage
// Based on: RemoveFile(bucketId string, paths []string) ([]FileUploadResponse, error)
func (c *Client) DeleteFile(ctx context.Context, bucket, path string) error {
	_, err := c.client.RemoveFile(bucket, []string{path})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteFiles deletes multiple files from Supabase Storage
func (c *Client) DeleteFiles(ctx context.Context, bucket string, paths []string) error {
	_, err := c.client.RemoveFile(bucket, paths)
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}
	return nil
}

// GetPublicURL returns the public URL for a file
// Based on: GetPublicUrl(bucketId string, filePath string, urlOptions ...UrlOptions) SignedUrlResponse
func (c *Client) GetPublicURL(bucket, path string) string {
	resp := c.client.GetPublicUrl(bucket, path)
	return resp.SignedURL
}

// GetSignedURL returns a signed URL for a private file (with expiration)
// Based on: CreateSignedUrl(bucketId string, filePath string, expiresIn int) (SignedUrlResponse, error)
func (c *Client) GetSignedURL(ctx context.Context, bucket, path string, expiresIn int) (string, error) {
	resp, err := c.client.CreateSignedUrl(bucket, path, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}
	return resp.SignedURL, nil
}

// ListFiles lists files in a bucket with optional prefix
// Based on: ListFiles(bucketId string, queryPath string, options FileSearchOptions) ([]FileObject, error)
func (c *Client) ListFiles(ctx context.Context, bucket, prefix string) ([]storage_go.FileObject, error) {
	opts := storage_go.FileSearchOptions{
		Limit: 100,
		Offset: 0,
		SortByOptions: storage_go.SortBy{
			Column: "name",
			Order:  "asc",
		},
	}
	resp, err := c.client.ListFiles(bucket, prefix, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return resp, nil
}

// FileExists checks if a file exists in a bucket
func (c *Client) FileExists(ctx context.Context, bucket, path string) (bool, error) {
	files, err := c.client.ListFiles(bucket, path, storage_go.FileSearchOptions{
		Limit: 1,
		Offset: 0,
		SortByOptions: storage_go.SortBy{
			Column: "name",
			Order:  "asc",
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return len(files) > 0, nil
}

// ================================================
// BUCKET OPERATIONS
// ================================================

// ListBuckets lists all buckets
// Based on: ListBuckets() ([]Bucket, error) - no context parameter
func (c *Client) ListBuckets(ctx context.Context) ([]storage_go.Bucket, error) {
	buckets, err := c.client.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	return buckets, nil
}