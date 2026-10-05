// internal/modules/payment/delivery/http/payment_read_handler.go

package http

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// STATUS / METHOD LABELS
// ============================================================

// paymentStatusLabel maps a payment status slug to a human-readable
// label. Kept here (not in the domain) because it is a presentation
// concern.
func paymentStatusLabel(status string) string {
	switch strings.ToLower(status) {
	case "succeeded":
		return "Completed"
	case "pending":
		return "Pending"
	case "failed":
		return "Failed"
	case "refunded":
		return "Refunded"
	case "expired":
		return "Expired"
	default:
		return status
	}
}

// paymentMethodLabel maps a payment method slug to a human-readable
// label.
func paymentMethodLabel(method string) string {
	switch strings.ToLower(method) {
	case "mpesa":
		return "M-Pesa"
	case "card":
		return "Card"
	default:
		return method
	}
}

// parseDate accepts either an ISO 8601 date ("2026-10-05") or an
// RFC3339 timestamp. Returns nil if the string is empty or invalid,
// so a malformed query param degrades to "no filter" rather than a
// 500.
func parseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

// ============================================================
// GET /payments
// ============================================================

// ListPayments returns a paginated payment ledger for the caller's
// active account.
//
// The account scope comes from the JWT (ContextKeyAccountID), set by
// the auth middleware. It is required — the request is rejected if
// missing.
//
// @Summary List payments
// @Description Returns a paginated ledger of payments for the caller's active account.
// @Tags Payments
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Rows per page (default 20, max 100)"
// @Param search query string false "Search attendee, event, or transaction ID"
// @Param status query string false "Filter by status (succeeded, pending, failed, expired, refunded)"
// @Param method query string false "Filter by method (mpesa, card)"
// @Param event_id query string false "Filter by event"
// @Param sort_by query string false "Sort field (created_at, amount, completed_at)"
// @Param sort_order query string false "Sort direction (asc, desc)"
// @Param date_from query string false "Filter by created_at >= date (RFC3339 or YYYY-MM-DD)"
// @Param date_to query string false "Filter by created_at <= date (RFC3339 or YYYY-MM-DD)"
// @Success 200 {object} response.BaseResponse{data=ListPaymentsResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/payments [get]
func (h *Handler) ListPayments(c fiber.Ctx) error {
	accountID, _ := c.Locals(authdomain.ContextKeyAccountID).(string)
	if accountID == "" {
		return response.Unauthorized(c, "Active account required", nil)
	}

	var q ListPaymentsQuery
	if err := c.Bind().Query(&q); err != nil {
		return response.BadRequest(c, "Invalid query parameters", fiber.Map{
			"error": err.Error(),
		})
	}

	cmd := service.ListPaymentsCommand{
		BilledAccountID: accountID,
		Page:            q.Page,
		PageSize:        q.PageSize,
		Search:          q.Search,
		Status:          q.Status,
		Method:          q.Method,
		EventID:         q.EventID,
		SortBy:          q.SortBy,
		SortOrder:       q.SortOrder,
		DateFrom:        parseDate(q.DateFrom),
		DateTo:          parseDate(q.DateTo),
	}

	result, err := h.svc.ListPayments(c.Context(), cmd)
	if err != nil {
		return response.InternalError(c, "Failed to list payments", fiber.Map{
			"error": err.Error(),
		})
	}

	items := make([]PaymentListItemResponse, 0, len(result.Payments))
	for _, row := range result.Payments {
		items = append(items, toPaymentListItemResponse(row))
	}

	return response.Success(c, "Payments retrieved successfully", ListPaymentsResponse{
		Payments: items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// ============================================================
// GET /payments/stats
// ============================================================

// GetPaymentStats returns aggregate statistics for the caller's
// active account, honouring the same filters as ListPayments minus
// pagination.
//
// @Summary Get payment statistics
// @Description Returns aggregate revenue, fees, and status counts for the caller's active account.
// @Tags Payments
// @Produce json
// @Param event_id query string false "Filter by event"
// @Param date_from query string false "Filter by created_at >= date"
// @Param date_to query string false "Filter by created_at <= date"
// @Success 200 {object} response.BaseResponse{data=PaymentStatsResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/payments/stats [get]
func (h *Handler) GetPaymentStats(c fiber.Ctx) error {
	accountID, _ := c.Locals(authdomain.ContextKeyAccountID).(string)
	if accountID == "" {
		return response.Unauthorized(c, "Active account required", nil)
	}

	var q PaymentStatsQuery
	if err := c.Bind().Query(&q); err != nil {
		return response.BadRequest(c, "Invalid query parameters", fiber.Map{
			"error": err.Error(),
		})
	}

	cmd := service.PaymentStatsCommand{
		BilledAccountID: accountID,
		EventID:         q.EventID,
		DateFrom:        parseDate(q.DateFrom),
		DateTo:          parseDate(q.DateTo),
	}

	stats, err := h.svc.GetPaymentStats(c.Context(), cmd)
	if err != nil {
		return response.InternalError(c, "Failed to load payment stats", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Payment stats retrieved successfully", PaymentStatsResponse{
		TotalRevenue:        stats.TotalRevenue,
		TotalPlatformFees:   stats.TotalPlatformFees,
		TotalProcessingFees: stats.TotalProcessingFees,
		TotalNet:            stats.TotalNet,
		TransactionCount:    stats.TransactionCount,
		Currency:            stats.Currency,
		ByStatus:            stats.ByStatus,
	})
}

// ============================================================
// MAPPERS
// ============================================================

func toPaymentListItemResponse(
	row *paymentdomain.PaymentListRow,
) PaymentListItemResponse {
	resp := PaymentListItemResponse{
		ID:                 row.ID,
		OrderID:            row.OrderID,
		RegistrationID:     row.RegistrationID,
		RegistrationNumber: row.RegistrationNumber,

		AttendeeName:  row.AttendeeName,
		AttendeeEmail: row.AttendeeEmail,
		AttendeePhone: row.AttendeePhone,

		EventID:       row.EventID,
		EventTitle:    row.EventTitle,
		EventImageURL: row.EventImageURL,

		Amount:            row.Amount,
		Currency:          row.Currency,
		PlatformFee:       row.PlatformFee,
		ProcessingFee:     row.ProcessingFee,
		NetToOrganizer:    row.NetToOrganizer,
		PlatformFeeRate:   row.PlatformFeeRate,
		ProcessingFeeRate: row.ProcessingFeeRate,

		Status:        row.Status,
		StatusLabel:   paymentStatusLabel(row.Status),
		Provider:      row.Provider,
		Method:        row.Method,
		MethodLabel:   paymentMethodLabel(row.Method),
		TransactionID: row.ProviderReference,

		InitiatedAt: row.InitiatedAt.Format(time.RFC3339),
		CreatedAt:   row.CreatedAt.Format(time.RFC3339),
	}

	if row.EventStartDate != nil {
		resp.EventStartDate = row.EventStartDate.Format(time.RFC3339)
	}
	if row.CompletedAt != nil {
		resp.CompletedAt = row.CompletedAt.Format(time.RFC3339)
	}

	return resp
}