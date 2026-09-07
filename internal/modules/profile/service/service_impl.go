// internal/modules/profile/service/service_impl.go

package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// profileService implements the Service interface
type profileService struct {
	repo        domain.Repository
	permChecker domain.PermissionChecker
	roleManager domain.RoleManager
	mediaSvc    domain.MediaService
}


// NewProfileService creates a new profile service instance
func NewProfileService(
	repo domain.Repository,
	permChecker domain.PermissionChecker,
	roleManager domain.RoleManager,
	mediaSvc domain.MediaService,
) Service {
	return &profileService{
		repo:        repo,
		permChecker: permChecker,
		roleManager: roleManager,
		mediaSvc:    mediaSvc,
	}
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// getUserIDFromContext extracts user ID from context
func (s *profileService) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// validateImageFile validates an image file
func (s *profileService) validateImageFile(file []byte, filename, contentType string) error {
	// Check file size (max 5MB)
	maxSize := 5 * 1024 * 1024 // 5MB
	if len(file) > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of 5MB")
	}

	// Check content type (allow common image types)
	allowedTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
		"image/bmp":     true,
	}
	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported file type: %s. Allowed types: JPEG, PNG, GIF, WEBP, SVG, BMP", contentType)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
		".bmp":  true,
	}
	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s. Allowed extensions: .jpg, .jpeg, .png, .gif, .webp, .svg, .bmp", ext)
	}

	return nil
}
