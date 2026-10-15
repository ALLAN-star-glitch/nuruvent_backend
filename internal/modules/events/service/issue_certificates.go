// internal/modules/events/service/issue_certificates.go

package service

import (
	"context"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// IssueCertificatesForEvent issues certificates for the supplied
// recipients on behalf of an event.
//
// The caller provides the recipient list. The events module forwards
// it to the certificate module through the CertificateIssuer port.
//
// Idempotent per recipient — the certificate module guarantees this.
func (s *eventService) IssueCertificatesForEvent(
	ctx context.Context,
	cmd IssueCertificatesForEventCommand,
) (*domain.IssueCertificatesResult, error) {
	event, err := s.repo.GetEventByID(ctx, cmd.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// The events module does not store account_id on the event row;
	// the parent account is derived from the team.
	accountID, err := s.repo.AccountIDForTeam(ctx, event.TeamID)
	if err != nil {
		return nil, err
	}

	recipients := make([]domain.CertificateRecipient, 0, len(cmd.Recipients))
	for _, r := range cmd.Recipients {
		recipients = append(recipients, domain.CertificateRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	return s.certIssuer.IssueForEvent(ctx, domain.IssueCertificatesRequest{
		EventID:       event.ID,
		AccountID:     accountID,
		TemplateID:    cmd.TemplateID,
		EventTitle:    event.DisplayName,
		EventDate:     event.StartDate,
		EventDuration: cmd.EventDuration,
		IssuerName:    cmd.IssuerName,
		Recipients:    recipients,
	})
}

// IssueCertificatesForEventCommand carries caller-supplied input.
type IssueCertificatesForEventCommand struct {
	EventID       string
	TemplateID    string
	EventDuration string
	IssuerName    string
	Recipients    []IssueCertificateRecipientInput
}

// IssueCertificateRecipientInput is one recipient from the caller.
type IssueCertificateRecipientInput struct {
	UserID      string
	Name        string
	Email       string
	CompletedAt time.Time
}