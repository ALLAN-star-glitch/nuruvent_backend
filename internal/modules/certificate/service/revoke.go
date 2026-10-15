// internal/modules/certificate/service/revoke.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Revoke implements Service.Revoke.
//
// Authorization is the caller's responsibility. Use RevokeByUser to
// have the service check.
func (s *serviceImpl) Revoke(
	ctx context.Context,
	certificateID, reason string,
) error {
	cert, err := s.deps.Certificates.FindByID(ctx, certificateID)
	if err != nil {
		return err
	}
	now := s.deps.Clock.Now().UTC()
	if err := cert.Revoke(reason, now); err != nil {
		return err
	}
	cert.UpdatedAt = now
	return s.deps.Certificates.Update(ctx, cert)
}

// RevokeByUser implements Service.RevokeByUser.
func (s *serviceImpl) RevokeByUser(
	ctx context.Context,
	userID, certificateID, reason string,
) error {
	cert, err := s.deps.Certificates.FindByID(ctx, certificateID)
	if err != nil {
		return err
	}

	ok, err := s.deps.Authz.CanRevoke(ctx, userID, cert.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}

	now := s.deps.Clock.Now().UTC()
	if err := cert.Revoke(reason, now); err != nil {
		return err
	}
	cert.UpdatedAt = now
	return s.deps.Certificates.Update(ctx, cert)
}