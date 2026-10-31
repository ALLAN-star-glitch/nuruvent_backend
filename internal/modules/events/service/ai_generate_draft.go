// internal/modules/events/service/ai_generate_draft.go

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

type GenerateEventDraftResult struct {
	Draft    *GeneratedEventDraft `json:"draft"`
	Warnings []string             `json:"warnings"`
}

func (s *eventService) GenerateEventDraft(
	ctx context.Context,
	req GenerateEventDraftRequest,
) (*GenerateEventDraftResult, error) {

	if s.aiSvc == nil {
		return nil, ErrAIDisabled
	}

	if err := s.checkEventCreatePermission(
		ctx, req.CreatedBy, req.TeamID, req.TeamType, req.AccountID,
	); err != nil {
		return nil, err
	}

	start := time.Now()
	log.Printf("[AI] generate-draft start  event_type=%s prompt_len=%d",
		req.EventTypeID, len(req.Prompt))

	req = applyRequestDefaults(req)
	if err := validateGenerateEventDraftRequest(req); err != nil {
		return nil, err
	}

	pctx, err := s.loadPromptContext(ctx, req)
	if err != nil {
		return nil, err
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(req, pctx)
	cctx := buildCorrectionContext(req, pctx)
	log.Printf("[AI] prompt built  chars=%d version=%s",
		len(systemPrompt)+len(userPrompt), PromptVersion)

	draft, warnings, retryErrs, hardErr := s.attemptGenerate(
		ctx, systemPrompt, userPrompt, cctx,
	)
	if hardErr != nil {
		log.Printf("[AI] parse failed: %v", hardErr)
		return nil, hardErr
	}

	if len(retryErrs) == 0 {
		log.Printf("[AI] generate-draft ok  shape=%s schedules=%d warnings=%d latency_ms=%d",
			draft.Shape(), len(draft.Schedules), len(warnings),
			time.Since(start).Milliseconds())
		return &GenerateEventDraftResult{Draft: draft, Warnings: warnings}, nil
	}

	// Retry once with a fix prompt that INCLUDES the failed draft.
	log.Printf("[AI] validation failed  errors=%v retry=true", retryErrs)

	failedJSON, _ := json.MarshalIndent(draft, "", "  ")
	fixPrompt := buildFixPrompt(userPrompt, string(failedJSON), retryErrs)

	draft2, warnings2, retryErrs2, hardErr2 := s.attemptGenerate(
		ctx, systemPrompt, fixPrompt, cctx,
	)
	if hardErr2 == nil && len(retryErrs2) == 0 {
		warnings2 = append(
			[]string{"Retried once after validation failure."},
			warnings2...,
		)
		log.Printf("[AI] generate-draft ok (retry)  shape=%s schedules=%d warnings=%d latency_ms=%d",
			draft2.Shape(), len(draft2.Schedules), len(warnings2),
			time.Since(start).Milliseconds())
		return &GenerateEventDraftResult{Draft: draft2, Warnings: warnings2}, nil
	}

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
// ATTEMPT
// ============================================================

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
		MaxTokens:    4096, // raised from 3000 — series drafts are longer
		Temperature:  0.7,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %v", ErrAIProvider, err)
	}

	cleaned := ai.CleanJSONResponse(raw)

	var parsed GeneratedEventDraft
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		// Parse failures are retryable: return as a validation error
		// so the caller can retry with the fix prompt.
		return nil, nil, []string{fmt.Sprintf(
			"response was not valid JSON: %v", err)}, nil
	}

	corrections, corrErr := applyCorrections(&parsed, cctx)
	if corrErr != nil {
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

func (s *eventService) checkPublishReadiness(d *GeneratedEventDraft) []string {
	var errs []string

	trimmedName := strings.TrimSpace(d.Name)
	if trimmedName == "" {
		errs = append(errs, "name is required")
	} else if len(trimmedName) < 3 {
		errs = append(errs, "name must be at least 3 characters")
	}

	if len(strings.TrimSpace(d.Description)) < 100 {
		errs = append(errs, fmt.Sprintf(
			"description is %d characters — minimum is 100; expand with the agenda, target audience, and outcomes",
			len(strings.TrimSpace(d.Description))))
	}

	if len(d.Schedules) == 0 {
		errs = append(errs, "at least one schedule is required")
	}
	if len(d.Tickets) == 0 {
		errs = append(errs, "at least one ticket is required")
	}
	if d.Capacity < 1 {
		errs = append(errs, "capacity must be at least 1")
	}

	// Shape contract.
	shape := d.Shape()
	if shape == ShapeInvalid {
		errs = append(errs,
			"invalid shape: is_recurring=true requires exactly one schedule")
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

	// Series ordering.
	if shape == ShapeSeries {
		var prev time.Time
		for i, sched := range d.Schedules {
			t, err := time.Parse("2006-01-02", sched.StartDate)
			if err != nil {
				continue
			}
			if !prev.IsZero() && !t.After(prev) {
				errs = append(errs, fmt.Sprintf(
					"series schedules must be in strictly increasing date order (schedule %d out of order)",
					i+1))
			}
			prev = t
			if sched.SessionNumber != i+1 {
				errs = append(errs, fmt.Sprintf(
					"series schedule %d has session_number %d, expected %d",
					i+1, sched.SessionNumber, i+1))
			}
		}
	}

	// Recurrence.
	if d.IsRecurring {
		if d.Recurrence == nil {
			errs = append(errs, "is_recurring is true but recurrence block is missing")
		} else {
			switch d.Recurrence.Frequency {
			case "weekly":
				if len(d.Recurrence.DaysOfWeek) == 0 {
					errs = append(errs, "weekly recurrence requires days_of_week")
				}
				for _, day := range d.Recurrence.DaysOfWeek {
					if !isValidFullWeekday(day) {
						errs = append(errs, fmt.Sprintf(
							"invalid weekday %q — use full lowercase name", day))
					}
				}
			case "daily":
				// no additional fields required
			case "monthly":
				// monthly requires only interval and (end_date|count)
			default:
				errs = append(errs, fmt.Sprintf(
					"invalid recurrence frequency %q", d.Recurrence.Frequency))
			}

			if d.Recurrence.EndDate == nil && d.Recurrence.Count == nil {
				errs = append(errs,
					"recurrence requires end_date or count — derive it from the prompt")
			}
		}
	} else if d.Recurrence != nil {
		errs = append(errs, "recurrence must be null when is_recurring is false")
	}

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

	var minDur, maxDur int
	if pctx.EventType != nil {
		minDur = pctx.EventType.MinDuration
		maxDur = pctx.EventType.MaxDuration
	}

	return correctionContext{
		Request:         req,
		EventTypeID:     pctx.EventType.ID,
		CategoryID:      req.CategoryID,
		TicketTypeIDs:   allowed,
		Timezone:        pctx.Timezone,
		Language:        pctx.Language,
		Currency:        pctx.Currency,
		MinCapacity:     pctx.MinCapacity,
		MaxCapacity:     pctx.MaxCapacity,
		EventTypeMinDur: minDur,
		EventTypeMaxDur: maxDur,
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

type DraftUnpublishableError struct {
	Reason           string
	ValidationErrors []string
	RawDraft         *GeneratedEventDraft
}

func (e *DraftUnpublishableError) Error() string {
	return fmt.Sprintf("draft unpublishable: %s", e.Reason)
}

var _ = errors.Is