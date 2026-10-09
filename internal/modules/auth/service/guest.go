// internal/modules/auth/service/guest.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// errGuestRace signals a concurrent request already created the same
// guest user. It is consumed inside CreateGuestUser and never
// surfaced to callers.
var errGuestRace = errors.New("guest user created concurrently")

// ============================================================
// GUEST PROVISIONING
// ============================================================

// FindOrCreateGuestByEmail returns the user with the given email, or
// creates a fully-provisioned guest:
//
//   - user row (is_guest = true, no password)
//   - personal account + account_member (owner/admin)
//   - personal team + team_member (via TeamService)
//
// Mirrors VerifyOTPAndCreateUser's two-phase pattern: user, account,
// and account_member are persisted inside one transaction; team
// creation runs after commit through s.teamSvc, and failure is logged
// but not fatal — the account already exists and a repair pass can
// create the team later.
//
// Idempotent: if the email already belongs to a user (guest or real),
// the existing user is returned unchanged and no new rows are created.
func (s *service) FindOrCreateGuestByEmail(
	ctx context.Context,
	email string,
	name string,
	phone string,
) (*authdomain.User, error) {
	user, accountID, err := s.CreateGuestUser(ctx, email, name, phone)
	if err != nil {
		return nil, err
	}

	// Personal team via the team module. Same posture as
	// VerifyOTPAndCreateUser → createAndAddToPersonalTeam.
	if accountID != "" {
		if err := s.createAndAddToPersonalTeam(ctx, user.ID, user.Name, accountID); err != nil {
			log.Printf("[FindOrCreateGuestByEmail] ⚠️ failed to create personal team for guest %s: %v",
				user.ID, err)
		}
	} else {
		log.Printf("[FindOrCreateGuestByEmail] ⚠️ guest %s has no account, skipping team", user.ID)
	}

	return user, nil
}

// CreateGuestUser creates a lightweight guest identity — user row,
// personal account, and account_member — but does NOT create a
// personal team.
//
// Use this from internal callers that only need the account-layer
// half (e.g. flows that provision teams themselves, or tests). Public
// guest registration should call FindOrCreateGuestByEmail instead so
// the user ends up fully provisioned.
//
// Returns the account ID that was provisioned (or, for idempotent
// hits and races, an empty string — the caller should not attempt
// team creation in those cases).
//
// Idempotent: if the email already belongs to any user, that user is
// returned unchanged and nothing new is written.
func (s *service) CreateGuestUser(
	ctx context.Context,
	email string,
	name string,
	phone string,
) (*authdomain.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, "", fmt.Errorf("email is required")
	}

	// 1. Existing user wins — no creation, no provisioning.
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", fmt.Errorf("lookup user by email: %w", err)
	}
	if existing != nil {
		return existing, "", nil
	}

	// 2. Resolve personal account type.
	accountType, err := s.repo.GetAccountTypeBySlug(ctx, types.AccountTypePersonalSlug)
	if err != nil || accountType == nil {
		return nil, "", fmt.Errorf("resolve personal account type: %w", err)
	}

	// 3. Allocate a unique username (reuses the same helper used by
	//    RegisterWithInvitation and VerifyOTPAndCreateUser).
	username, err := s.allocateUsername(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("allocate username: %w", err)
	}

	// 4. Build the guest entity.
	short := uuid.New().String()[:12]
	guest := &authdomain.User{
		ID:               uuid.New().String(),
		Slug:             "guest-" + strings.ToLower(short[:8]),
		Name:             guestFallbackName(name),
		DisplayName:      guestFallbackName(name),
		Email:            email,
		PasswordHash:     "", // no password; must authenticate via link or by setting one
		Phone:            phone,
		AccountTypeID:    accountType.ID,
		Username:         username,
		EmailVerified:    false,
		PhoneVerified:    false,
		IdentityVerified: false,
		IsActive:         true,
		IsGuest:          true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 5. Persist user + account + account_member in one transaction.
	//    Team creation happens outside (see FindOrCreateGuestByEmail),
	//    matching the existing signup flow.
	var accountID string
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateUser(txCtx, guest); err != nil {
			if isUniqueViolation(err) {
				return errGuestRace
			}
			return fmt.Errorf("create guest user: %w", err)
		}

		acctID, err := s.repo.ProvisionUserAccount(txCtx, guest)
		if err != nil {
			return err
		}
		accountID = acctID

		return nil
	})

	// 6. Race: another request created this email first. Re-fetch and
	//    return the winner. Do NOT run team provisioning here — the
	//    other request is responsible for the full workspace, and we
	//    don't have the account ID.
	if errors.Is(err, errGuestRace) {
		raced, err2 := s.repo.GetUserByEmail(ctx, email)
		if err2 != nil {
			return nil, "", fmt.Errorf("guest race: failed to re-fetch user: %w", err2)
		}
		if raced == nil {
			return nil, "", fmt.Errorf("guest race: user disappeared after unique violation")
		}
		return raced, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	return guest, accountID, nil
}

// ============================================================
// HELPERS
// ============================================================

// guestFallbackName returns the given name, or "Guest" if empty.
func guestFallbackName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Guest"
	}
	return name
}

// isUniqueViolation reports whether err is a Postgres unique
// constraint violation (SQLSTATE 23505).
//
// If you already have a shared helper for this in internal/shared,
// import that and delete this function.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value")
}
