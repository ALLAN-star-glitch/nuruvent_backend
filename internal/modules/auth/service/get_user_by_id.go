package service

import (
	"context"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

func (s *service) GetUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
    return s.repo.GetUserByID(ctx, userID)
}

