// internal/modules/certificate/service/send_certificate.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// SendCertificateEmail regenerates and re-sends the certificate email
// to the recipient.
//
// Authorized for:
//   - The issuing account's admins (certificate:update)
//   - The certificate's recipient (own-scoped)
//
// Idempotent at the notification layer — the email is queued again
// regardless, but the caller's actions are gated by the authorization
// check.
func (s *serviceImpl) SendCertificateEmail(
	ctx context.Context,
	userID, certificateID string,
) error {
	cert, err := s.deps.Certificates.FindByID(ctx, certificateID)
	if err != nil {
		return err
	}

	// Authorize: recipient self-access OR account admin.
	if err := s.authorizeSend(ctx, userID, cert); err != nil {
		return err
	}

	s.sendCertificateEmail(ctx, cert)
	return nil
}

// authorizeSend mirrors authorizeDownload but uses CanReadAll for the
// admin path since there is no dedicated "send" permission.
func (s *serviceImpl) authorizeSend(
	ctx context.Context,
	userID string,
	cert *domain.Certificate,
) error {
	// Recipient self-access.
	if cert.RecipientUserID != "" && cert.RecipientUserID == userID {
		return nil
	}

	// Account-admin access.
	ok, err := s.deps.Authz.CanReadAll(ctx, userID, cert.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}