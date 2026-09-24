// internal/app/adapters/payment/user_email_resolver.go

package payment

import (
	"context"
	"fmt"

	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// UserEmailResolver resolves a user ID to an email address by calling
// the account module.
type UserEmailResolver struct {
	accountSvc accountService.Service
}

func NewUserEmailResolver(accountSvc accountService.Service) *UserEmailResolver {
	return &UserEmailResolver{accountSvc: accountSvc}
}

func (r *UserEmailResolver) ResolveEmail(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user id is empty")
	}

	// TODO: confirm the exact method on accountService that returns a
	// user's email. Adjust to the actual signature.
	user, err := r.accountSvc.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("fetch user %s: %w", userID, err)
	}
	if user == nil {
		return "", fmt.Errorf("user %s not found", userID)
	}
	return user.Email, nil
}

var _ paymentdomain.UserEmailResolver = (*UserEmailResolver)(nil)