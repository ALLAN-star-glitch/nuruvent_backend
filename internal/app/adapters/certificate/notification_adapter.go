// internal/app/adapters/certificate/notification_adapter.go

package certificate

import (
	"context"
	"fmt"
	"time"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationAdapter bridges the notification module's service to
// the certificate module's NotificationSender port.
type NotificationAdapter struct {
	notifSvc     notificationdomain.NotificationService
	verifyBase   string
	dashboardURL string
	tokenTTL     time.Duration
}

// NewNotificationAdapter constructs a NotificationAdapter.
func NewNotificationAdapter(
	notifSvc notificationdomain.NotificationService,
	verifyBase, dashboardURL string,
	tokenTTL time.Duration,
) *NotificationAdapter {
	return &NotificationAdapter{
		notifSvc:     notifSvc,
		verifyBase:   verifyBase,
		dashboardURL: dashboardURL,
		tokenTTL:     tokenTTL,
	}
}

// SendCertificate implements certdomain.NotificationSender.
func (a *NotificationAdapter) SendCertificate(
	ctx context.Context,
	in certdomain.SendCertificateInput,
) error {
	if in.Certificate == nil {
		return fmt.Errorf("certificate is nil")
	}

	cert := in.Certificate

	return a.notifSvc.SendCertificateReady(ctx, notificationdomain.SendCertificateReadyRequest{
		To:           cert.RecipientEmail,
		Name:         cert.RecipientName,
		EventTitle:   cert.Metadata.Get(certdomain.MetadataKeyContextTitle),
		IssuerName:   cert.Metadata.Get(certdomain.MetadataKeyIssuerName),
		DownloadURL:  in.DownloadURL,
		VerifyURL:    fmt.Sprintf("%s/verify/%s", a.verifyBase, cert.VerificationCode.String()),
		DashboardURL: a.dashboardURL,
		ExpiresIn:    humanizeDuration(a.tokenTTL),
	})
}

func humanizeDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// compile-time assertion
var _ certdomain.NotificationSender = (*NotificationAdapter)(nil)