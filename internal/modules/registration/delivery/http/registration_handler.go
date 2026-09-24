// internal/modules/registration/delivery/http/registration_handler.go

package http

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// Handler holds the service and exposes HTTP methods.
type Handler struct {
	svc service.Service
}

// NewHandler constructs the handler.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// ============================================================
// REGISTER
// ============================================================

// RegisterForEvent handles POST /events/:id/register
func (h *Handler) RegisterForEvent(c fiber.Ctx) error {
	log.Println("[registration] ENTER RegisterForEvent") 
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	var body RegisterRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	// Basic shape validation
	if len(body.Selections) == 0 {
		return response.BadRequest(c, "At least one ticket selection is required", nil)
	}

	userID := handlerhelper.GetUserIDOptional(c)

	// Identity check: must be either an authenticated user OR a guest.
	if userID == "" && body.Guest == nil {
		return response.Unauthorized(c, "Authentication or guest details required", nil)
	}

	cmd := service.RegisterCommand{
		EventID:    eventID,
		UserID:     userID,
		Selections: toSelectionInputs(body.Selections),
	}
	if body.Guest != nil {
		cmd.Guest = &service.GuestIdentity{
			Name:  body.Guest.Name,
			Email: body.Guest.Email,
			Phone: body.Guest.Phone,
		}
	}

	reg, err := h.svc.RegisterForEvent(c.Context(), cmd)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Registration created successfully", toRegistrationResponse(reg))
}

// ============================================================
// GET / LIST
// ============================================================

// GetByID handles GET /registrations/:id
func (h *Handler) GetByID(c fiber.Ctx) error {
    id := c.Params("id")
    if id == "" {
        return response.BadRequest(c, "Registration ID is required", nil)
    }

    // The frontend can supply an email for guest registrations.
    // The service verifies it matches the stored guest_email.
    guestEmail := c.Query("email")

    actorID := handlerhelper.GetUserIDOptional(c)

    er, err := h.svc.GetByID(c.Context(), id, actorID, guestEmail)
    if err != nil {
        return mapDomainError(c, err)
    }

    return response.Success(c, "Registration retrieved successfully", toRegistrationResponse(er))
}

// ListByEvent handles GET /events/:id/registrations
func (h *Handler) ListByEvent(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)

	filter := service.ListFilterInput{
		Statuses: parseCSV(c.Query("status")),
		Page:     parseQueryInt(c, "page", 1),
		PageSize: parseQueryInt(c, "page_size", 20),
	}

	regs, total, err := h.svc.ListByEvent(c.Context(), eventID, actorID, filter)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Registrations retrieved successfully", toListResponse(regs, total, filter))
}

// ListMine handles GET /me/registrations
func (h *Handler) ListMine(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	filter := service.ListFilterInput{
		Statuses: parseCSV(c.Query("status")),
		Page:     parseQueryInt(c, "page", 1),
		PageSize: parseQueryInt(c, "page_size", 20),
	}

	regs, total, err := h.svc.ListByUser(c.Context(), userID, filter)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Registrations retrieved successfully", toListResponse(regs, total, filter))
}

// ============================================================
// CANCEL
// ============================================================

// Cancel handles DELETE /registrations/:id
func (h *Handler) Cancel(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Registration ID is required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)
	if actorID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body CancelRequest
	_ = c.Bind().Body(&body) // body is optional

	cmd := service.CancelCommand{
		RegistrationID: id,
		ActorID:        actorID,
		Reason:         body.Reason,
	}

	if err := h.svc.CancelRegistration(c.Context(), cmd); err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Registration cancelled successfully", nil)
}

// ============================================================
// WAITLIST
// ============================================================

// JoinWaitlist handles POST /events/:id/waitlist
func (h *Handler) JoinWaitlist(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	var body JoinWaitlistRequest
	_ = c.Bind().Body(&body)

	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" && body.Guest == nil {
		return response.Unauthorized(c, "Authentication or guest details required", nil)
	}

	cmd := service.JoinWaitlistCommand{
		EventID:      eventID,
		UserID:       userID,
		TicketTypeID: body.TicketTypeID,
	}
	if body.Guest != nil {
		cmd.Guest = &service.GuestIdentity{
			Name:  body.Guest.Name,
			Email: body.Guest.Email,
			Phone: body.Guest.Phone,
		}
	}

	entry, err := h.svc.JoinWaitlist(c.Context(), cmd)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Joined waitlist successfully", toWaitlistResponse(entry))
}

// PromoteFromWaitlist handles POST /events/:id/waitlist/promote
// Typically called internally; exposed for organizer use.
func (h *Handler) PromoteFromWaitlist(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	reg, err := h.svc.PromoteFromWaitlist(c.Context(), eventID)
	if err != nil {
		return mapDomainError(c, err)
	}

	if reg == nil {
		return response.Success(c, "No waitlist entries to promote", nil)
	}

	return response.Success(c, "Waitlist entry promoted", toRegistrationResponse(reg))
}

// ============================================================
// HELPERS (private to this file)
// ============================================================

func toSelectionInputs(in []TicketSelectionRequest) []service.TicketSelectionInput {
	out := make([]service.TicketSelectionInput, 0, len(in))
	for _, s := range in {
		out = append(out, service.TicketSelectionInput{
			TicketTypeID: s.TicketTypeID,
			Quantity:     s.Quantity,
		})
	}
	return out
}

func parseQueryInt(c fiber.Ctx, key string, fallback int) int {
	v := c.Query(key)
	if v == "" {
		return fallback
	}
	n := fallback
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

// parseCSV splits a comma-separated query value into a slice.
// Returns nil for empty input.
func parseCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}