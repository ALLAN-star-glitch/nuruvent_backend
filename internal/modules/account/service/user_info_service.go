// internal/modules/account/service/user_info_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// USER INFO PROJECTIONS
// ============================================================
//
// Read-only user lookups for cross-module consumers (events creator
// display, audit logs, etc.). These replace the profile module's old
// GetUserProfile* methods.
//
// No permission checks are performed here — callers authorize access
// at their own boundary before calling these methods.

// GetUserByID returns basic user info.
// Returns (nil, nil) if the user does not exist.
func (s *accountService) GetUserByID(ctx context.Context, userID string) (*accountdomain.UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	return toUserInfo(user), nil
}

// GetUserByIDWithDetails is a richer variant of GetUserByID.
// Currently identical to GetUserByID — reserved for future joins
// (avatar, bio, etc.) without changing the call site.
func (s *accountService) GetUserByIDWithDetails(ctx context.Context, userID string) (*accountdomain.UserInfo, error) {
	return s.GetUserByID(ctx, userID)
}

// GetUsersByIDs returns user info for multiple IDs in one call.
// Missing IDs are silently skipped (not returned as nil entries).
func (s *accountService) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*accountdomain.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*accountdomain.UserInfo{}, nil
	}

	users, err := s.repo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	out := make([]*accountdomain.UserInfo, 0, len(users))
	for _, u := range users {
		if u != nil {
			out = append(out, toUserInfo(u))
		}
	}
	return out, nil
}

// toUserInfo maps a domain user to the read-only projection.
func toUserInfo(u *accountdomain.User) *accountdomain.UserInfo {
	if u == nil {
		return nil
	}
	return &accountdomain.UserInfo{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
	}
}

// ============================================================
// USER AVATAR
// ============================================================

// UploadUserAvatar stores a new avatar for the user and updates avatar_url.
// Self-service: the caller may only change their own avatar.
func (s *accountService) UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*accountdomain.UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if len(file) == 0 {
		return nil, fmt.Errorf("avatar file is required")
	}

	// Self-service guard: caller may only modify their own avatar.
	callerID := accountdomain.GetUserID(ctx)
	if callerID != "" && callerID != userID {
		return nil, accountdomain.ErrForbidden
	}

	// 1. Upload via media port (media_type_profile)
	media, err := s.mediaSvc.UploadFile(ctx, accountdomain.UploadMediaCommand{
		File:          file,
		FileName:      filename,
		ContentType:   contentType,
		MediaTypeName: types.MediaTypeProfileName,
		EntityID:      userID,
		UploadedBy:    callerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// 2. Persist new URL on the user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	user.AvatarURL = media.URL
	user.UpdatedAt = time.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return toUserInfo(user), nil
}

// DeleteUserAvatar removes the user's avatar from storage and clears avatar_url.
// Self-service: only the caller may delete their own avatar.
func (s *accountService) DeleteUserAvatar(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	callerID := accountdomain.GetUserID(ctx)
	if callerID != "" && callerID != userID {
		return accountdomain.ErrForbidden
	}

	// 1. Resolve media type
	mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, types.MediaTypeProfileName)
	if err != nil {
		return fmt.Errorf("failed to resolve media type: %w", err)
	}
	if mediaType == nil {
		return fmt.Errorf("media type %q not found", types.MediaTypeProfileName)
	}

	// 2. Delete files
	if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, userID, mediaType.ID); err != nil {
		return fmt.Errorf("failed to delete avatar files: %w", err)
	}

	// 3. Clear avatar_url
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	user.AvatarURL = ""
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(ctx, user)
}


// ============================================================
// USER PROFILE OPERATIONS
// ============================================================

const (
	maxBioLength      = 500
	maxLocationLength = 255
	maxSocialLinks    = 10
	maxDisplayNameLen = 150
)

// GetMyProfile returns the authenticated user's own profile.
func (s *accountService) GetMyProfile(ctx context.Context, userID string) (*accountdomain.UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return toUserInfo(user), nil
}

