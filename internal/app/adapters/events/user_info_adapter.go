// internal/app/adapters/events/user_info_adapter.go

package events

import (
	"context"
	"fmt"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountsvc "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	eventsdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// UserInfoAdapter implements eventsdomain.UserInfoProvider by delegating
// to the account module's Service.
type UserInfoAdapter struct {
	accounts accountsvc.Service
}

// NewUserInfoAdapter wires the account service to the events port.
func NewUserInfoAdapter(accounts accountsvc.Service) eventsdomain.UserInfoProvider {
	return &UserInfoAdapter{accounts: accounts}
}

func (a *UserInfoAdapter) GetUserByID(ctx context.Context, userID string) (*eventsdomain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	user, err := a.accounts.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}
	return mapToEventsUserInfo(user), nil
}

func (a *UserInfoAdapter) GetUserByIDWithDetails(ctx context.Context, userID string) (*eventsdomain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	user, err := a.accounts.GetUserByIDWithDetails(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}
	return mapToEventsUserInfo(user), nil
}

func (a *UserInfoAdapter) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*eventsdomain.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*eventsdomain.UserInfo{}, nil
	}

	users, err := a.accounts.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	out := make([]*eventsdomain.UserInfo, 0, len(users))
	for _, u := range users {
		if mapped := mapToEventsUserInfo(u); mapped != nil {
			out = append(out, mapped)
		}
	}
	return out, nil
}

// mapToEventsUserInfo converts an account-module UserInfo into an events-module UserInfo.
func mapToEventsUserInfo(u *accountdomain.UserInfo) *eventsdomain.UserInfo {
	if u == nil {
		return nil
	}
	return &eventsdomain.UserInfo{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
	}
}