// internal/modules/registration/delivery/http/registration_handler.go

package http

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
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

	// Enriched context carries userID / teamID / teamType / accountID
	// into downstream services (events, tickets, etc.) so private-event
	// visibility rules apply correctly.
	ctx := handlerhelper.EnrichUserContext(c)

	reg, err := h.svc.RegisterForEvent(ctx, cmd)
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

	ctx := handlerhelper.EnrichUserContext(c)

	er, err := h.svc.GetByID(ctx, id, actorID, guestEmail)
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

	ctx := handlerhelper.EnrichUserContext(c)

	regs, total, err := h.svc.ListByEvent(ctx, eventID, actorID, filter)
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

	ctx := handlerhelper.EnrichUserContext(c)

	res, err := h.svc.ListMineRegistrations(ctx, service.ListMineRegistrationsCommand{
		UserID:    userID,
		EventID:   c.Query("event_id"),
		Search:    c.Query("search"),
		Statuses:  parseCSV(c.Query("status")),
		SortBy:    c.Query("sort_by", "created_at"),
		SortOrder: c.Query("sort_order", "desc"),
		Page:      parseQueryInt(c, "page", 1),
		PageSize:  parseQueryInt(c, "page_size", 20),
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	out := CrossEventRegistrationListResponse{
		Registrations: make([]CrossEventRegistrationResponse, 0, len(res.Registrations)),
		Total:         res.Total,
		Page:          res.Page,
		PageSize:      res.PageSize,
	}
	for _, r := range res.Registrations {
		out.Registrations = append(out.Registrations, toCrossEventRegistrationResponse(r))
	}

	return response.Success(c, "Registrations retrieved successfully", out)
}

// ListAllRegistrations handles GET /registrations.
//
// Cross-event organizer view. Returns registrations across every
// event owned by any account the caller belongs to, or — when
// team_id+scope=team are provided — across only the events in that
// team.
//
// Query params:
//   page        — 1-indexed (default 1)
//   page_size   — rows per page (default 20, max 100)
//   event_id    — optional; filter to one event
//   team_id     — optional; filter to one team (requires scope=team)
//   scope       — set to "team" to apply team_id
//   search      — optional; matches attendee name or email
//   status      — optional; comma-separated status slugs
//   sort_by     — created_at | attendee_name | event_name | status
//   sort_order  — asc | desc
func (h *Handler) ListAllRegistrations(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	res, err := h.svc.ListAllRegistrations(ctx, service.ListAllRegistrationsCommand{
		UserID:    userID,
		EventID:   c.Query("event_id"),
		Search:    c.Query("search"),
		Statuses:  parseCSV(c.Query("status")),
		SortBy:    c.Query("sort_by", "created_at"),
		SortOrder: c.Query("sort_order", "desc"),
		Page:      parseQueryInt(c, "page", 1),
		PageSize:  parseQueryInt(c, "page_size", 20),
		ActorID:   userID,
		TeamID:    c.Query("team_id"),
		Scope:     c.Query("scope"),
	})
	if err != nil {
		if errors.Is(err, registrationdomain.ErrTeamAccessDenied) {
			return response.Forbidden(c, "You don't have access to this team's registrations", nil)
		}
		return mapDomainError(c, err)
	}

	out := CrossEventRegistrationListResponse{
		Registrations: make([]CrossEventRegistrationResponse, 0, len(res.Registrations)),
		Total:         res.Total,
		Page:          res.Page,
		PageSize:      res.PageSize,
	}
	for _, r := range res.Registrations {
		out.Registrations = append(
			out.Registrations,
			toCrossEventRegistrationResponse(r),
		)
	}

	return response.Success(c, "Registrations retrieved successfully", out)
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

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.CancelRegistration(ctx, cmd); err != nil {
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

	ctx := handlerhelper.EnrichUserContext(c)

	entry, err := h.svc.JoinWaitlist(ctx, cmd)
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

	ctx := handlerhelper.EnrichUserContext(c)

	reg, err := h.svc.PromoteFromWaitlist(ctx, eventID)
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

// GetMySessionLinks handles GET /me/session-links.
//
// Returns one group per confirmed registration for the authenticated
// user, each carrying a fresh personalized join link per session.
func (h *Handler) GetMySessionLinks(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.GetMySessionLinks(ctx, userID)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Session links retrieved", toMySessionLinksResponse(result))
}