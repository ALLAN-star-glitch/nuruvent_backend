// internal/modules/auth/providers.go

package auth

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/jwt"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
)

// ProviderSet contains all auth module dependencies
var ProviderSet = wire.NewSet(
	// ============================================================
	// Infrastructure Layer (Outbound Ports)
	// ============================================================

	// Repository - implements authdomain.Repository
	postgres.NewPostgresRepository,

	// Authorization - implements authdomain.PermissionChecker, RoleManager, PolicyManager
	authorization.NewEnforcer,
	authorization.NewPermissionChecker,
	authorization.NewRoleManager,
	authorization.NewPolicyManager,

	// Token Service - implements authdomain.TokenService
	jwt.NewTokenService,

	// ============================================================
	// Application Layer (Inbound Ports)
	// ============================================================

	// Auth Service - implements service.Service interface
	// Depends on: authdomain.Repository, *config.Config, *redis.Client,
	// authdomain.QueueService, authdomain.PermissionChecker,
	// authdomain.RoleManager, authdomain.PolicyManager,
	// authdomain.TokenService, authdomain.NotificationService
	service.NewService,

	// ============================================================
	// Delivery Layer
	// ============================================================

	// Handler - HTTP handlers
	authhandler.NewAuthHandler,
)