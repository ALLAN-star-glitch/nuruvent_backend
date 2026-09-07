// internal/modules/profile/service/media.go

package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// UploadUserAvatar uploads an avatar for a user
func (s *profileService) UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*domain.UserInfo, error) {
	if userID == "" {
		return nil, domain.ErrInvalidUserID
	}

	// 1. Validate file
	if err := s.validateImageFile(file, filename, contentType); err != nil {
		return nil, err
	}

	// 2. Check permission
	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}
	if viewerID != userID {
		viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
		if viewerPersonalDomain == "" {
			return nil, domain.ErrInvalidUserID
		}
		allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, viewerPersonalDomain)
		if err != nil {
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			return nil, domain.ErrPermissionDenied
		}
	}

	// 3. Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// 4. Delete old avatar if exists
	if user.AvatarURL != "" {
		log.Printf("🗑️ Deleting old avatar: %s", user.AvatarURL)
		mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_profile")
		if err != nil {
			log.Printf("⚠️ Failed to get media type: %v", err)
		} else if mediaType != nil {
			if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, userID, mediaType.ID); err != nil {
				log.Printf("⚠️ Failed to delete old avatar: %v", err)
			} else {
				log.Printf("✅ Old avatar deleted")
			}
		}
	}

	// 5. Generate clean filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	cleanFilename := userID + ext

	// 6. Upload to storage using media service
	uploadCmd := domain.UploadMediaCommand{
		File:          file,
		FileName:      cleanFilename,
		ContentType:   contentType,
		MediaTypeName: "media_type_profile",
		EntityID:      userID,
		UploadedBy:    viewerID,
	}

	mediaInfo, err := s.mediaSvc.UploadFile(ctx, uploadCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// 7. Update user with avatar URL
	user.AvatarURL = mediaInfo.URL
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("✅ User avatar uploaded for user: %s", userID)

	// 8. Return updated user info
	return &domain.UserInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Slug:        user.Slug,
		Email:       user.Email,
		Phone:       user.Phone,
		AccountType: user.AccountType,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		Location:    user.Location,
		Website:     user.Website,
		SocialLinks: user.SocialLinks,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// UploadAccountLogo uploads a logo for an account
func (s *profileService) UploadAccountLogo(ctx context.Context, accountID string, file []byte, filename, contentType string) (*domain.AccountInfo, error) {
	if accountID == "" {
		return nil, domain.ErrInvalidAccountID
	}

	// 1. Validate file
	if err := s.validateImageFile(file, filename, contentType); err != nil {
		return nil, err
	}

	// 2. Get viewer ID
	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}

	// 3. Check if viewer is an account admin in this account
	accountDomain := domain.InstitutionTeamDomain(accountID)
	if accountDomain == "" {
		return nil, domain.ErrInvalidAccountID
	}

	hasRole, err := s.permChecker.IsTeamAdmin(ctx, viewerID, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasRole {
		return nil, domain.ErrPermissionDenied
	}

	// 4. Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	// 5. Delete old logo if exists
	if account.LogoURL != "" {
		log.Printf("🗑️ Deleting old logo: %s", account.LogoURL)
		mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_business")
		if err != nil {
			log.Printf("⚠️ Failed to get media type: %v", err)
		} else if mediaType != nil {
			if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, accountID, mediaType.ID); err != nil {
				log.Printf("⚠️ Failed to delete old logo: %v", err)
			} else {
				log.Printf("✅ Old logo deleted")
			}
		}
	}

	// 6. Generate clean filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	cleanFilename := accountID + ext

	// 7. Upload to storage using media service
	uploadCmd := domain.UploadMediaCommand{
		File:          file,
		FileName:      cleanFilename,
		ContentType:   contentType,
		MediaTypeName: "media_type_business",
		EntityID:      accountID,
		UploadedBy:    viewerID,
	}

	mediaInfo, err := s.mediaSvc.UploadFile(ctx, uploadCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to upload logo: %w", err)
	}

	// 8. Update account with logo URL
	account.LogoURL = mediaInfo.URL
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("✅ Account logo uploaded for account: %s", accountID)

	return &domain.AccountInfo{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}, nil
}

// ============================================================
// MEDIA - DELETE METHODS
// ============================================================

// DeleteUserAvatar deletes a user's avatar
func (s *profileService) DeleteUserAvatar(ctx context.Context, userID string) error {
	if userID == "" {
		return domain.ErrInvalidUserID
	}

	// 1. Check permission
	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return domain.ErrPermissionDenied
	}
	if viewerID != userID {
		viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
		if viewerPersonalDomain == "" {
			return domain.ErrInvalidUserID
		}
		allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, viewerPersonalDomain)
		if err != nil {
			return fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			return domain.ErrPermissionDenied
		}
	}

	// 2. Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// 3. Delete from storage using media service
	if user.AvatarURL != "" {
		log.Printf("🗑️ Deleting avatar for user: %s", userID)

		mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_profile")
		if err != nil {
			log.Printf("⚠️ Failed to get media type: %v", err)
		} else if mediaType != nil {
			if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, userID, mediaType.ID); err != nil {
				log.Printf("⚠️ Failed to delete avatar files: %v", err)
			} else {
				log.Printf("✅ Avatar files deleted from storage")
			}
		}
	} else {
		log.Printf("ℹ️ No avatar URL to delete")
	}

	// 4. Clear avatar URL in database
	user.AvatarURL = ""
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("✅ User avatar deleted for user: %s", userID)
	return nil
}

// DeleteAccountLogo deletes an account's logo
func (s *profileService) DeleteAccountLogo(ctx context.Context, accountID string) error {
	if accountID == "" {
		return domain.ErrInvalidAccountID
	}

	// 1. Check permission - must be account admin
	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return domain.ErrPermissionDenied
	}

	accountDomain := domain.InstitutionTeamDomain(accountID)
	if accountDomain == "" {
		return domain.ErrInvalidAccountID
	}

	hasRole, err := s.permChecker.IsTeamAdmin(ctx, viewerID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !hasRole {
		return domain.ErrPermissionDenied
	}

	// 2. Get account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrAccountNotFound
	}

	// 3. Delete file from storage using media service
	if account.LogoURL != "" {
		log.Printf("🗑️ Deleting logo for account: %s", accountID)

		mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_business")
		if err != nil {
			log.Printf("⚠️ Failed to get media type: %v", err)
		} else if mediaType != nil {
			if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, accountID, mediaType.ID); err != nil {
				log.Printf("⚠️ Failed to delete logo files: %v", err)
			} else {
				log.Printf("✅ Logo files deleted from storage")
			}
		}
	}

	// 4. Clear logo URL in database
	account.LogoURL = ""
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("✅ Account logo deleted for account: %s", accountID)
	return nil
}