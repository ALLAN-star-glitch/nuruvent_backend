//go:build wireinject
// +build wireinject

package auth

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization" // <-- Add this
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/jwt"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
)

var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewPostgresRepository,

	// Authorization (Casbin Enforcer & Permission Service)
	authorization.NewEnforcer, // Make sure you have a constructor for Enforcer if needed, or pass database setup
	authorization.NewService,  // Returns domain.PermissionService

	// Token Service
	jwt.NewTokenService,

	// OTP Service
	redis.NewOTPService,

	// Auth Service
	service.NewService,

	// Handler
	authhandler.NewAuthHandler,
)