package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"
)

// ============================================================
// GENERATE EVENT DRAFT
// ============================================================

// GenerateEventDraftResult is the payload returned to the caller.
type GenerateEventDraftResult struct {
	Draft    *GeneratedEventDraft `json:"draft"`
	Warnings []string             `json:"warnings"`
}

// GenerateEventDraft orchestrates the AI generation pipeline:
//
//  1. Authorization — same two-tier check as CreateDraft.
//  2. Shape validation + defaults.
//  3. Resolve DB context (event type, category, ticket types).
//  4. Build prompts.
//  5. Call AI (attempt 1) → parse → correct → check publish-readiness.
//  6. On validation failure, retry once with a fix prompt.
//  7. Return draft + warnings, or a structured error.
func (s *eventService) GenerateEventDraft(
	ctx context.Context,
	req GenerateEventDraftRequest,
) (*GenerateEventDraftResult, error) {

	if s.aiSvc == nil {
		return nil, ErrAIDisabled
	}

	// 0. Authorization — same two-tier check as CreateDraft.
	if err := s.checkEventCreatePermission(
		ctx, req.CreatedBy, req.TeamID, req.TeamType, req.AccountID,
	); err != nil {
		return nil, err
	}

	start := time.Now()
	log.Printf("[AI] generate-draft start  event_type=%s prompt_len=%d",
		req.EventTypeID, len(req.Prompt))

	// 1. Shape validation + defaults.
	req = applyRequestDefaults(req)
	if err := validateGenerateEventDraftRequest(req); err != nil {
		return nil, err
	}

	// 2. Resolve DB context.
	pctx, err := s.loadPromptContext(ctx, req)
	if err != nil {
		return nil, err
	}

	// 3. Build prompts.
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(req, pctx)
	cctx := buildCorrectionContext(req, pctx)
	log.Printf("[AI] prompt built  chars=%d version=%s",
		len(systemPrompt)+len(userPrompt), PromptVersion)

	// 4. Attempt 1 — a hard parse failure here is fatal (502).
	draft, warnings, retryErrs, hardErr := s.attemptGenerate(
		ctx, systemPrompt, userPrompt, cctx,
	)
	if hardErr != nil {
		log.Printf("[AI] parse failed: %v", hardErr)
		return nil, hardErr // already wrapped as ErrAIProvider / ErrAIParseFailure
	}

	// 5. If the first attempt is publishable, return it.
	if len(retryErrs) == 0 {
		log.Printf("[AI] generate-draft ok  warnings=%d latency_ms=%d",
			len(warnings), time.Since(start).Milliseconds())
		return &GenerateEventDraftResult{Draft: draft, Warnings: warnings}, nil
	}

	// 6. Retry once with a fix prompt.
	log.Printf("[AI] validation failed  errors=%v retry=true", retryErrs)

	fixPrompt := buildFixPrompt(userPrompt, retryErrs)
	draft2, warnings2, retryErrs2, hardErr2 := s.attemptGenerate(
		ctx, systemPrompt, fixPrompt, cctx,
	)
	if hardErr2 == nil && len(retryErrs2) == 0 {
		warnings2 = append(
			[]string{"Retried once after validation failure."},
			warnings2...,
		)
		log.Printf("[AI] generate-draft ok (retry)  warnings=%d latency_ms=%d",
			len(warnings2), time.Since(start).Milliseconds())
		return &GenerateEventDraftResult{Draft: draft2, Warnings: warnings2}, nil
	}

	// 7. Both attempts failed — return a structured 422.
	finalDraft := draft
	finalErrs := retryErrs
	if draft2 != nil {
		finalDraft = draft2
	}
	if len(retryErrs2) > 0 {
		finalErrs = retryErrs2
	}
	return nil, &DraftUnpublishableError{
		Reason:           "both attempts failed validation",
		ValidationErrors: finalErrs,
		RawDraft:         finalDraft,
	}
}

// ============================================================
// ATTEMPT — one full AI + parse + correct + validate cycle
// ============================================================

