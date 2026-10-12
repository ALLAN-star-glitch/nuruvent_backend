package http

import (
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// SessionAccountResolver is the one method the middleware needs.
// attendance.Service satisfies it structurally.
type SessionAccountResolver interface {
	ResolveSessionAccount(ctx context.Context, sessionID string) (string, error)
}

// SessionContextMiddleware resolves the owning account for a
// /sessions/:id/* request and injects it into Locals as the Casbin
// domain, so AuthorizationMiddleware can run a domain-scoped check.
//
// MUST run after authMiddleware and before authzMiddleware.
//
// Failure modes:
//
//   - resolver nil       → DomainDeferred (authz steps aside; the
//                          handler's own existence check produces
//                          the correct 404/403)
//   - session not found  → DomainDeferred (same reason)
//   - resolver error     → 500
//   - success            → account:<uuid> in Locals
func SessionContextMiddleware(resolver SessionAccountResolver) fiber.Handler {
	return func(c fiber.Ctx) error {
		sessionID := c.Params("id")
		if sessionID == "" {
			c.Locals(authdomain.ContextKeyDomain, authdomain.DomainDeferred)
			return c.Next()
		}

		if resolver == nil {
			log.Printf("⚠️  SESSION CTX: resolver not wired for %s", c.Path())
			c.Locals(authdomain.ContextKeyDomain, authdomain.DomainDeferred)
			return c.Next()
		}

		accountID, err := resolver.ResolveSessionAccount(c.Context(), sessionID)
		if err != nil {
			if errors.Is(err, attendance.ErrSessionNotFound) {
				log.Printf("⏸  SESSION CTX: session %s not found, deferring", sessionID)
				c.Locals(authdomain.ContextKeyDomain, authdomain.DomainDeferred)
				return c.Next()
			}
			log.Printf("❌ SESSION CTX: resolve failed for %s: %v", sessionID, err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to resolve session owner")
		}

		domain := authdomain.AccountDomain(accountID)
		log.Printf("✅ SESSION CTX: session=%s domain=%s", sessionID, domain)
		c.Locals(authdomain.ContextKeyDomain, domain)
		return c.Next()
	}
}