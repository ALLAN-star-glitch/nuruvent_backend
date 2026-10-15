// internal/modules/certificate/delivery/http/template_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// Create handles POST /certificate-templates.
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
	AccountID:   body.AccountID,
	Name:        body.Name,
	Description: body.Description,
	Design:      designFromDTO(body.Design),
	PriceMinor:  body.PriceMinor, // ← NEW
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

// Update handles PUT /certificate-templates/:id.
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
	Name:        body.Name,
	Description: body.Description,
	PriceMinor:  body.PriceMinor, // ← NEW
}
	if body.Design != nil {
		d := designFromDTO(*body.Design)
		in.Design = &d
	}

	tmpl, err := h.svc.UpdateTemplate(c.Context(), userID, id, in)
	if err != nil {
		return respondClassifiedError(c, "update template", err)
	}

	return response.Success(c, "Template updated", toTemplateResponse(tmpl))
}

// Delete handles DELETE /certificate-templates/:id.
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

// SetDefault handles POST /certificate-templates/:id/default.
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

// ============================================================
// MAPPERS
// ============================================================

func designFromDTO(d TemplateDesignDTO) domain.TemplateDesign {
	return domain.TemplateDesign{
		BackgroundColor: d.BackgroundColor,
		TextColor:       d.TextColor,
		AccentColor:     d.AccentColor,
		FontFamily:      d.FontFamily,
		LogoURL:         d.LogoURL,
		SealURL:         d.SealURL,
		SignatureURL:    d.SignatureURL,
		BackgroundURL:   d.BackgroundURL,
		SignatoryName:   d.SignatoryName,
		SignatoryTitle:  d.SignatoryTitle,
		TitleText:       d.TitleText,
		SubtitleText:    d.SubtitleText,
		BodyText:        d.BodyText,
		FooterText:      d.FooterText,
		QRPosition:      d.QRPosition,
	}
}

func designToDTO(d domain.TemplateDesign) TemplateDesignDTO {
	return TemplateDesignDTO{
		BackgroundColor: d.BackgroundColor,
		TextColor:       d.TextColor,
		AccentColor:     d.AccentColor,
		FontFamily:      d.FontFamily,
		LogoURL:         d.LogoURL,
		SealURL:         d.SealURL,
		SignatureURL:    d.SignatureURL,
		BackgroundURL:   d.BackgroundURL,
		SignatoryName:   d.SignatoryName,
		SignatoryTitle:  d.SignatoryTitle,
		TitleText:       d.TitleText,
		SubtitleText:    d.SubtitleText,
		BodyText:        d.BodyText,
		FooterText:      d.FooterText,
		QRPosition:      d.QRPosition,
	}
}