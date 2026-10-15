// internal/modules/certificate/service/get_by_code.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// GetByCode implements Service.GetByCode.
func (s *serviceImpl) GetByCode(
	ctx context.Context,
	code string,
) (*domain.Certificate, error) {
	vc, err := domain.NewVerificationCode(code)
	if err != nil {
		return nil, err
	}
	return s.deps.Certificates.FindByVerificationCode(ctx, vc)
}