package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

// normalizeAndCheckPhone validates + normalizes a phone number to E.164
// and rejects numbers already held by another active user.
// Returns "" when raw is empty (phone remains optional).
func (s *service) normalizeAndCheckPhone(ctx context.Context, raw string) (string, error) {
	normalized, err := s.validator.Phone.NormalizeAndValidate(raw)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrInvalidPhone):
			return "", authdomain.ErrInvalidPhone
		case errors.Is(err, validation.ErrPhoneRegionNotSupported):
			return "", authdomain.ErrPhoneRegionNotSupported
		default:
			return "", err
		}
	}
	if normalized == "" {
		return "", nil
	}

	exists, err := s.repo.UserExistsByPhone(ctx, normalized)
	if err != nil {
		return "", fmt.Errorf("check phone uniqueness: %w", err)
	}
	if exists {
		return "", authdomain.ErrPhoneAlreadyRegistered
	}
	return normalized, nil
}

// NormalizeAndCheckPhone is the public form used by the Service interface
// and by cross-module callers (e.g. the Account module's profile update).
func (s *service) NormalizeAndCheckPhone(ctx context.Context, raw string) (string, error) {
	return s.normalizeAndCheckPhone(ctx, raw)
}