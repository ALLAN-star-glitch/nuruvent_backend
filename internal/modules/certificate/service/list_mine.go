// internal/modules/certificate/service/list_mine.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ListMine implements Service.ListMine.
//
// Self-service: returns certificates where recipient_user_id = caller.
// No Casbin check — this is scoped by identity, not by account role.
// The repository's ListByRecipient applies the recipient_user_id filter,
// so a caller can only ever see their own certificates.
func (s *serviceImpl) ListMine(
	ctx context.Context,
	userID string,
	f domain.CertificateFilter,
) ([]*domain.Certificate, error) {
	return s.deps.Certificates.ListByRecipient(ctx, userID, f)
}