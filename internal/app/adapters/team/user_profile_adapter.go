package team

import (
	"context"
	"fmt"

	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// UserProfileAdapter implements teamService.UserProfileProvider on top
// of the account module's user projection.
//
// The account service is resolved lazily through AccountServiceHolder
// to break the wire-time cycle:
//
//   auth → team → UserProfileProvider → account → auth
//
// By the time GetUserProfilesByIDs is called, app.go has populated the
// holder and the account service is available.
type UserProfileAdapter struct {
	holder *AccountServiceHolder
}

func NewUserProfileAdapter(
	holder *AccountServiceHolder,
) teamService.UserProfileProvider {
	return &UserProfileAdapter{holder: holder}
}

func (a *UserProfileAdapter) GetUserProfilesByIDs(
	ctx context.Context,
	userIDs []string,
) (map[string]*teamService.UserProfile, error) {
	out := make(map[string]*teamService.UserProfile, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}

	svc, err := a.holder.Get()
	if err != nil {
		return nil, fmt.Errorf("account service unavailable: %w", err)
	}

	users, err := svc.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u == nil {
			continue
		}
		out[u.ID] = &teamService.UserProfile{
			ID:          u.ID,
			Name:        u.Name,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			AvatarURL:   u.AvatarURL,
		}
	}
	return out, nil
}