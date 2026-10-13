
// internal/app/adapters/video/user_lookup.go

package video

import (
	"context"
	"strings"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	videoDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// UserLookupAdapter bridges the auth module's user repository to the
// video module's narrow UserLookup port.
type UserLookupAdapter struct {
	repo authDomain.Repository
}

// NewUserLookupAdapter constructs the adapter.
func NewUserLookupAdapter(repo authDomain.Repository) videoDomain.UserLookup {
	return &UserLookupAdapter{repo: repo}
}

// GetDisplayName returns the user's display name, falling back to
// their name, then to the local part of their email. Returns "" for
// missing users without erroring — the caller treats that as "no
// name available" and continues.
func (a *UserLookupAdapter) GetDisplayName(
	ctx context.Context,
	userID string,
) (string, error) {
	if userID == "" {
		return "", nil
	}

	user, err := a.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}

	if name := strings.TrimSpace(user.DisplayName); name != "" {
		return name, nil
	}
	if name := strings.TrimSpace(user.Name); name != "" {
		return name, nil
	}
	if email := strings.TrimSpace(user.Email); email != "" {
		if at := strings.Index(email, "@"); at > 0 {
			return email[:at], nil
		}
		return email, nil
	}
	return "", nil
}

// Compile-time assertion.
var _ videoDomain.UserLookup = (*UserLookupAdapter)(nil)