// attemptGenerate runs one full pipeline cycle:
//
//	AI call → parse JSON → apply corrections → check publish-readiness
//
// Returns:
//   - draft: the parsed and corrected draft (may be non-nil even on retry error)
//   - warnings: human-readable corrections applied
//   - validationErrs: fields that would trigger a retry
//   - hardErr: a genuine parse/provider failure that should NOT be retried
//
// hardErr and validationErrs are mutually exclusive. If hardErr is nil,
// validationErrs tells you whether to retry.
func (s *eventService) attemptGenerate(
	ctx context.Context,
	systemPrompt, userPrompt string,
	cctx correctionContext,
) (
	draft *GeneratedEventDraft,
	warnings []string,
	validationErrs []string,
	hardErr error,
) {
	raw, err := s.aiSvc.GenerateEventDraft(ctx, GenerateEventDraftAIRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    3000,
		Temperature:  0.7,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %v", ErrAIProvider, err)
	}

	cleaned := ai.CleanJSONResponse(raw)

	var parsed GeneratedEventDraft
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %v", ErrAIParseFailure, err)
	}

	corrections, corrErr := applyCorrections(&parsed, cctx)
	if corrErr != nil {
		// Correction failure is a *validation* error, not a parse error.
		// The retry can fix it.
		return &parsed, corrections, []string{corrErr.Error()}, nil
	}

	readinessErrs := s.checkPublishReadiness(&parsed)
	return &parsed, corrections, readinessErrs, nil
}

// ============================================================
// PROMPT CONTEXT
// ============================================================

func (s *eventService) loadPromptContext(
	ctx context.Context,
	req GenerateEventDraftRequest,
) (*promptContext, error) {

	// Event type — direct lookup exists.
	et, err := s.repo.GetEventTypeByID(ctx, req.EventTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load event type: %w", err)
	}
	if et == nil {
		return nil, ErrEventTypeNotFound
	}

	pctx := &promptContext{
		EventType:   et,
		Language:    req.Language,
		Timezone:    req.Timezone,
		Currency:    req.Currency,
		MinCapacity: req.MinCapacity,
		MaxCapacity: req.MaxCapacity,
	}

	// Category — filter in memory.
	if req.CategoryID != nil && *req.CategoryID != "" {
		cats, err := s.repo.GetAllCategories(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load categories: %w", err)
		}
		for _, c := range cats {
			if c.ID == *req.CategoryID {
				pctx.Category = c
				break
			}
		}
		if pctx.Category == nil {
			return nil, ErrCategoryNotFound
		}
	}

	// Ticket types — filter in memory.
	allTicketTypes, err := s.repo.GetAllTicketTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket types: %w", err)
	}
	byID := make(map[string]*domain.TicketTypeRow, len(allTicketTypes))
	for _, tt := range allTicketTypes {
		byID[tt.ID] = tt
	}
	for _, id := range req.TicketTypeIDs {
		tt, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrTicketTypeNotFound, id)
		}
		pctx.TicketTypes = append(pctx.TicketTypes, tt)
	}

	return pctx, nil
}

// ============================================================
// PUBLISH READINESS
// ============================================================

