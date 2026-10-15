// internal/modules/certificate/domain/delivery.go

package domain

import "time"

// DeliveryChannel is the medium used to deliver a certificate.
type DeliveryChannel string

const (
	DeliveryChannelEmail    DeliveryChannel = "email"
	DeliveryChannelDownload DeliveryChannel = "download"
)

// IsValid reports whether the channel is known.
func (c DeliveryChannel) IsValid() bool {
	switch c {
	case DeliveryChannelEmail, DeliveryChannelDownload:
		return true
	default:
		return false
	}
}

// DeliveryStatus is the outcome of a delivery attempt.
type DeliveryStatus string

const (
	DeliveryStatusPending DeliveryStatus = "pending"
	DeliveryStatusSuccess DeliveryStatus = "success"
	DeliveryStatusFailed  DeliveryStatus = "failed"
)

// IsValid reports whether the status is known.
func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending, DeliveryStatusSuccess, DeliveryStatusFailed:
		return true
	default:
		return false
	}
}

// Delivery records an attempt to deliver a certificate.
type Delivery struct {
	ID            string
	CertificateID string
	Channel       DeliveryChannel
	Recipient     string // email or user id
	Status        DeliveryStatus
	Error         string
	Attempts      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}