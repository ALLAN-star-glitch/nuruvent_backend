// internal/shared/storage/client.go

package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	storage_go "github.com/supabase-community/storage-go"
)

// Client wraps the Supabase storage client
type Client struct {
	client  *storage_go.Client
	buckets BucketConfig
}

// BucketConfig holds all bucket names
type BucketConfig struct {
	Events       string
	Businesses   string
	Profiles     string
	Certificates string
	Recordings   string
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

// NewClientFromConfig creates a new storage client from app config
// ✅ Takes *config.Config directly
func NewClientFromConfig(cfg *config.Config) *Client {
	bucketConfig := BucketConfig{
		Events:       cfg.Supabase.BucketEvent,
		Businesses:   cfg.Supabase.BucketBusiness,
		Profiles:     cfg.Supabase.BucketProfile,
		Certificates: cfg.Supabase.BucketCertificate,
		Recordings:   cfg.Supabase.BucketRecording,
	}

	return NewClient(
		cfg.Supabase.URL,
		cfg.Supabase.SecretKey,
		bucketConfig,
	)
}

// NewClientFromSupabaseConfig creates a new storage client from Supabase config
// ✅ Kept for backward compatibility or direct use
func NewClientFromSupabaseConfig(supabaseConfig *config.SupabaseConfig) *Client {
	bucketConfig := BucketConfig{
		Events:       supabaseConfig.BucketEvent,
		Businesses:   supabaseConfig.BucketBusiness,
		Profiles:     supabaseConfig.BucketProfile,
		Certificates: supabaseConfig.BucketCertificate,
		Recordings:   supabaseConfig.BucketRecording,
	}

	return NewClient(
		supabaseConfig.URL,
		supabaseConfig.SecretKey,
		bucketConfig,
	)
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

// UploadFile uploads a file to Supabase Storage
func (c *Client) UploadFile(ctx context.Context, bucket, path string, file io.Reader, contentType string) (string, error) {
	opts := storage_go.FileOptions{
		ContentType: &contentType,
	}

	_, err := c.client.UploadFile(bucket, path, file, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	publicURLResp := c.client.GetPublicUrl(bucket, path)
	return publicURLResp.SignedURL, nil
}

// DeleteFile deletes a file from Supabase Storage
func (c *Client) DeleteFile(ctx context.Context, bucket, path string) error {
	_, err := c.client.RemoveFile(bucket, []string{path})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetPublicURL returns the public URL for a file
func (c *Client) GetPublicURL(bucket, path string) string {
	resp := c.client.GetPublicUrl(bucket, path)
	return resp.SignedURL
}

// GetSignedURL returns a signed URL for a private file
func (c *Client) GetSignedURL(ctx context.Context, bucket, path string, expiresIn int) (string, error) {
	resp, err := c.client.CreateSignedUrl(bucket, path, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}
	return resp.SignedURL, nil
}

// ListFiles lists files in a bucket
func (c *Client) ListFiles(ctx context.Context, bucket, prefix string) ([]storage_go.FileObject, error) {
	opts := storage_go.FileSearchOptions{
		Limit:  100,
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