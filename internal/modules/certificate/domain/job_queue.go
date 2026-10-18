package domain

import "context"

// CertificateRenderJob is the payload enqueued for background rendering.
type CertificateRenderJob struct {
	CertificateID string
	AccountID     string
}

// RenderJobQueue is the outbound port for enqueueing background render jobs.
// The adapter wraps Asynq.
type RenderJobQueue interface {
	// EnqueueRender schedules a PDF render for the given certificate.
	// Idempotent — enqueueing the same certificate twice doesn't create duplicates.
	EnqueueRender(ctx context.Context, job CertificateRenderJob) error
}