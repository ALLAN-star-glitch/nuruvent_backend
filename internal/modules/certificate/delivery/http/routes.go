// internal/modules/certificate/delivery/http/routes.go

package http

import (
	"github.com/gofiber/fiber/v3"

	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
)

type Handlers struct {
	Handler *Handler
}

func NewHandlers(svc certificateService.Service) *Handlers {
	return &Handlers{Handler: NewHandler(svc)}
}

func RegisterRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	// ---- Public (no auth) ----
	r.Get("/certificates/verify/:code", h.Handler.Verify)
	r.Get("/certificates/download/:token", h.Handler.DownloadByToken)

	// ---- Certificates (auth + authz) ----
	r.Post("/certificates/bulk", authMiddleware, authzMiddleware, h.Handler.IssueBulk)
	r.Post("/certificates/:id/revoke", authMiddleware, authzMiddleware, h.Handler.RevokeByUser)
	r.Get("/certificates/:id/download", authMiddleware, authzMiddleware, h.Handler.Download)
	r.Post("/certificates/:id/send", authMiddleware, authzMiddleware, h.Handler.SendCertificate)

	// ---- Templates (auth + authz) ----
	r.Post("/certificate-templates", authMiddleware, authzMiddleware, h.Handler.CreateTemplate)
	r.Get("/certificate-templates", authMiddleware, authzMiddleware, h.Handler.ListTemplates)

	// Media routes (register BEFORE /:id to avoid conflicts)
	r.Post("/certificate-templates/:id/upload", authMiddleware, authzMiddleware, h.Handler.UploadTemplateImage)
	r.Delete("/certificate-templates/media", authMiddleware, authzMiddleware, h.Handler.DeleteTemplateImage)

	// Template CRUD by ID
	r.Put("/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.UpdateTemplate)
	r.Delete("/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.DeleteTemplate)
	r.Post("/certificate-templates/:id/default", authMiddleware, authzMiddleware, h.Handler.SetDefaultTemplate)
}

// RegisterUserRoutes mounts self-service certificate routes.
//
// These routes are scoped to the authenticated user, not to an account.
// They use authMiddleware only — authzMiddleware requires an :accountId
// path parameter to derive a Casbin domain, and /me/* has none.
//
// ListMine filters by recipient_user_id inside the service, so a caller
// can only ever see their own certificates.
func RegisterUserRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	r.Get("/me/certificates", authMiddleware, h.Handler.ListMine)
}

func RegisterAccountRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	r.Get("/accounts/:id/certificates", authMiddleware, authzMiddleware, h.Handler.ListByAccount)
}

// RegisterAdminRoutes mounts platform-admin certificate-template routes.
// Casbin's domain resolver sees /api/v1/admin/* → platform domain,
// so only super_admin (and admin if granted) can pass.
func RegisterAdminRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	r.Get("/admin/certificate-templates", authMiddleware, authzMiddleware, h.Handler.ListSystemTemplates)
	r.Post("/admin/certificate-templates", authMiddleware, authzMiddleware, h.Handler.CreateSystemTemplate)
	r.Put("/admin/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.UpdateTemplate)
	r.Delete("/admin/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.DeleteTemplate)

	r.Post("/admin/certificate-templates/:id/upload", authMiddleware, authzMiddleware, h.Handler.UploadTemplateImage)
	r.Delete("/admin/certificate-templates/media", authMiddleware, authzMiddleware, h.Handler.DeleteTemplateImage)
}