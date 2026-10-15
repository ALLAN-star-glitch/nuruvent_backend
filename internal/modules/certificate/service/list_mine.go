// internal/modules/certificate/service/list_mine.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ListMine implements Service.ListMine.
func (s *serviceImpl) ListMine(
	ctx context.Context,
	userID string,
	f domain.CertificateFilter,
) ([]*domain.Certificate, error) {
	ok, err := s.deps.Authz.CanReadOwn(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.deps.Certificates.ListByRecipient(ctx, userID, f)
}