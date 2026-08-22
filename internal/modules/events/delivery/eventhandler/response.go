// internal/modules/events/delivery/eventhandler/response.go

package eventhandler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// NewEventResponseFromEvent converts a domain Event to EventResponse
func NewEventResponseFromEvent(event *domain.Event) EventResponse {
	resp := EventResponse{
		ID:               event.ID,
		Slug:             event.Slug,
		Name:             event.Name,
		DisplayName:      event.DisplayName,
		Description:      event.Description,
		EventTypeID:      event.EventTypeID,
		EventStatusID:    event.EventStatusID,
		AccountID:        event.AccountID,
		ImageURL:         event.ImageURL,
		ThumbnailURL:     event.ThumbnailURL,
		Date:             event.Date.Format("2006-01-02"),
		Time:             event.Time,
		Duration:         event.Duration,
		Price:            event.Price,
		CertificatePrice: event.CertificatePrice,
		Location:         event.Location,
		IsVirtual:        event.IsVirtual,
		IsFeatured:       event.IsFeatured,
		IsPrivate:        event.IsPrivate,
		ZoomLink:         event.ZoomLink,
		MeetLink:         event.MeetLink,
		MaxAttendees:     event.MaxAttendees,
		CurrentAttendees: event.CurrentAttendees,
		IsActive:         event.IsActive,
		CreatedAt:        event.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        event.UpdatedAt.Format(time.RFC3339),
	}

	if event.DeletedAt != nil {
		deletedAt := event.DeletedAt.Format(time.RFC3339)
		resp.DeletedAt = &deletedAt
	}

	if event.RestoredAt != nil {
		restoredAt := event.RestoredAt.Format(time.RFC3339)
		resp.RestoredAt = &restoredAt
	}

	if event.Creator != nil {
		resp.Creator = CreatorDTO{
			ID:              event.Creator.ID,
			Name:            event.Creator.Name,
			DisplayName:     event.Creator.DisplayName,
			Email:           event.Creator.Email,
			Phone:           event.Creator.Phone,
			AccountType:     event.Creator.AccountType,
			InstitutionName: event.Creator.InstitutionName,
		}
	}

	return resp
}