// internal/modules/payment/infrastructure/postgres/assertions.go

package postgres

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"

// Compile-time assertions that every concrete type in this package
// satisfies its corresponding domain port. If any method signature
// drifts from the interface, the build fails here with a clear message.

var (
	_ paymentdomain.OrderRepository        = (*OrderRepository)(nil)
	_ paymentdomain.PaymentRepository      = (*PaymentRepository)(nil)
	_ paymentdomain.RefundRepository       = (*RefundRepository)(nil)
	_ paymentdomain.WebhookEventRepository = (*WebhookEventRepository)(nil)
)