// internal/modules/registration/delivery/http/mappers.go

package http

import (
	"time"

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

func toMySessionLinksResponse(r *service.SessionLinksResult) *MySessionLinksResponse {
	groups := make([]SessionLinkGroupResponse, 0, len(r.Groups))
	for _, g := range r.Groups {
		links := make([]SessionLinkResponse, 0, len(g.Links))
		for _, l := range g.Links {
			links = append(links, SessionLinkResponse{
				SessionID:      l.SessionID,
				SessionTitle:   l.SessionTitle,
				ScheduledStart: l.ScheduledStart,
				ScheduledEnd:   l.ScheduledEnd,
				Platform:       l.Platform,
				JoinURL:        l.URL,
				ExpiresAt:      l.ExpiresAt,
			})
		}
		groups = append(groups, SessionLinkGroupResponse{
			RegistrationID: g.RegistrationID,
			EventID:        g.EventID,
			EventName:      g.EventName,
			EventDate:      g.EventDate,
			Links:          links,
		})
	}
	return &MySessionLinksResponse{Groups: groups}
}

// toCrossEventRegistrationResponse flattens an EventRegistration into
// the cross-event list row shape. Event name/date and ticket name are
// populated by the service-layer read model; if they're empty here,
// the frontend shows a placeholder.
func toCrossEventRegistrationResponse(
    r *registrationdomain.CrossEventRegistrationRow,
) CrossEventRegistrationResponse {
    statusSlug := string(r.Status)
    statusLabel := statusSlug
    if info, ok := registrationdomain.GetRegistrationStatusInfo(r.Status); ok {
        statusLabel = info.DisplayName
    }
    return CrossEventRegistrationResponse{
        ID:                 r.ID,
        RegistrationNumber: r.RegistrationNumber,
        Status:             statusSlug,
        StatusLabel:        statusLabel,
        AttendeeName:       r.AttendeeName,
        Email:              r.Email,
        Phone:              r.Phone,
        IsGuest:            r.IsGuest,
        UserID:             r.UserID,
        EventID:            r.EventID,
        EventName:          r.EventName,
        EventStartDate:     formatTimeOrEmpty(r.EventStartDate),
		EventImageURL:      r.EventImageURL, 
        TicketName:         r.TicketName,
        CreatedAt:          r.CreatedAt,
		IsVirtual:        r.IsVirtual,
		IsHybrid:         r.IsHybrid,
		VenueName:        r.VenueName,
		VenueAddress:     r.VenueAddress,
		VenueCity:        r.VenueCity,
		VenueCountry:     r.VenueCountry,
		InPersonLocation: r.InPersonLocation,
    }
}

func formatTimeOrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}