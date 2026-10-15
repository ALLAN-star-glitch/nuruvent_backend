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
	PDF         []byte
	FileName    string
}