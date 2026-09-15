// internal/modules/registration/delivery/http/mappers.go

package http

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
)

func toRegistrationResponse(r *registrationdomain.EventRegistration) RegistrationResponse {
	resp := RegistrationResponse{
		ID:                 r.Registration.ID,
		RegistrationNumber: r.Registration.RegistrationNumber,
		Status:             toStatusResponse(r.Registration.Status),
		EventID:            r.EventID,
		Pricing: PricingResponse{
			Currency:      r.Pricing.Currency,
			Subtotal:      r.Pricing.Subtotal,
			DiscountTotal: r.Pricing.DiscountTotal,
			Total:         r.Pricing.Total,
		},
		CreatedAt:          r.Registration.CreatedAt,
		ConfirmedAt:        r.Registration.ConfirmedAt,
		CancelledAt:        r.Registration.CancelledAt,
		CancellationReason: r.Registration.CancellationReason,
	}

	if r.Registration.UserID != "" {
		resp.UserID = r.Registration.UserID
	} else {
		resp.Guest = &GuestResponse{
			Name:  r.Registration.GuestName,
			Email: r.Registration.GuestEmail,
			Phone: r.Registration.GuestPhone,
		}
	}

	resp.Selections = make([]TicketSelectionResponse, 0, len(r.Selections))
	for _, s := range r.Selections {
		resp.Selections = append(resp.Selections, TicketSelectionResponse{
			TicketTypeID: s.TicketTypeID,
			Quantity:     s.Quantity,
			UnitPrice:    s.UnitPrice,
			Discount:     s.Discount,
			LineTotal:    s.LineTotal(),
		})
	}

	return resp
}

func toStatusResponse(s registrationdomain.RegistrationStatusValue) StatusResponse {
	info, ok := registrationdomain.GetRegistrationStatusInfo(s)
	if !ok {
		return StatusResponse{Slug: string(s), Name: string(s)}
	}
	return StatusResponse{
		Slug:        info.Slug,
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Color:       info.Color,
		Icon:        info.Icon,
	}
}

func toListResponse(regs []*registrationdomain.EventRegistration, total int, f service.ListFilterInput) ListResponse {
	data := make([]RegistrationResponse, 0, len(regs))
	for _, r := range regs {
		data = append(data, toRegistrationResponse(r))
	}
	return ListResponse{
		Data:     data,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}
}

func toWaitlistResponse(w *registrationdomain.WaitlistEntry) WaitlistResponse {
	return WaitlistResponse{
		ID:        w.ID,
		EventID:   w.EventID,
		Position:  w.Position,
		CreatedAt: w.CreatedAt,
	}
}