// UpdateMyProfile updates the authenticated user's own profile.
//
// Partial-update semantics: fields with nil pointers are left untouched.
// Empty strings/empty maps clear the field.
func (s *accountService) UpdateMyProfile(ctx context.Context, userID string, updates ProfileUpdates) (*accountdomain.UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Load the current user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Apply updates field by field (nil = skip)
	if err := s.applyProfileUpdates(ctx, user, updates); err != nil {
		return nil, err
	}

	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	log.Printf("✅ Profile updated for user %s", user.ID)
	return toUserInfo(user), nil
}

// applyProfileUpdates mutates the user entity with non-nil fields from updates.
func (s *accountService) applyProfileUpdates(ctx context.Context, user *accountdomain.User, updates ProfileUpdates) error {
	// Display name → also regenerates display_name + slug
	if updates.DisplayName != nil {
		raw := strings.TrimSpace(*updates.DisplayName)
		if raw == "" {
			return fmt.Errorf("display name cannot be empty")
		}

		sanitized := s.sanitizer.DisplayName(raw)
		if sanitized == "" {
			return fmt.Errorf("display name is invalid after sanitization")
		}
		if len(sanitized) > maxDisplayNameLen {
			return fmt.Errorf("display name must be at most %d characters", maxDisplayNameLen)
		}

		// Regenerate slug from the new display name, ensuring uniqueness
		baseSlug := s.sanitizer.GenerateSlugFromName(sanitized)
		if baseSlug == "" {
			baseSlug = "user"
		}
		uniqueSlug := s.sanitizer.GenerateUniqueSlug(
			baseSlug,
			user.ID,
			func(slug string, excludeID string) bool {
				existing, err := s.repo.GetUserBySlug(ctx, slug)
				if err != nil {
					log.Printf("⚠️ Slug check error: %v", err)
					return false
				}
				return existing != nil && existing.ID != excludeID
			},
		)

		user.DisplayName = sanitized
		user.Slug = uniqueSlug
	}

	if updates.Phone != nil {
		user.Phone = strings.TrimSpace(*updates.Phone)
	}

	if updates.Bio != nil {
		bio := strings.TrimSpace(*updates.Bio)
		if len(bio) > maxBioLength {
			return fmt.Errorf("bio must be at most %d characters", maxBioLength)
		}
		user.Bio = bio
	}

	if updates.Location != nil {
		loc := strings.TrimSpace(*updates.Location)
		if len(loc) > maxLocationLength {
			return fmt.Errorf("location must be at most %d characters", maxLocationLength)
		}
		user.Location = loc
	}

	if updates.Website != nil {
		site := strings.TrimSpace(*updates.Website)
		if site != "" {
			if _, err := url.ParseRequestURI(site); err != nil {
				return fmt.Errorf("website must be a valid URL")
			}
		}
		user.Website = site
	}

	if updates.SocialLinks != nil {
		links := *updates.SocialLinks
		if len(links) > maxSocialLinks {
			return fmt.Errorf("at most %d social links are allowed", maxSocialLinks)
		}
		// Normalize keys + trim values
		normalized := make(map[string]string, len(links))
		for k, v := range links {
			key := strings.ToLower(strings.TrimSpace(k))
			val := strings.TrimSpace(v)
			if key == "" || val == "" {
				continue
			}
			normalized[key] = val
		}
		user.SocialLinks = normalized
	}

	return nil
}

// GetPublicProfile returns a user's public-facing profile.
func (s *accountService) GetPublicProfile(ctx context.Context, userID string) (*accountdomain.PublicProfile, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, nil
	}
	if !user.IsActive {
		return nil, nil
	}

	return toPublicProfile(user), nil
}

// GetPublicProfileBySlug returns a user's public profile by slug.
func (s *accountService) GetPublicProfileBySlug(ctx context.Context, slug string) (*accountdomain.PublicProfile, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	user, err := s.repo.GetUserBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}
	if user == nil {
		return nil, nil
	}
	if !user.IsActive {
		return nil, nil
	}

	return toPublicProfile(user), nil
}

// toPublicProfile maps a user entity to the public projection.
func toPublicProfile(u *accountdomain.User) *accountdomain.PublicProfile {
	if u == nil {
		return nil
	}
	return &accountdomain.PublicProfile{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		Location:    u.Location,
		Website:     u.Website,
		SocialLinks: u.SocialLinks,
	}
}

