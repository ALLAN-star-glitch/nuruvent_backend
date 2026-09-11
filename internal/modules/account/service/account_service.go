// internal/modules/account/service/account_service.go

package service

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

// accountService implements Service.
//
// Methods are grouped by concern in separate files:
//   - account_type_service.go    — account type lookups
//   - account_crud_service.go    — account CRUD + team resolution
//   - account_member_service.go  — account member operations
//   - user_info_service.go       — user info projections
//
// This file only defines the struct and constructor.
type accountService struct {
	repo        accountdomain.Repository
	authSvc     AuthService
	notifSvc    NotificationService
	permChecker accountdomain.PermissionChecker
	mediaSvc    accountdomain.MediaService
	sanitizer   validation.Sanitize
}

func NewAccountService(
	repo accountdomain.Repository,
	authSvc AuthService,
	notifSvc NotificationService,
	permChecker accountdomain.PermissionChecker,
	mediaSvc accountdomain.MediaService,
) Service {
	return &accountService{
		repo:        repo,
		authSvc:     authSvc,
		notifSvc:    notifSvc,
		permChecker: permChecker,
		mediaSvc:    mediaSvc,
		sanitizer:   validation.Sanitize{},
	}
}