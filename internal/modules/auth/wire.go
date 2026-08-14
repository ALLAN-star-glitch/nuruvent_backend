

package auth

import (
	"github.com/google/wire"


	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/jwt"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"


)

var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewPostgresRepository,

	// Authorization
	authorization.NewEnforcer,
	authorization.NewService,

	// Token Service
	jwt.NewTokenService,

	// OTP Service
	redis.NewOTPService,

	// Auth Service - needs notificationdomain.NotificationService
	service.NewService,

	// Handler
	authhandler.NewAuthHandler,
)