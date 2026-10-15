// internal/modules/certificate/domain/notification_sender.go

package domain

import "context"

// NotificationSender delivers a rendered certificate to a recipient.
type NotificationSender interface {
	// SendCertificate emails the PDF to the recipient. Implementations
	// must be retriable — a failed call may be retried.
	SendCertificate(ctx context.Context, in SendCertificateInput) error
}

// SendCertificateInput carries everything the notification layer needs.
type SendCertificateInput struct {
	Certificate *Certificate

	// DownloadURL is the signed download link embedded in the email.
	// Works without authentication — necessary for guest recipients.
	DownloadURL string

	// PDF is optional. If set, the notification adapter may attach it.
	// Currently unused — links are preferred for delivery size.
	PDF      []byte
	FileName string
}