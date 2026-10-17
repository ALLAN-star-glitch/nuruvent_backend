package http

import (
	"github.com/gofiber/fiber/v3"

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
		PriceMinor:     body.PriceMinor,
		TierTemplateID: "",
		IsSystem:       true,
	})
	if err != nil {
		return respondClassifiedError(c, "create system template", err)
	}

	return response.Created(c, "System template created", toTemplateResponse(template))
}