// internal/app/adapters/registration/notifier.go

package registration

import (
	"context"
	"fmt"
	"log"

	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// Notifier implements registrationDomain.Notifier by delegating to the
// notification module's service.
//
// RegistrationConfirmed sends a real email. The other three methods
// are logged no-ops for now.
type Notifier struct {
	notifSvc   notificationDomain.NotificationService
	userEmails registrationDomain.UserInfoProvider
}

func NewNotifier(
	notifSvc notificationDomain.NotificationService,
	userEmails registrationDomain.UserInfoProvider,
) registrationDomain.Notifier {
	return &Notifier{
		notifSvc:   notifSvc,
		userEmails: userEmails,
	}
}

func (n *Notifier) RegistrationCreated(
	ctx context.Context,
	r *registrationDomain.Registration,
) error {
	log.Printf("[RegistrationNotifier] RegistrationCreated (no-op) reg=%s", r.ID)
	return nil
}

func (n *Notifier) RegistrationConfirmed(
	ctx context.Context,
	p registrationDomain.RegistrationConfirmedPayload,
) error {
	if p.Registration == nil {
		return nil
	}

	recipient, name := n.resolveRecipient(ctx, p.Registration)
	if recipient == "" {
		log.Printf("[RegistrationNotifier] no recipient for registration %s", p.Registration.ID)
		return nil
	}

	links := make([]notificationDomain.RegistrationJoinLink, 0, len(p.JoinLinks))
	for _, l := range p.JoinLinks {
		links = append(links, notificationDomain.RegistrationJoinLink{
			SessionTitle: l.SessionTitle,
			URL:          l.URL,
		})
	}

	req := notificationDomain.SendRegistrationConfirmedRequest{
		To:                 recipient,
		Name:               name,
		RegistrationNumber: p.Registration.RegistrationNumber,
		EventName:          p.EventName,
		EventID:            p.EventID,
		JoinLinks:          links,
	}

	if err := n.notifSvc.SendRegistrationConfirmed(ctx, req); err != nil {
		return fmt.Errorf("send registration confirmed: %w", err)
	}
	return nil
}

func (n *Notifier) RegistrationCancelled(
	ctx context.Context,
	r *registrationDomain.Registration,
) error {
	log.Printf("[RegistrationNotifier] RegistrationCancelled (no-op) reg=%s", r.ID)
	return nil
}

func (n *Notifier) WaitlistPromoted(
	ctx context.Context,
	w *registrationDomain.WaitlistEntry,
) error {
	log.Printf("[RegistrationNotifier] WaitlistPromoted (no-op)")
	return nil
}

func (n *Notifier) resolveRecipient(
	ctx context.Context,
	r *registrationDomain.Registration,
) (email, name string) {
	if r.GuestEmail != "" {
		displayName := r.GuestName
		if displayName == "" {
			displayName = r.GuestEmail
		}
		return r.GuestEmail, displayName
	}

	if r.UserID == "" || n.userEmails == nil {
		return "", ""
	}
	fullName, mail, err := n.userEmails.GetUserInfo(ctx, r.UserID)
	if err != nil {
		log.Printf("[RegistrationNotifier] user lookup failed for %s: %v", r.UserID, err)
		return "", ""
	}
	if fullName == "" {
		fullName = mail
	}
	return mail, fullName
}

var _ registrationDomain.Notifier = (*Notifier)(nil)