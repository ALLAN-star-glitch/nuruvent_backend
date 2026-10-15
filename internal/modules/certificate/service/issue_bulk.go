// internal/modules/certificate/service/issue_bulk.go

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// IssueBulk implements Service.IssueBulk.
//
// Does not stop on the first failure — records per-item errors so
// callers can retry only what failed.
func (s *serviceImpl) IssueBulk(
	ctx context.Context,
	completions []domain.ExternalCompletion,
) (*BulkResult, error) {
	result := &BulkResult{Items: make([]BulkItemResult, 0, len(completions))}

	for _, c := range completions {
		item := BulkItemResult{Completion: c}

		// Detect existing certificate before issuing, to distinguish
		// "already had" from "newly created".
		if existing, err := s.deps.Certificates.FindByIdentity(ctx, c.Identity()); err == nil {
			item.Certificate = existing
			item.AlreadyHad = true
			result.Reused++
			result.Succeeded++
			result.Items = append(result.Items, item)
			continue
		}

		cert, err := s.Issue(ctx, c)
		if err != nil {
			item.Err = err
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		item.Certificate = cert
		result.Succeeded++
		result.Items = append(result.Items, item)
	}

	return result, nil
}

// uniqueVerificationCode generates codes until one is unused.
func (s *serviceImpl) uniqueVerificationCode(
	ctx context.Context,
) (domain.VerificationCode, error) {
	const maxAttempts = 5
	for range maxAttempts {
		code, err := s.deps.Codes.NewCode()
		if err != nil {
			return "", err
		}
		if _, err := s.deps.Certificates.FindByVerificationCode(ctx, code); err != nil {
			if errors.Is(err, domain.ErrCertificateNotFound) {
				return code, nil
			}
			return "", err
		}
	}
	return "", fmt.Errorf("%w after %d attempts", domain.ErrVerificationCodeCollision, maxAttempts)
}