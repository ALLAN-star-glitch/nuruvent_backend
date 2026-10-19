// internal/modules/certificate/service/get_by_code.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

func (s *serviceImpl) GetByCode(ctx context.Context, code string) (*domain.Certificate, error) {
	cert, err := s.deps.Certificates.FindByVerificationCode(ctx, domain.VerificationCode(code))
	if err != nil {
		return nil, err
	}

	// Fetch the account name to use as the Issuer on the public page.
	// This overrides any stale issuer_name stored in metadata.
	if s.deps.Accounts != nil && cert.AccountID != "" {
		accName, err := s.deps.Accounts.GetAccountName(ctx, cert.AccountID)
		if err == nil && accName != "" {
			if cert.Metadata == nil {
				cert.Metadata = make(domain.Metadata)
			}
			cert.Metadata[domain.MetadataKeyIssuerName] = accName
		}
	}

	return cert, nil
}