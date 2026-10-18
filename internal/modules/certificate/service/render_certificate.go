// internal/modules/certificate/service/render_certificate.go
package service

import "context"


func (s *serviceImpl) RenderCertificatePDF(ctx context.Context, certificateID string) error {
	cert, err := s.deps.Certificates.FindByID(ctx, certificateID)
	if err != nil {
		return err
	}

	// Already rendered (race between two workers) → nothing to do
	if cert.PDFURL != "" {
		return nil
	}

	_, err = s.renderAndStore(ctx, cert)
	return err
}