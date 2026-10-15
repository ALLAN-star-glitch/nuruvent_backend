// internal/modules/certificate/service/issue_batch.go

package service

import (
	"context"
	"errors"


	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// IssueBatch is the public entry point invoked by the payment webhook
// adapter after a batch order is paid. It builds the ExternalCompletion
// for each recipient and delegates to IssueBulk.
//
// Idempotent per recipient — existing certificates are returned
// unchanged, so a webhook re-delivery cannot duplicate issuance.
func (s *serviceImpl) IssueBatch(
	ctx context.Context,
	req domain.BatchIssuanceRequest,
) (*domain.BatchIssuanceResult, error) {
	if req.EventID == "" {
		return nil, domain.ErrContextIDRequired
	}
	if req.AccountID == "" {
		return nil, domain.ErrAccountIDRequired
	}
	if len(req.Recipients) == 0 {
		return &domain.BatchIssuanceResult{}, nil
	}

	completions := make([]domain.ExternalCompletion, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		completions = append(completions, domain.ExternalCompletion{
			ContextType: "event",
			ContextID:   req.EventID,
			Recipient: domain.Recipient{
				UserID: r.UserID,
				Name:   r.Name,
				Email:  r.Email,
			},
			CompletedAt: r.CompletedAt,
			Metadata: domain.Metadata{
				domain.MetadataKeyContextTitle:    req.EventTitle,
				domain.MetadataKeyContextDate:     req.EventDate.UTC().Format("2006-01-02"),
				domain.MetadataKeyContextDuration: req.EventDuration,
				domain.MetadataKeyIssuerName:      req.IssuerName,
				domain.MetadataKeyRecipientName:   r.Name,
			},
			TemplateID:       req.TemplateID,
			AccountID:        req.AccountID,
			PaymentReference: req.PaymentReference,
		})
	}

	bulk, err := s.IssueBulk(ctx, completions)
	if err != nil {
		return nil, err
	}

	result := &domain.BatchIssuanceResult{
		Issued: bulk.Succeeded - bulk.Reused,
		Reused: bulk.Reused,
		Failed: bulk.Failed,
	}
	for _, item := range bulk.Items {
		if item.Err != nil {
			result.Errors = append(result.Errors, item.Err.Error())
		}
	}
	return result, nil
}

// issueBatch is an internal alias for IssueBatch used by the free
// path. Kept separate so the public method can validate independently.
func (s *serviceImpl) issueBatch(
	ctx context.Context,
	req domain.BatchIssuanceRequest,
) (*domain.BatchIssuanceResult, error) {
	return s.IssueBatch(ctx, req)
}

// compile-time assertion for the unused-errors import guard.
var _ = errors.Is