// checkPublishReadiness returns a list of structural errors that must
// be resolved before the draft is considered publishable. These errors
// trigger the retry path — they are not fatal on the first attempt.
func (s *eventService) checkPublishReadiness(d *GeneratedEventDraft) []string {
	var errs []string

	// Name — the raw title from which display_name and slug are derived.
	trimmedName := strings.TrimSpace(d.Name)
	if trimmedName == "" {
		errs = append(errs, "name is required")
	} else if len(trimmedName) < 3 {
		errs = append(errs, "name must be at least 3 characters")
	}

	// Description.
	if len(strings.TrimSpace(d.Description)) < 100 {
		errs = append(errs, fmt.Sprintf(
			"description must be at least 100 characters (got %d)",
			len(strings.TrimSpace(d.Description))))
	}

	// Structure.
	if len(d.Schedules) == 0 {
		errs = append(errs, "at least one schedule is required")
	}
	if len(d.Tickets) == 0 {
		errs = append(errs, "at least one ticket is required")
	}
	if d.Capacity < 1 {
		errs = append(errs, "capacity must be at least 1")
	}

	// Past-date guard.
	now := time.Now().UTC()
	for i, sched := range d.Schedules {
		if start, err := time.Parse("2006-01-02", sched.StartDate); err == nil {
			if start.Before(now) {
				errs = append(errs, fmt.Sprintf(
					"schedule %d start_date is in the past", i+1))
			}
		}
	}

	// Venue consistency — soft issues become retry triggers.
	errs = append(errs, validateVenueConsistency(d)...)

	return errs
}

// ============================================================
// CORRECTION CONTEXT
// ============================================================

func buildCorrectionContext(
	req GenerateEventDraftRequest,
	pctx *promptContext,
) correctionContext {
	allowed := make(map[string]struct{}, len(req.TicketTypeIDs))
	for _, id := range req.TicketTypeIDs {
		allowed[id] = struct{}{}
	}

	return correctionContext{
		Request:       req,
		EventTypeID:   pctx.EventType.ID,
		CategoryID:    req.CategoryID,
		TicketTypeIDs: allowed,
		Timezone:      pctx.Timezone,
		Language:      pctx.Language,
		MinCapacity:   pctx.MinCapacity,
		MaxCapacity:   pctx.MaxCapacity,
	}
}

// ============================================================
// REQUEST VALIDATION
// ============================================================

func applyRequestDefaults(req GenerateEventDraftRequest) GenerateEventDraftRequest {
	if req.Language == "" {
		req.Language = "en"
	}
	if req.Timezone == "" {
		req.Timezone = "Africa/Nairobi"
	}
	if req.Currency == "" {
		req.Currency = "KES"
	}
	if req.MinCapacity == 0 {
		req.MinCapacity = 10
	}
	if req.MaxCapacity == 0 {
		req.MaxCapacity = 5000
	}
	if len(req.TicketTypeIDs) > 10 {
		req.TicketTypeIDs = req.TicketTypeIDs[:10]
	}
	return req
}

func validateGenerateEventDraftRequest(req GenerateEventDraftRequest) error {
	prompt := strings.TrimSpace(req.Prompt)
	if len(prompt) < 10 || len(prompt) > 500 {
		return fmt.Errorf("prompt must be between 10 and 500 characters")
	}
	if req.EventTypeID == "" {
		return fmt.Errorf("event_type_id is required")
	}
	if len(req.TicketTypeIDs) < 1 || len(req.TicketTypeIDs) > 10 {
		return fmt.Errorf("ticket_type_ids must contain 1 to 10 IDs")
	}
	if req.MinCapacity > req.MaxCapacity {
		return fmt.Errorf("min_capacity cannot exceed max_capacity")
	}
	return nil
}

// ============================================================
// SENTINEL ERRORS
// ============================================================

var (
	ErrAIDisabled         = fmt.Errorf("AI service is not enabled")
	ErrAIProvider         = fmt.Errorf("AI provider error")
	ErrAIParseFailure     = fmt.Errorf("AI response could not be parsed")
	ErrEventTypeNotFound  = domain.ErrEventTypeNotFound
	ErrCategoryNotFound   = fmt.Errorf("category not found")
	ErrTicketTypeNotFound = fmt.Errorf("ticket type not found")
)

// DraftUnpublishableError carries the details of an AI output that
// could not be made publishable after the retry attempt.
type DraftUnpublishableError struct {
	Reason           string
	ValidationErrors []string
	RawDraft         *GeneratedEventDraft
}

func (e *DraftUnpublishableError) Error() string {
	return fmt.Sprintf("draft unpublishable: %s", e.Reason)
}

// ensure "errors" package is used (referenced in caller helpers if any).
var _ = errors.Is