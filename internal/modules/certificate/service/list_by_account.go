// internal/modules/certificate/service/list_by_account.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ListByAccount implements Service.ListByAccount.
func (s *serviceImpl) ListByAccount(
	ctx context.Context,
	userID, accountID string,
	f domain.CertificateFilter,
) ([]*domain.Certificate, error) {
	ok, err := s.deps.Authz.CanReadAll(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.deps.Certificates.ListByAccount(ctx, accountID, f)
}