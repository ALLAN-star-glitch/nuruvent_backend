// internal/modules/events/service/issue_certificates.go

package service

import (
	"context"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// Issuance modes.
const (
	IssueModeAdd     = "add"
	IssueModeReplace = "replace"
)

// IssueCertificatesForEvent initiates a certificate batch for an
// event's attendees.
//
// Recipient resolution:
//
//   - mode "add" (default) — attendance-derived + caller-supplied,
//     deduplicated by user ID or email
//   - mode "replace"       — caller-supplied only
//
// After resolving recipients, the events module delegates to the
// certificate module, which decides whether the batch is free
// (issue now) or paid (create a pending payment order).
//
// Hosts are excluded from the attendance-derived list.
func (s *eventService) IssueCertificatesForEvent(
	ctx context.Context,
	userID string,
	cmd IssueCertificatesForEventCommand,
) (*domain.InitiateBatchIssuanceResult, error) {
	event, err := s.repo.GetEventByID(ctx, cmd.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	accountID, err := s.repo.AccountIDForTeam(ctx, event.TeamID)
	if err != nil {
		return nil, err
	}

	recipients, err := s.resolveCertificateRecipients(ctx, cmd, event)
	if err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		return &domain.InitiateBatchIssuanceResult{
			Type:          "free",
			AttendeeCount: 0,
		}, nil
	}

	return s.certIssuer.InitiateBatchIssuance(ctx, domain.InitiateBatchIssuanceRequest{
		UserID:        userID,
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

// ============================================================
// RECIPIENT RESOLUTION
// ============================================================

func (s *eventService) resolveCertificateRecipients(
	ctx context.Context,
	cmd IssueCertificatesForEventCommand,
	event *domain.Event,
) ([]domain.CertificateRecipient, error) {
	mode := cmd.Mode
	if mode == "" {
		mode = IssueModeAdd
	}

	manual := toDomainRecipients(cmd.Recipients)

	var recipients []domain.CertificateRecipient

	switch mode {
	case IssueModeReplace:
		recipients = manual

	case IssueModeAdd:
		attended, err := s.attendanceReader.GetAttendedRecipients(ctx, event.ID)
		if err != nil {
			return nil, err
		}
		recipients = mergeRecipients(attended, manual)

	default:
		return nil, domain.ErrInvalidIssueMode
	}

	return fillCompletedAt(recipients, event.StartDate), nil
}

func toDomainRecipients(
	in []IssueCertificateRecipientInput,
) []domain.CertificateRecipient {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.CertificateRecipient, 0, len(in))
	for _, r := range in {
		out = append(out, domain.CertificateRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}
	return out
}

func mergeRecipients(
	a, b []domain.CertificateRecipient,
) []domain.CertificateRecipient {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]domain.CertificateRecipient, 0, len(a)+len(b))

	key := func(r domain.CertificateRecipient) string {
		if r.UserID != "" {
			return "user:" + r.UserID
		}
		return "email:" + strings.ToLower(strings.TrimSpace(r.Email))
	}

	for _, r := range a {
		k := key(r)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	for _, r := range b {
		k := key(r)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	return out
}

func fillCompletedAt(
	recipients []domain.CertificateRecipient,
	eventStart time.Time,
) []domain.CertificateRecipient {
	out := make([]domain.CertificateRecipient, len(recipients))
	for i, r := range recipients {
		if r.CompletedAt.IsZero() {
			r.CompletedAt = eventStart
		}
		out[i] = r
	}
	return out
}

// ============================================================
// COMMAND / INPUT TYPES
// ============================================================

// IssueCertificatesForEventCommand carries caller-supplied input.
type IssueCertificatesForEventCommand struct {
	EventID       string
	TemplateID    string
	EventDuration string
	IssuerName    string

	// Mode controls recipient resolution.
	//   "add" (default) — attendance-derived + caller-supplied
	//   "replace"       — caller-supplied only
	Mode string

	// Recipients is optional in "add" mode, required in "replace".
	Recipients []IssueCertificateRecipientInput
}

// IssueCertificateRecipientInput is one recipient from the caller.
type IssueCertificateRecipientInput struct {
	UserID      string
	Name        string
	Email       string
	CompletedAt time.Time
}