package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func (s *service) GetPayment(ctx context.Context, paymentID string) (*paymentdomain.Payment, error) {
	return s.deps.Payments.FindByID(ctx, paymentID)
}