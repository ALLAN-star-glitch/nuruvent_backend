// internal/modules/certificate/worker/render_certificate_task.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
)

type RenderCertificateHandler struct {
	svc certservice.Service
}

func NewRenderCertificateHandler(svc certservice.Service) *RenderCertificateHandler {
	return &RenderCertificateHandler{svc: svc}
}

func (h *RenderCertificateHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var job certdomain.CertificateRenderJob
	if err := json.Unmarshal(task.Payload(), &job); err != nil {
		return fmt.Errorf("unmarshal render job: %w", err)
	}

	log.Printf("[worker] rendering certificate %s", job.CertificateID)

	if err := h.svc.RenderCertificatePDF(ctx, job.CertificateID); err != nil {
		// Returning an error triggers Asynq's retry with backoff
		return fmt.Errorf("render certificate %s: %w", job.CertificateID, err)
	}

	log.Printf("[worker] certificate %s rendered", job.CertificateID)
	return nil
}