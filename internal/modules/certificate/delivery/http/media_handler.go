// internal/modules/certificate/delivery/http/media_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// UploadTemplateImage handles POST /certificate-templates/:id/upload
//
// Multipart form with a "file" field.
// Response: { success, message, data: { url } }
func (h *Handler) UploadTemplateImage(c fiber.Ctx) error {
	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	templateID := c.Params("id")
	if templateID == "" {
		return response.BadRequest(c, "Template ID is required", nil)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "File is required", fiber.Map{
			"error": err.Error(),
		})
	}

	// 5MB max size guard
	if file.Size > 5*1024*1024 {
		return response.BadRequest(c, "File too large — max 5MB", nil)
	}

	f, err := file.Open()
	if err != nil {
		return response.BadRequest(c, "Cannot open uploaded file", nil)
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		return response.BadRequest(c, "Cannot read uploaded file", nil)
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url, err := h.svc.UploadTemplateImage(
		c.Context(),
		templateID,
		file.Filename,
		contentType,
		data,
	)
	if err != nil {
		return respondClassifiedError(c, "upload template image", err)
	}

	return response.Success(c, "Image uploaded", fiber.Map{"url": url})
}

// DeleteTemplateImage handles DELETE /certificate-templates/media
//
// Body: { url: "https://..." }
func (h *Handler) DeleteTemplateImage(c fiber.Ctx) error {
	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}
	if body.URL == "" {
		return response.BadRequest(c, "url is required", nil)
	}

	if err := h.svc.DeleteTemplateImage(c.Context(), body.URL); err != nil {
		return respondClassifiedError(c, "delete template image", err)
	}

	return response.Success(c, "Image deleted", nil)
}