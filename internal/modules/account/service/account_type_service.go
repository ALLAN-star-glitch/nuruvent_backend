// internal/modules/account/service/account_type_service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

func (s *accountService) GetAccountTypes(ctx context.Context) ([]*accountdomain.AccountType, error) {
	return s.repo.GetAccountTypes(ctx)
}

func (s *accountService) GetAccountTypeByID(ctx context.Context, id string) (*accountdomain.AccountType, error) {
	if id == "" {
		return nil, accountdomain.ErrAccountTypeNotFound
	}
	return s.repo.GetAccountTypeByID(ctx, id)
}

func (s *accountService) GetAccountTypeBySlug(ctx context.Context, slug string) (*accountdomain.AccountType, error) {
	if slug == "" {
		return nil, accountdomain.ErrAccountTypeNotFound
	}
	return s.repo.GetAccountTypeBySlug(ctx, slug)
}