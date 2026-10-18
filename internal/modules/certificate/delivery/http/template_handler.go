// internal/modules/certificate/delivery/http/template_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// CreateTemplate handles POST /certificate-templates.
//
// Account-scope endpoint. IsSystem and SystemSlug are forced to
// false/"" here so callers cannot sneak a system template through the
// account route — that only happens via /admin/certificate-templates.
func (h *Handler) CreateTemplate(c fiber.Ctx) error {
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

	tmpl, err := h.svc.CreateTemplate(c.Context(), userID, service.CreateTemplateInput{
		AccountID:      body.AccountID,
		Name:           body.Name,
		Description:    body.Description,
		Design:         designFromDTO(body.Design),
		PriceMinor:     body.PriceMinor,
		TierTemplateID: body.TierTemplateID,
		IsSystem:       false, // account route — force
		SystemSlug:     "",    // account route — force
	})
	if err != nil {
		return respondClassifiedError(c, "create template", err)
	}

	return response.Created(c, "Template created", toTemplateResponse(tmpl))
}

// ListTemplates handles GET /certificate-templates?account_id=...
func (h *Handler) ListTemplates(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	accountID := handlerhelper.GetQueryString(c, "account_id", "")
	if accountID == "" {
		return response.BadRequest(c, "account_id query parameter is required", nil)
	}

	templates, err := h.svc.ListTemplates(c.Context(), userID, accountID)
	if err != nil {
		return respondClassifiedError(c, "list templates", err)
	}

	out := make([]TemplateResponse, 0, len(templates))
	for _, t := range templates {
		out = append(out, toTemplateResponse(t))
	}
	return response.Success(c, "Templates retrieved", fiber.Map{
		"count":     len(out),
		"templates": out,
	})
}

// UpdateTemplate handles PUT /certificate-templates/:id and
// PUT /admin/certificate-templates/:id.
//
// The same handler serves both routes; the service branches on
// tmpl.IsSystem and enforces the correct slug rules for each. The
// handler just passes TierTemplateID and SystemSlug through.
func (h *Handler) UpdateTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
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
		PriceMinor:     body.PriceMinor,
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

	tmpl, err := h.svc.UpdateTemplate(c.Context(), userID, id, in)
	if err != nil {
		return respondClassifiedError(c, "update template", err)
	}

	return response.Success(c, "Template updated", toTemplateResponse(tmpl))
}

// DeleteTemplate handles DELETE /certificate-templates/:id and
// DELETE /admin/certificate-templates/:id.
//
// The service rejects system templates with
// domain.ErrCannotDeleteSystemTemplate, so the same handler is safe
// for both scopes.
func (h *Handler) DeleteTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Template ID is required", nil)
	}

	if err := h.svc.DeleteTemplate(c.Context(), userID, id); err != nil {
		return respondClassifiedError(c, "delete template", err)
	}

	return response.Success(c, "Template deleted", nil)
}

// SetDefaultTemplate handles POST /certificate-templates/:id/default.
func (h *Handler) SetDefaultTemplate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Template ID is required", nil)
	}

	if err := h.svc.SetDefaultTemplate(c.Context(), userID, id); err != nil {
		return respondClassifiedError(c, "set default template", err)
	}

	return response.Success(c, "Default template set", nil)
}