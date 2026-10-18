// internal/modules/certificate/delivery/http/admin_template_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ListSystemTemplates handles GET /admin/certificate-templates.
func (h *Handler) ListSystemTemplates(c fiber.Ctx) error {
	templates, err := h.svc.ListSystemTemplates(c.Context())
	if err != nil {
		return respondClassifiedError(c, "list system templates", err)
	}

	out := make([]TemplateResponse, 0, len(templates))
	for _, t := range templates {
		out = append(out, toTemplateResponse(t))
	}

	return response.Success(c, "System templates retrieved", fiber.Map{
		"templates": out,
	})
}

// CreateSystemTemplate handles POST /admin/certificate-templates.
//
// System templates MUST provide system_slug — one of "default",
// "professional", "premium". The price is derived from the slug by the
// service layer; any price_minor in the request is ignored.
func (h *Handler) CreateSystemTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body CreateTemplateRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	// Force IsSystem=true — this endpoint only creates system templates.
	template, err := h.svc.CreateTemplate(c.Context(), userID, service.CreateTemplateInput{
		AccountID:      "",
		Name:           body.Name,
		Description:    body.Description,
		Design:         designFromDTO(body.Design),
		PriceMinor:     nil, // ignored for system templates — derived from slug
		TierTemplateID: "",
		IsSystem:       true,
		SystemSlug:     domain.SystemTemplateSlug(body.SystemSlug),
	})
	if err != nil {
		return respondClassifiedError(c, "create system template", err)
	}

	return response.Created(c, "System template created", toTemplateResponse(template))
}

// UpdateSystemTemplate handles PATCH /admin/certificate-templates/:id.
//
// system_slug is optional. If provided, it must be one of the three
// known tiers; changing it changes the price (derived from the slug).
func (h *Handler) UpdateSystemTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	templateID := c.Params("id")
	if templateID == "" {
		return response.BadRequest(c, "Template ID is required", nil)
	}

	var body UpdateTemplateRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	in := service.UpdateTemplateInput{
		Name:           body.Name,
		Description:    body.Description,
		PriceMinor:     nil, // ignored for system templates
		TierTemplateID: body.TierTemplateID,
	}
	if body.Design != nil {
		d := designFromDTO(*body.Design)
		in.Design = &d
	}
	if body.SystemSlug != nil {
		slug := domain.SystemTemplateSlug(*body.SystemSlug)
		in.SystemSlug = &slug
	}

	template, err := h.svc.UpdateTemplate(c.Context(), userID, templateID, in)
	if err != nil {
		return respondClassifiedError(c, "update system template", err)
	}

	return response.Success(c, "System template updated", toTemplateResponse(template))
}

// DeleteSystemTemplate handles DELETE /admin/certificate-templates/:id.
//
// Always fails — system templates are tier parents and cannot be
// deleted. The service returns ErrCannotDeleteSystemTemplate.
func (h *Handler) DeleteSystemTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	templateID := c.Params("id")
	if templateID == "" {
		return response.BadRequest(c, "Template ID is required", nil)
	}

	if err := h.svc.DeleteTemplate(c.Context(), userID, templateID); err != nil {
		return respondClassifiedError(c, "delete system template", err)
	}

	return response.Success(c, "System template deleted", nil)
}